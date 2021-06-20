package scheduler

import (
	"context"
	"net/http"
	"net/http/httptest"

	"k8s.io/kubernetes/test/integration/framework"

	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/events"
	"k8s.io/klog/v2"
	"k8s.io/kube-scheduler/config/v1beta1"

	"k8s.io/kubernetes/cmd/scheduler-simulator/shutdownfn"
	"k8s.io/kubernetes/pkg/scheduler"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	kubeschedulerscheme "k8s.io/kubernetes/pkg/scheduler/apis/config/scheme"
	"k8s.io/kubernetes/pkg/scheduler/profile"
	"k8s.io/kubernetes/test/integration/util"
)

// SetupSchedulerOrDie starts k8s-apiserver and scheduler.
func SetupSchedulerOrDie() (clientset.Interface, coreinformers.PodInformer, shutdownfn.Shutdownfn, error) {
	apiURL, apiShutdown := startAPIServerOrDie()

	cfg := &restclient.Config{
		Host:          apiURL,
		ContentConfig: restclient.ContentConfig{GroupVersion: &schema.GroupVersion{Group: "", Version: "v1"}},
		QPS:           5000.0,
		Burst:         5000,
	}

	schedCfg, err := defaultComponentConfig()
	if err != nil {
		return nil, nil, nil, xerrors.Errorf("get default component config: %w", err)
	}

	client := clientset.NewForConfigOrDie(cfg)

	podInformer, schedulerShutdown, err := startScheduler(client, cfg, schedCfg)
	if err != nil {
		return nil, nil, nil, xerrors.Errorf("start scheduler: %w", err)
	}
	fakePVControllerShutdown := util.StartFakePVController(client)

	shutdownFunc := func() {
		fakePVControllerShutdown()
		schedulerShutdown()
		apiShutdown()
	}

	return client, podInformer, shutdownFunc, nil
}

// startAPIServerOrDie starts API server, and it make panic when a error happen.
func startAPIServerOrDie() (string, shutdownfn.Shutdownfn) {
	h := &framework.APIServerHolder{Initialized: make(chan struct{})}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		<-h.Initialized
		h.M.GenericAPIServer.Handler.ServeHTTP(w, req)
	}))

	c := framework.NewIntegrationTestControlPlaneConfig()
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

// defaultComponentConfig creates KubeSchedulerConfiguration default configuration.
func defaultComponentConfig() (*config.KubeSchedulerConfiguration, error) {
	gvk := v1beta1.SchemeGroupVersion.WithKind("KubeSchedulerConfiguration")
	cfg := config.KubeSchedulerConfiguration{}
	_, _, err := kubeschedulerscheme.Codecs.UniversalDecoder().Decode(nil, &gvk, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// startScheduler starts scheduler.
func startScheduler(
	clientSet clientset.Interface,
	kubeConfig *restclient.Config,
	cfg *config.KubeSchedulerConfiguration,
) (coreinformers.PodInformer, shutdownfn.Shutdownfn, error) {
	ctx, cancel := context.WithCancel(context.Background())

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
	)
	if err != nil {
		cancel()
		return nil, nil, xerrors.Errorf("create scheduler: %w", err)
	}

	informerFactory.Start(ctx.Done())
	informerFactory.WaitForCacheSync(ctx.Done())

	go sched.Run(ctx)

	shutdownFunc := func() {
		klog.Infof("destroying scheduler")
		cancel()
		klog.Infof("destroyed scheduler")
	}
	return informerFactory.Core().V1().Pods(), shutdownFunc, nil
}
