package main

import (
	"os"
	"os/signal"
	"syscall"

	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/util"

	"golang.org/x/xerrors"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"k8s.io/kubernetes/cmd/scheduler-simulator/config"
	"k8s.io/kubernetes/cmd/scheduler-simulator/k8sapiserver"
	"k8s.io/kubernetes/cmd/scheduler-simulator/pvcontroller"
	"k8s.io/kubernetes/cmd/scheduler-simulator/server"
	"k8s.io/kubernetes/cmd/scheduler-simulator/server/di"
)

// entry point.
func main() {
	if err := startSimulator(); err != nil {
		klog.Fatalf("failed with error on running simulator: %+v", err)
	}
}

// startSimulator starts simulator and needed k8s components.
func startSimulator() error {
	cfg, err := config.NewConfig()
	if err != nil {
		return xerrors.Errorf("get config: %w", err)
	}

	restclientCfg, apiShutdown := k8sapiserver.StartAPIServerOrDie(cfg.EtcdURL)
	defer apiShutdown()

	client := clientset.NewForConfigOrDie(restclientCfg)

	pvshutdown, err := pvcontroller.StartPersistentVolumeController(client)
	if err != nil {
		return xerrors.Errorf("start pv controller: %w", err)
	}
	defer pvshutdown()

	dic := di.NewDIContainer(client, restclientCfg)

	sc, err := util.DefaultSchedulerConfig()
	if err != nil {
		return xerrors.Errorf("create scheduler config")
	}

	if err := dic.SchedulerService().StartScheduler(sc); err != nil {
		return xerrors.Errorf("start scheduler: %w", err)
	}
	defer dic.SchedulerService().ShutdownScheduler()

	// start simulator server
	s := server.NewSimulatorServer(cfg, dic)
	shutdownFn3, err := s.Start(cfg.Port)
	if err != nil {
		return xerrors.Errorf("start simulator server: %w", err)
	}
	defer shutdownFn3()

	// wait the signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	<-quit

	return nil
}
