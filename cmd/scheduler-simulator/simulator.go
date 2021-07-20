package main

import (
	"golang.org/x/xerrors"
	"k8s.io/klog/v2"

	"k8s.io/kubernetes/cmd/scheduler-simulator/config"
	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler"
	"k8s.io/kubernetes/cmd/scheduler-simulator/server"
	"k8s.io/kubernetes/cmd/scheduler-simulator/server/di"
	"k8s.io/kubernetes/cmd/scheduler-simulator/shutdownfn"
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

	// start kube-apiserver and kube-scheduler
	clientset, podInformer, shutdownFn1, err := scheduler.SetupSchedulerOrDie()
	if err != nil {
		return xerrors.Errorf("start scheduler and some needed k8s components: %w", err)
	}

	dic := di.NewDIContainer(clientset, podInformer)

	// start simulator server
	s := server.NewSimulatorServer(cfg, dic)
	shutdownFn2, err := s.Start(cfg.Port)
	if err != nil {
		shutdownfn.WaitShutdown(shutdownFn1)
		return xerrors.Errorf("start simulator server: %w", err)
	}

	shutdownfn.WaitShutdown(shutdownFn1, shutdownFn2)
	return nil
}
