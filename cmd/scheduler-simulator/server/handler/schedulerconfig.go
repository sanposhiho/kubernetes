package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	schedulerapi "k8s.io/kube-scheduler/config/v1beta1"
)

// SchedulerConfigHandler is handler for manage scheduler config.
type SchedulerConfigHandler struct {
	service SchedulerService
}

// SchedulerService represents service for manage scheduler.
type SchedulerService interface {
	GetSchedulerConfig() *schedulerapi.KubeSchedulerConfiguration
	RestartScheduler(cfg *schedulerapi.KubeSchedulerConfiguration) error
}

func NewSchedulerConfigHandler(s SchedulerService) *SchedulerConfigHandler {
	return &SchedulerConfigHandler{
		service: s,
	}
}

func (h *SchedulerConfigHandler) GetSchedulerConfig(c echo.Context) error {
	cfg := h.service.GetSchedulerConfig()
	return c.JSON(http.StatusOK, convertToJSON(cfg))
}

func (h *SchedulerConfigHandler) ApplySchedulerConfig(c echo.Context) error {
	reqSchedulerCfg := new(schedulerapi.KubeSchedulerConfiguration)
	if err := c.Bind(reqSchedulerCfg); err != nil {
		klog.Errorf("failed to bind scheduler config request: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if err := h.service.RestartScheduler(reqSchedulerCfg); err != nil {
		klog.Errorf("failed to restart scheduler: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

func convertToJSON(cfg *schedulerapi.KubeSchedulerConfiguration) *schedulerConfigurationJSON {
	return &schedulerConfigurationJSON{
		TypeMeta: cfg.TypeMeta,
		Profiles: cfg.Profiles,
	}
}

type schedulerConfigurationJSON struct {
	metav1.TypeMeta `json:",inline"`
	Profiles        []schedulerapi.KubeSchedulerProfile `json:"profiles,omitempty"`
}
