package handler

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/klog/v2"
)

// NamespaceHandler is handler for manage namespace.
type NamespaceHandler struct {
	service NamespaceService
}

// NamespaceService represents service for manage Namespaces.
type NamespaceService interface {
	Apply(ctx context.Context, namespace *v1.NamespaceApplyConfiguration) (*corev1.Namespace, error)
}

// NewNamespaceHandler initializes NamespaceHandler.
func NewNamespaceHandler(s NamespaceService) *NamespaceHandler {
	return &NamespaceHandler{service: s}
}

// CreateNamespace handles the endpoint for creating namespace.
// It creates unique namespace with uuid.
func (h *NamespaceHandler) CreateNamespace(c echo.Context) error {
	ctx := c.Request().Context()

	// create namespace with uuid
	namespace := v1.Namespace(uuid.New().String())
	n, err := h.service.Apply(ctx, namespace)
	if err != nil {
		klog.Errorf("failed to apply namespace: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, n)
}
