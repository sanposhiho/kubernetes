package server

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"golang.org/x/xerrors"

	"k8s.io/kubernetes/cmd/scheduler-simulator/config"
	"k8s.io/kubernetes/cmd/scheduler-simulator/config/env"
	"k8s.io/kubernetes/cmd/scheduler-simulator/server/di"
	"k8s.io/kubernetes/cmd/scheduler-simulator/server/handler"
	"k8s.io/kubernetes/cmd/scheduler-simulator/shutdownfn"
)

type SimulatorServer struct {
	e *echo.Echo
}

func NewSimulatorServer(cfg *config.Config, dic *di.Container) *SimulatorServer {
	e := echo.New()

	e.Use(middleware.Logger())

	// initialize each handler
	nodeHandler := handler.NewNodeHandler(dic.NodeService())
	podHandler := handler.NewPodHandler(dic.PodService())

	// register apis
	v1 := e.Group("/api/v1")
	// FIXME: create nodes with POST
	v1.GET("/nodes/create", nodeHandler.CreateNode)
	v1.GET("/nodes", nodeHandler.ListNode)
	v1.GET("/nodes/:name", nodeHandler.GetNode)
	// FIXME: delete nodes with DELETE
	v1.GET("/nodes/delete/:name", nodeHandler.DeleteNode)

	// FIXME: create pods with POST
	v1.GET("/pods/create", podHandler.CreatePod)
	v1.GET("/pods", podHandler.ListPod)
	v1.GET("/pods/:name", podHandler.GetPod)
	// FIXME: delete pods with DELETE
	v1.GET("/pods/delete/:name", podHandler.DeletePod)

	// initialize SimulatorServer.
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
		if err := e.Start(":" + strconv.Itoa(port)); err != nil && !xerrors.Is(err, http.ErrServerClosed) {
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
