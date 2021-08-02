package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	v1 "k8s.io/client-go/applyconfigurations/storage/v1"
	"k8s.io/klog/v2"

	"k8s.io/kubernetes/cmd/scheduler-simulator/server/di"
)

// StorageClassHandler is handler for manage storageClass.
type StorageClassHandler struct {
	service di.StorageClassService
}

// NewStorageClassHandler initializes StorageClassHandler.
func NewStorageClassHandler(s di.StorageClassService) *StorageClassHandler {
	return &StorageClassHandler{service: s}
}

// ApplyStorageClass handles the endpoint for applying storageClass.
func (h *StorageClassHandler) ApplyStorageClass(c echo.Context) error {
	ctx := c.Request().Context()

	id := c.Param("simulatorID")

	storageClass := new(v1.StorageClassApplyConfiguration)
	if err := c.Bind(storageClass); err != nil {
		klog.Errorf("failed to bind apply storageClass request: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if err := h.service.Apply(ctx, id, storageClass); err != nil {
		klog.Errorf("failed to apply storageClass: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

// GetStorageClass handles the endpoint for getting storageClass.
func (h *StorageClassHandler) GetStorageClass(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.Param("name")
	id := c.Param("simulatorID")

	p, err := h.service.Get(ctx, name, id)
	if err != nil {
		klog.Errorf("failed to get storageClass: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, p)
}

// ListStorageClass handles the endpoint for listing storageClass.
func (h *StorageClassHandler) ListStorageClass(c echo.Context) error {
	ctx := c.Request().Context()

	id := c.Param("simulatorID")

	ps, err := h.service.List(ctx, id)
	if err != nil {
		klog.Errorf("failed to list storageClasss: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, ps)
}

// DeleteStorageClass handles the endpoint for deleting storageClass.
func (h *StorageClassHandler) DeleteStorageClass(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.Param("name")
	id := c.Param("simulatorID")

	if err := h.service.Delete(ctx, name, id); err != nil {
		klog.Errorf("failed to delete storageClass: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}
