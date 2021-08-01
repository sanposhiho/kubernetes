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
)

// SimulatorServer is server for simulator.
type SimulatorServer struct {
	e *echo.Echo
}

// NewSimulatorServer initialize SimulatorServer.
func NewSimulatorServer(cfg *config.Config, dic *di.Container) *SimulatorServer {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowCredentials: true,
	}))

	// initialize each handler
	nodeHandler := handler.NewNodeHandler(dic.NodeService())
	podHandler := handler.NewPodHandler(dic.PodService())
	namespaceHandler := handler.NewNamespaceHandler(dic.NamespaceService())
	pvHandler := handler.NewPersistentVolumeHandler(dic.PersistentVolumeService())
	pvcHandler := handler.NewPersistentVolumeClaimHandler(dic.PersistentVolumeClaimService())
	storageClassHandler := handler.NewStorageClassHandler(dic.StorageClassService())
	schedulercfgHandler := handler.NewSchedulerConfigHandler(dic.SchedulerConfigService())

	// register apis
	v1 := e.Group("/api/v1")
	v1.POST("/namespaces", namespaceHandler.CreateNamespace)

	v1simulator := v1.Group("/simulators/:simulatorID")

	v1simulator.GET("/schedulerconfiguration", schedulercfgHandler.GetSchedulerConfig)
	v1simulator.POST("/schedulerconfiguration", schedulercfgHandler.ApplySchedulerConfig)

	v1simulator.GET("/nodes", nodeHandler.ListNode)
	v1simulator.POST("/nodes", nodeHandler.ApplyNode)
	v1simulator.GET("/nodes/:name", nodeHandler.GetNode)
	v1simulator.DELETE("/nodes/:name", nodeHandler.DeleteNode)

	v1simulator.GET("/pods", podHandler.ListPod)
	v1simulator.POST("/pods", podHandler.ApplyPod)
	v1simulator.GET("/pods/:name", podHandler.GetPod)
	v1simulator.DELETE("/pods/:name", podHandler.DeletePod)

	v1simulator.GET("/persistentvolumes", pvHandler.ListPersistentVolume)
	v1simulator.POST("/persistentvolumes", pvHandler.ApplyPersistentVolume)
	v1simulator.GET("/persistentvolumes/:name", pvHandler.GetPersistentVolume)
	v1simulator.DELETE("/persistentvolumes/:name", pvHandler.DeletePersistentVolume)

	v1simulator.GET("/persistentvolumeclaims", pvcHandler.ListPersistentVolumeClaim)
	v1simulator.POST("/persistentvolumeclaims", pvcHandler.ApplyPersistentVolumeClaim)
	v1simulator.GET("/persistentvolumeclaims/:name", pvcHandler.GetPersistentVolumeClaim)
	v1simulator.DELETE("/persistentvolumeclaims/:name", pvcHandler.DeletePersistentVolumeClaim)

	v1simulator.GET("/storageclasses", storageClassHandler.ListStorageClass)
	v1simulator.POST("/storageclasses", storageClassHandler.ApplyStorageClass)
	v1simulator.GET("/storageclasses/:name", storageClassHandler.GetStorageClass)
	v1simulator.DELETE("/storageclasses/:name", storageClassHandler.DeleteStorageClass)

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
func (s *SimulatorServer) Start(port int) (
	func(), // function for shutdown
	error,
) {
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
