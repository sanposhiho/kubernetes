package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/klog/v2"

	"k8s.io/kubernetes/cmd/scheduler-simulator/server/di"
)

// PersistentVolumeClaimHandler is handler for manage persistentVolumeClaim.
type PersistentVolumeClaimHandler struct {
	service di.PersistentVolumeClaimService
}

// NewPersistentVolumeClaimHandler initializes PersistentVolumeClaimHandler.
func NewPersistentVolumeClaimHandler(s di.PersistentVolumeClaimService) *PersistentVolumeClaimHandler {
	return &PersistentVolumeClaimHandler{service: s}
}

// ApplyPersistentVolumeClaim handles the endpoint for applying persistentVolumeClaim.
func (h *PersistentVolumeClaimHandler) ApplyPersistentVolumeClaim(c echo.Context) error {
	ctx := c.Request().Context()

	id := c.Param("simulatorID")

	persistentVolumeClaim := new(v1.PersistentVolumeClaimApplyConfiguration)
	if err := c.Bind(persistentVolumeClaim); err != nil {
		klog.Errorf("failed to bind apply persistentVolumeClaim request: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if err := h.service.Apply(ctx, id, persistentVolumeClaim); err != nil {
		klog.Errorf("failed to apply persistentVolumeClaim: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

// GetPersistentVolumeClaim handles the endpoint for getting persistentVolumeClaim.
func (h *PersistentVolumeClaimHandler) GetPersistentVolumeClaim(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.Param("name")
	id := c.Param("simulatorID")

	p, err := h.service.Get(ctx, name, id)
	if err != nil {
		klog.Errorf("failed to get persistentVolumeClaim: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, p)
}

// ListPersistentVolumeClaim handles the endpoint for listing persistentVolumeClaim.
func (h *PersistentVolumeClaimHandler) ListPersistentVolumeClaim(c echo.Context) error {
	ctx := c.Request().Context()

	id := c.Param("simulatorID")

	ps, err := h.service.List(ctx, id)
	if err != nil {
		klog.Errorf("failed to list persistentVolumeClaims: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, ps)
}

// DeletePersistentVolumeClaim handles the endpoint for deleting persistentVolumeClaim.
func (h *PersistentVolumeClaimHandler) DeletePersistentVolumeClaim(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.Param("name")
	id := c.Param("simulatorID")

	if err := h.service.Delete(ctx, name, id); err != nil {
		klog.Errorf("failed to delete persistentVolumeClaim: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}
