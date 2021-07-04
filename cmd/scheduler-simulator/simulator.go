package main

import (
	"k8s.io/klog/v2"

	"k8s.io/kubernetes/cmd/scheduler-simulator/config"
	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler"
	"k8s.io/kubernetes/cmd/scheduler-simulator/server"
	"k8s.io/kubernetes/cmd/scheduler-simulator/server/di"
	"k8s.io/kubernetes/cmd/scheduler-simulator/shutdownfn"
	"k8s.io/kubernetes/test/integration/framework"
)

// entry point.
func main() {
	// start etcd and then start simulator and needed k8s components.
	framework.EtcdMain(startSimulator)
}

// startSimulator starts simulator and needed k8s components.
// It will return exit code.
func startSimulator() int {
	cfg, err := config.NewConfig()
	if err != nil {
		klog.Errorf("failed to get config: %v", err)
		return 1
	}

	// start kube-apiserver and kube-scheduler
	clientset, podInformer, shutdownFn1, err := scheduler.SetupSchedulerOrDie()
	if err != nil {
		klog.Errorf("failed to start scheduler: %v", err)
		return 1
	}

	dic := di.NewDIContainer(clientset, podInformer)

	// start simulator server
	s := server.NewSimulatorServer(cfg, dic)
	shutdownFn2, err := s.Start(cfg.Port)
	if err != nil {
		klog.Errorf("failed to start simulator server: %v", err)
		shutdownfn.WaitShutdown(shutdownFn1)
		return 1
	}

	shutdownfn.WaitShutdown(shutdownFn1, shutdownFn2)
	return 0
}
