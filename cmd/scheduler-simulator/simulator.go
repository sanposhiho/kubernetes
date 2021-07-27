package main

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/xerrors"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/cmd/scheduler-simulator/etcd"

	"k8s.io/kubernetes/cmd/scheduler-simulator/config"
	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler"
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

	// start kube-apiserver and kube-scheduler
	clientset, shutdownFn1, err := scheduler.SetupSchedulerOrDie(cfg)
	if err != nil {
		return xerrors.Errorf("start scheduler and some needed k8s components: %w", err)
	}
	defer shutdownFn1()

	etcdclient := etcd.NewClient(cfg)

	dic := di.NewDIContainer(clientset, etcdclient)

	// start simulator server
	s := server.NewSimulatorServer(cfg, dic)
	shutdownFn2, err := s.Start(cfg.Port)
	if err != nil {
		return xerrors.Errorf("start simulator server: %w", err)
	}
	defer shutdownFn2()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	<-quit

	return nil
}
