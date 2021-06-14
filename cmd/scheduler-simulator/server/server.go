package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"golang.org/x/xerrors"

	"k8s.io/kubernetes/cmd/scheduler-simulator/config"
	"k8s.io/kubernetes/cmd/scheduler-simulator/config/env"
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
func (s *SimulatorServer) Start(port int) {
	e := s.e

	go func() {
		if err := e.Start(":" + strconv.Itoa(port)); err != nil && !xerrors.Is(err, http.ErrServerClosed) {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
