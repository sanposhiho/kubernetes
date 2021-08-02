package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/klog/v2"
	config2 "k8s.io/kube-scheduler/config/v1beta1"

	"k8s.io/kubernetes/cmd/scheduler-simulator/server/di"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	"k8s.io/kubernetes/pkg/scheduler/apis/config/v1beta1"
)

// SchedulerConfigHandler is handler for manage scheduler config.
type SchedulerConfigHandler struct {
	service di.SchedulerService
}

func NewSchedulerConfigHandler(s di.SchedulerService) *SchedulerConfigHandler {
	return &SchedulerConfigHandler{
		service: s,
	}
}

func (h *SchedulerConfigHandler) GetSchedulerConfig(c echo.Context) error {
	cfg := h.service.GetSchedulerConfig()
	cfg2, err := convertToOmitEmpty(cfg)
	if err != nil {
		klog.Errorf("failed to convert scheduler configuration to omit empty: %+v", err)
		c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, cfg2)
}

func (h *SchedulerConfigHandler) ApplySchedulerConfig(c echo.Context) error {
	reqSchedulerCfg := new(config2.KubeSchedulerConfiguration)
	if err := c.Bind(reqSchedulerCfg); err != nil {
		klog.Errorf("failed to bind scheduler config request: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	converted, err := convertFromOmitEmpty(reqSchedulerCfg)
	if err != nil {
		klog.Errorf("failed to convert scheduler configuration from omit empty: %+v", err)
		c.JSON(http.StatusInternalServerError, err)
	}

	if err := h.service.RestartScheduler(converted); err != nil {
		klog.Errorf("failed to restart scheduler: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

// convertToOmitEmpty converts config.KubeSchedulerConfiguration to config2.KubeSchedulerConfiguration
// This is needed to omit empty.
// (config.KubeSchedulerConfiguration doesn't have omitempty json tags on fields.)
func convertToOmitEmpty(cfg *config.KubeSchedulerConfiguration) (*config2.KubeSchedulerConfiguration, error) {
	cfg2 := &config2.KubeSchedulerConfiguration{}
	converter := conversion.NewConverter(conversion.DefaultNameFunc)
	if err := converter.RegisterGeneratedUntypedConversionFunc(cfg, cfg2, func(a, b interface{}, scope conversion.Scope) error {
		typeda := a.(*config.KubeSchedulerConfiguration)
		typedb := b.(*config2.KubeSchedulerConfiguration)
		return v1beta1.Convert_config_KubeSchedulerConfiguration_To_v1beta1_KubeSchedulerConfiguration(typeda, typedb, scope)
	}); err != nil {
		return nil, xerrors.Errorf("register generated untyped conversion: %w", err)
	}

	err := converter.Convert(cfg, cfg2, nil)
	if err != nil {
		return nil, xerrors.Errorf("convert scheduler configuration type: %w", err)
	}
	return cfg2, nil
}

// convertFromOmitEmpty converts config.KubeSchedulerConfiguration to config2.KubeSchedulerConfiguration
// This is needed to omit empty.
// (config.KubeSchedulerConfiguration doesn't have omitempty json tags on fields.)
func convertFromOmitEmpty(cfg *config2.KubeSchedulerConfiguration) (*config.KubeSchedulerConfiguration, error) {
	cfg2 := &config.KubeSchedulerConfiguration{}
	converter := conversion.NewConverter(conversion.DefaultNameFunc)
	if err := converter.RegisterGeneratedUntypedConversionFunc(cfg, cfg2, func(a, b interface{}, scope conversion.Scope) error {
		typeda := a.(*config2.KubeSchedulerConfiguration)
		typedb := b.(*config.KubeSchedulerConfiguration)
		return v1beta1.Convert_v1beta1_KubeSchedulerConfiguration_To_config_KubeSchedulerConfiguration(typeda, typedb, scope)
	}); err != nil {
		return nil, xerrors.Errorf("register generated untyped conversion: %w", err)
	}

	err := converter.Convert(cfg, cfg2, nil)
	if err != nil {
		return nil, xerrors.Errorf("convert scheduler configuration type: %w", err)
	}
	return cfg2, nil
}
