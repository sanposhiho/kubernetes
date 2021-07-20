package scheduler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/google/uuid"
	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/server/options"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/events"
	"k8s.io/klog/v2"
	"k8s.io/kube-scheduler/config/v1beta1"

	simulatorcfg "k8s.io/kubernetes/cmd/scheduler-simulator/config"
	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/plugins"
	"k8s.io/kubernetes/cmd/scheduler-simulator/shutdownfn"
	"k8s.io/kubernetes/pkg/controller/volume/persistentvolume"
	"k8s.io/kubernetes/pkg/scheduler"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	kubeschedulerscheme "k8s.io/kubernetes/pkg/scheduler/apis/config/scheme"
	"k8s.io/kubernetes/pkg/scheduler/profile"
	"k8s.io/kubernetes/pkg/volume"
	"k8s.io/kubernetes/pkg/volume/hostpath"
	"k8s.io/kubernetes/pkg/volume/local"
	"k8s.io/kubernetes/test/integration/framework"
)

// SetupSchedulerOrDie starts k8s-apiserver and scheduler.
func SetupSchedulerOrDie(simulatorcfg *simulatorcfg.Config) (clientset.Interface, coreinformers.PodInformer, shutdownfn.Shutdownfn, error) {
	apiURL, apiShutdown := startAPIServerOrDie(simulatorcfg.EtcdURL)

	cfg := &restclient.Config{
		Host:          apiURL,
		ContentConfig: restclient.ContentConfig{GroupVersion: &schema.GroupVersion{Group: "", Version: "v1"}},
		QPS:           5000.0,
		Burst:         5000,
	}

	schedCfg, err := schedulerConfig()
	if err != nil {
		return nil, nil, nil, xerrors.Errorf("get default component config: %w", err)
	}

	client := clientset.NewForConfigOrDie(cfg)

	ctx, cancel := context.WithCancel(context.Background())

	podInformer, err := startScheduler(ctx, client, cfg, schedCfg)
	if err != nil {
		cancel()
		return nil, nil, nil, xerrors.Errorf("start scheduler: %w", err)
	}

	if err := startPersistentVolumeController(ctx, client); err != nil {
		cancel()
		return nil, nil, nil, xerrors.Errorf("start pv controller: %w", err)
	}

	shutdownFunc := func() {
		cancel()
		apiShutdown()
	}

	return client, podInformer, shutdownFunc, nil
}

// startAPIServerOrDie starts API server, and it make panic when a error happen.
// TODO: change it not to use integration framework
func startAPIServerOrDie(etcdURL string) (string, shutdownfn.Shutdownfn) {
	h := &framework.APIServerHolder{Initialized: make(chan struct{})}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		<-h.Initialized
		h.M.GenericAPIServer.Handler.ServeHTTP(w, req)
	}))

	// etcdOption for control plane
	etcdOptions := options.NewEtcdOptions(storagebackend.NewDefaultConfig(uuid.New().String(), nil))
	etcdOptions.StorageConfig.Transport.ServerList = []string{etcdURL}
	c := framework.NewIntegrationTestControlPlaneConfigWithOptions(&framework.MasterConfigOptions{
		EtcdOptions: etcdOptions,
	})
	c.GenericConfig.OpenAPIConfig = framework.DefaultOpenAPIConfig()

	// Note: This function die when a error happen.
	_, _, closeFn := framework.RunAnAPIServerUsingServer(c, s, h)

	shutdownFunc := func() {
		klog.Infof("destroying API server")
		closeFn()
		s.Close()
		klog.Infof("destroyed API server")
	}
	return s.URL, shutdownFunc
}

// schedulerConfig creates KubeSchedulerConfiguration default configuration.
func schedulerConfig() (*config.KubeSchedulerConfiguration, error) {
	gvk := v1beta1.SchemeGroupVersion.WithKind("KubeSchedulerConfiguration")
	cfg := config.KubeSchedulerConfiguration{}
	_, _, err := kubeschedulerscheme.Codecs.UniversalDecoder().Decode(nil, &gvk, &cfg)
	if err != nil {
		return nil, xerrors.Errorf("decode config: %w", err)
	}

	pc, err := plugins.NewPluginConfig()
	if err != nil {
		return nil, xerrors.Errorf("get plugin configs: %w", err)
	}

	cfg.Profiles = []config.KubeSchedulerProfile{
		{
			SchedulerName: v1.DefaultSchedulerName,
			Plugins:       plugins.NewPlugin(),
			PluginConfig:  pc,
		},
	}
	return &cfg, nil
}

// startScheduler starts scheduler.
func startScheduler(
	ctx context.Context,
	clientSet clientset.Interface,
	kubeConfig *restclient.Config,
	cfg *config.KubeSchedulerConfiguration,
) (coreinformers.PodInformer, error) {
	informerFactory := scheduler.NewInformerFactory(clientSet, 0)
	evtBroadcaster := events.NewBroadcaster(&events.EventSinkImpl{
		Interface: clientSet.EventsV1(),
	})

	evtBroadcaster.StartRecordingToSink(ctx.Done())

	sched, err := scheduler.New(
		clientSet,
		informerFactory,
		profile.NewRecorderFactory(evtBroadcaster),
		ctx.Done(),
		scheduler.WithKubeConfig(kubeConfig),
		scheduler.WithProfiles(cfg.Profiles...),
		scheduler.WithAlgorithmSource(cfg.AlgorithmSource),
		scheduler.WithPercentageOfNodesToScore(cfg.PercentageOfNodesToScore),
		scheduler.WithPodMaxBackoffSeconds(cfg.PodMaxBackoffSeconds),
		scheduler.WithPodInitialBackoffSeconds(cfg.PodInitialBackoffSeconds),
		scheduler.WithExtenders(cfg.Extenders...),
		scheduler.WithParallelism(cfg.Parallelism),
		scheduler.WithFrameworkOutOfTreeRegistry(plugins.NewRegistry(informerFactory, clientSet)),
	)
	if err != nil {
		return nil, xerrors.Errorf("create scheduler: %w", err)
	}

	informerFactory.Start(ctx.Done())
	informerFactory.WaitForCacheSync(ctx.Done())

	go sched.Run(ctx)

	return informerFactory.Core().V1().Pods(), nil
}

func startPersistentVolumeController(ctx context.Context, client clientset.Interface) error {
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	params := persistentvolume.ControllerParameters{
		KubeClient:                client,
		SyncPeriod:                1 * time.Second,
		VolumePlugins:             append(local.ProbeVolumePlugins(), hostpath.ProbeVolumePlugins(volume.VolumeConfig{})...),
		VolumeInformer:            informerFactory.Core().V1().PersistentVolumes(),
		ClaimInformer:             informerFactory.Core().V1().PersistentVolumeClaims(),
		ClassInformer:             informerFactory.Storage().V1().StorageClasses(),
		PodInformer:               informerFactory.Core().V1().Pods(),
		NodeInformer:              informerFactory.Core().V1().Nodes(),
		EnableDynamicProvisioning: true,
	}
	volumeController, err := persistentvolume.NewController(params)
	if err != nil {
		return fmt.Errorf("failed to construct persistentvolume controller: %w", err)
	}

	go volumeController.Run(ctx.Done())
	informerFactory.Start(ctx.Done())
	informerFactory.WaitForCacheSync(ctx.Done())

	return nil
}
