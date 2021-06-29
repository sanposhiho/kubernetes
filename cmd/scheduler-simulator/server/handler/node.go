package handler

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
)

// NodeHandler is handler for manage nodes.
type NodeHandler struct {
	service NodeService
}

// NewNodeHandler initializes NodeHandler.
func NewNodeHandler(s NodeService) *NodeHandler {
	return &NodeHandler{service: s}
}

// ApplyNode handles the endpoint for applying node.
func (h *NodeHandler) ApplyNode(c echo.Context) error {
	ctx := c.Request().Context()

	simulatorID := c.Param("simulatorID")

	reqNode := new(v1.NodeApplyConfiguration)
	if err := c.Bind(reqNode); err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if err := h.service.Apply(ctx, simulatorID, reqNode); err != nil {
		log.Println(err)
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
		log.Println(err)
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
		log.Println(err)
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
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}
