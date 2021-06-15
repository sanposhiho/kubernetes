package scheduler

import (
	"context"

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

func SetupScheduler() (clientset.Interface, coreinformers.PodInformer, shutdownfn.Shutdownfn, error) {
	// Note: This function die when a error happen.
	apiURL, apiShutdown := util.StartApiserver()

	cfg := &restclient.Config{
		Host:          apiURL,
		ContentConfig: restclient.ContentConfig{GroupVersion: &schema.GroupVersion{Group: "", Version: "v1"}},
		QPS:           5000.0,
		Burst:         5000,
	}

	client := clientset.NewForConfigOrDie(cfg)

	schedCfg, err := defaultComponentConfig()
	if err != nil {
		return nil, nil, nil, xerrors.Errorf("get default component config: %w", err)
	}

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

// defaultComponentConfig create KubeSchedulerConfiguration default configuration.
func defaultComponentConfig() (*config.KubeSchedulerConfiguration, error) {
	gvk := v1beta1.SchemeGroupVersion.WithKind("KubeSchedulerConfiguration")
	cfg := config.KubeSchedulerConfiguration{}
	_, _, err := kubeschedulerscheme.Codecs.UniversalDecoder().Decode(nil, &gvk, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

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
