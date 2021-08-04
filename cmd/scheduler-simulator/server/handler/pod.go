package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/klog/v2"

	"k8s.io/kubernetes/cmd/scheduler-simulator/server/di"
)

// PodHandler is handler for manage pod.
type PodHandler struct {
	service di.PodService
}

// NewPodHandler initializes PodHandler.
func NewPodHandler(s di.PodService) *PodHandler {
	return &PodHandler{service: s}
}

// ApplyPod handles the endpoint for applying pod.
func (h *PodHandler) ApplyPod(c echo.Context) error {
	ctx := c.Request().Context()

	pod := new(v1.PodApplyConfiguration)
	if err := c.Bind(pod); err != nil {
		klog.Errorf("failed to bind apply pod request: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if err := h.service.Apply(ctx, pod); err != nil {
		klog.Errorf("failed to apply pod: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

// GetPod handles the endpoint for getting pod.
func (h *PodHandler) GetPod(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.Param("name")

	p, err := h.service.Get(ctx, name)
	if err != nil {
		klog.Errorf("failed to get pod: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, p)
}

// ListPod handles the endpoint for listing pod.
func (h *PodHandler) ListPod(c echo.Context) error {
	ctx := c.Request().Context()

	ps, err := h.service.List(ctx)
	if err != nil {
		klog.Errorf("failed to list pods: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, ps)
}

// DeletePod handles the endpoint for deleting pod.
func (h *PodHandler) DeletePod(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.Param("name")

	if err := h.service.Delete(ctx, name); err != nil {
		klog.Errorf("failed to delete pod: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}
