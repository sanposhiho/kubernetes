package handler

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

type NodeHandler struct {
	service NodeService
}

func NewNodeHandler(s NodeService) *NodeHandler {
	return &NodeHandler{service: s}
}

func (h *NodeHandler) CreateNode(c echo.Context) error {
	ctx := c.Request().Context()

	n, err := h.service.Create(ctx)
	if err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, n)
}

func (h *NodeHandler) ListNode(c echo.Context) error {
	ctx := c.Request().Context()

	ns, err := h.service.List(ctx)
	if err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, ns)
}
