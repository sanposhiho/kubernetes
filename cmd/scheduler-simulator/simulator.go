package main

import (
	"log"

	"k8s.io/kubernetes/cmd/scheduler-simulator/config"
	"k8s.io/kubernetes/cmd/scheduler-simulator/server"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("failed to get config: %v", err)
	}

	s := server.NewSimulatorServer(cfg)
	s.Start(cfg.Port)
}
