package server

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"k8s.io/kubernetes/cmd/scheduler-simulator/config"
	"k8s.io/kubernetes/cmd/scheduler-simulator/config/env"
	"k8s.io/kubernetes/cmd/scheduler-simulator/shutdownfn"
)

type SimulatorServer struct {
	e *echo.Echo
}

func NewSimulatorServer(cfg *config.Config) *SimulatorServer {
	e := echo.New()

	e.Use(middleware.Logger())

	_ = e.Group("/api/v1")

	// register apis

	s := &SimulatorServer{e: e}
	s.setLogLevel(cfg.Env)

	return s
}

func (s *SimulatorServer) setLogLevel(e env.Env) {
	switch e {
	case env.Production:
		s.e.Logger.SetLevel(log.INFO)
	case env.Development:
		s.e.Logger.SetLevel(log.DEBUG)
	}
}

// Start starts SimulatorServer.
func (s *SimulatorServer) Start(port int) (shutdownfn.Shutdownfn, error) {
	e := s.e

	go func() {
		if err := e.Start(":" + strconv.Itoa(port)); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatalf("failed to start server successfully: %v", err)
		}
	}()

	shutdownFn := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := e.Shutdown(ctx); err != nil {
			e.Logger.Warnf("failed to shutdown simulator server successfully: %v", err)
		}
	}

	return shutdownFn, nil
}
