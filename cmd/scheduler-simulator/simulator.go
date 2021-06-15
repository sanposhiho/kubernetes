package main

import (
	"log"

	"k8s.io/kubernetes/test/integration/framework"

	"k8s.io/kubernetes/cmd/scheduler-simulator/config"
	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler"
	"k8s.io/kubernetes/cmd/scheduler-simulator/server"
	"k8s.io/kubernetes/cmd/scheduler-simulator/shutdownfn"
)

func main() {
	framework.EtcdMain(startSimulator)
}

func startSimulator() int {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Printf("failed to get config: %v", err)
		return 1
	}

	// start kube-apiserver and kube-scheduler
	_, shutdownFn1, err := scheduler.SetupScheduler()
	if err != nil {
		log.Printf("failed to start scheduler: %v", err)
		return 1
	}

	// start simulator server
	s := server.NewSimulatorServer(cfg)
	shutdownFn2, err := s.Start(cfg.Port)
	if err != nil {
		shutdownfn.WaitShutdown(shutdownFn1)
		log.Printf("failed to start simulator server: %v", err)
		return 1
	}

	shutdownfn.WaitShutdown(shutdownFn1, shutdownFn2)
	return 0
}
