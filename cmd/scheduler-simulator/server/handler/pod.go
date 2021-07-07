package handler

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/klog/v2"
)

// PodHandler is handler for manage pod.
type PodHandler struct {
	service PodService
}

// PodService represents service for manage Pods.
type PodService interface {
	Get(ctx context.Context, name string, simulatorID string) (*corev1.Pod, error)
	List(ctx context.Context, simulatorID string) (*corev1.PodList, error)
	Apply(ctx context.Context, simulatorID string, pod *v1.PodApplyConfiguration) error
	Delete(ctx context.Context, name string, simulatorID string) error
}

// NewPodHandler initializes PodHandler.
func NewPodHandler(s PodService) *PodHandler {
	return &PodHandler{service: s}
}

// ApplyPod handles the endpoint for applying pod.
func (h *PodHandler) ApplyPod(c echo.Context) error {
	ctx := c.Request().Context()

	id := c.Param("simulatorID")

	pod := new(v1.PodApplyConfiguration)
	if err := c.Bind(pod); err != nil {
		klog.Errorf("failed to bind apply pod request: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if err := h.service.Apply(ctx, id, pod); err != nil {
		klog.Errorf("failed to apply pod: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

// GetPod handles the endpoint for getting pod.
func (h *PodHandler) GetPod(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.Param("name")
	id := c.Param("simulatorID")

	p, err := h.service.Get(ctx, name, id)
	if err != nil {
		klog.Errorf("failed to get pod: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, p)
}

// ListPod handles the endpoint for listing pod.
func (h *PodHandler) ListPod(c echo.Context) error {
	ctx := c.Request().Context()

	id := c.Param("simulatorID")

	ps, err := h.service.List(ctx, id)
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
	id := c.Param("simulatorID")

	if err := h.service.Delete(ctx, name, id); err != nil {
		klog.Errorf("failed to delete pod: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}
