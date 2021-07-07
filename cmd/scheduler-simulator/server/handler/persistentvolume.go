package handler

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/klog/v2"
)

// PersistentVolumeHandler is handler for manage persistentVolume.
type PersistentVolumeHandler struct {
	service PersistentVolumeService
}

// PersistentVolumeService represents service for manage Pods.
type PersistentVolumeService interface {
	Get(ctx context.Context, name string, simulatorID string) (*corev1.PersistentVolume, error)
	List(ctx context.Context, simulatorID string) (*corev1.PersistentVolumeList, error)
	Apply(ctx context.Context, simulatorID string, pv *v1.PersistentVolumeApplyConfiguration) error
	Delete(ctx context.Context, name string, simulatorID string) error
}

// NewPersistentVolumeHandler initializes PersistentVolumeHandler.
func NewPersistentVolumeHandler(s PersistentVolumeService) *PersistentVolumeHandler {
	return &PersistentVolumeHandler{service: s}
}

// ApplyPersistentVolume handles the endpoint for applying persistentVolume.
func (h *PersistentVolumeHandler) ApplyPersistentVolume(c echo.Context) error {
	ctx := c.Request().Context()

	id := c.Param("simulatorID")

	persistentVolume := new(v1.PersistentVolumeApplyConfiguration)
	if err := c.Bind(persistentVolume); err != nil {
		klog.Errorf("failed to bind apply persistentVolume request: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if err := h.service.Apply(ctx, id, persistentVolume); err != nil {
		klog.Errorf("failed to apply persistentVolume: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

// GetPersistentVolume handles the endpoint for getting persistentVolume.
func (h *PersistentVolumeHandler) GetPersistentVolume(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.Param("name")
	id := c.Param("simulatorID")

	p, err := h.service.Get(ctx, name, id)
	if err != nil {
		klog.Errorf("failed to get persistentVolume: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, p)
}

// ListPersistentVolume handles the endpoint for listing persistentVolume.
func (h *PersistentVolumeHandler) ListPersistentVolume(c echo.Context) error {
	ctx := c.Request().Context()

	id := c.Param("simulatorID")

	ps, err := h.service.List(ctx, id)
	if err != nil {
		klog.Errorf("failed to list persistentVolumes: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, ps)
}

// DeletePersistentVolume handles the endpoint for deleting persistentVolume.
func (h *PersistentVolumeHandler) DeletePersistentVolume(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.Param("name")
	id := c.Param("simulatorID")

	if err := h.service.Delete(ctx, name, id); err != nil {
		klog.Errorf("failed to delete persistentVolume: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}
