package handler

import (
	"context"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/kubernetes/cmd/scheduler-simulator/schedulerconfig"

	"github.com/labstack/echo/v4"
	"golang.org/x/xerrors"
	"k8s.io/klog/v2"
	schedulerapi "k8s.io/kube-scheduler/config/v1beta1"

	"k8s.io/kubernetes/cmd/scheduler-simulator/errors"
)

// SchedulerConfigHandler is handler for manage scheduler config.
type SchedulerConfigHandler struct {
	service SchedulerConfigService
}

// SchedulerConfigService represents service for manage scheduler config.
type SchedulerConfigService interface {
	GetSchedulerConfig(ctx context.Context, simulatorID string) (*schedulerapi.KubeSchedulerConfiguration, error)
	PutSchedulerConfig(ctx context.Context, simulatorID string, cfg *schedulerapi.KubeSchedulerConfiguration) error
}

func NewSchedulerConfigHandler(s SchedulerConfigService) *SchedulerConfigHandler {
	return &SchedulerConfigHandler{
		service: s,
	}
}

func (h *SchedulerConfigHandler) GetSchedulerConfig(c echo.Context) error {
	ctx := c.Request().Context()

	simulatorID := c.Param("simulatorID")

	cfg, err := h.service.GetSchedulerConfig(ctx, simulatorID)
	if err != nil {
		if !xerrors.Is(err, errors.ErrNotFound) {
			klog.Errorf("failed to get scheduler config: %+v", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		cfg = schedulerconfig.DefaultSchedulerConfig()
	}

	return c.JSON(http.StatusOK, convertToJSON(cfg))
}

func (h *SchedulerConfigHandler) PutSchedulerConfig(c echo.Context) error {
	ctx := c.Request().Context()

	simulatorID := c.Param("simulatorID")

	reqSchedulerCfg := new(schedulerapi.KubeSchedulerConfiguration)
	if err := c.Bind(reqSchedulerCfg); err != nil {
		klog.Errorf("failed to bind scheduler config request: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if err := h.service.PutSchedulerConfig(ctx, simulatorID, reqSchedulerCfg); err != nil {
		klog.Errorf("failed to put scheduler config: %+v", err)
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
