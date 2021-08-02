package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/klog/v2"

	"k8s.io/kubernetes/cmd/scheduler-simulator/server/di"
)

// NodeHandler is handler for manage nodes.
type NodeHandler struct {
	service di.NodeService
}

// NewNodeHandler initializes NodeHandler.
func NewNodeHandler(s di.NodeService) *NodeHandler {
	return &NodeHandler{service: s}
}

// ApplyNode handles the endpoint for applying node.
func (h *NodeHandler) ApplyNode(c echo.Context) error {
	ctx := c.Request().Context()

	simulatorID := c.Param("simulatorID")

	reqNode := new(v1.NodeApplyConfiguration)
	if err := c.Bind(reqNode); err != nil {
		klog.Errorf("failed to bind apply node request: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if err := h.service.Apply(ctx, simulatorID, reqNode); err != nil {
		klog.Errorf("failed to apply node: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

// GetNode handles the endpoint for getting node.
func (h *NodeHandler) GetNode(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.Param("name")
	simulatorID := c.Param("simulatorID")

	n, err := h.service.Get(ctx, name, simulatorID)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		klog.Errorf("failed to get node: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, n)
}

// ListNode handles the endpoint for listing node.
func (h *NodeHandler) ListNode(c echo.Context) error {
	ctx := c.Request().Context()

	simulatorID := c.Param("simulatorID")

	ns, err := h.service.List(ctx, simulatorID)
	if err != nil {
		klog.Errorf("failed to list nodes: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, ns)
}

// DeleteNode handles the endpoint for deleting node.
func (h *NodeHandler) DeleteNode(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.Param("name")
	simulatorID := c.Param("simulatorID")

	if err := h.service.Delete(ctx, name, simulatorID); err != nil {
		klog.Errorf("failed to delete node: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}
