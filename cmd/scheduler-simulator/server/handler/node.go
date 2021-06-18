package handler

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

// NodeHandler is handler for manage nodes.
type NodeHandler struct {
	service NodeService
}

// NewNodeHandler initializes NodeHandler.
func NewNodeHandler(s NodeService) *NodeHandler {
	return &NodeHandler{service: s}
}

// CreateNode handles the endpoint for creating node.
func (h *NodeHandler) CreateNode(c echo.Context) error {
	ctx := c.Request().Context()

	n, err := h.service.Create(ctx)
	if err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, n)
}

// GetNode handles the endpoint for getting node.
func (h *NodeHandler) GetNode(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.Param("name")

	n, err := h.service.Get(ctx, name)
	if err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, n)
}

// ListNode handles the endpoint for listing node.
func (h *NodeHandler) ListNode(c echo.Context) error {
	ctx := c.Request().Context()

	ns, err := h.service.List(ctx)
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

	if err := h.service.Delete(ctx, name); err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}
