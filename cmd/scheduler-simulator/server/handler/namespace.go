package handler

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
)

// NamespaceHandler is handler for manage namespace.
type NamespaceHandler struct {
	service NamespaceService
}

// NewNamespaceHandler initializes NamespaceHandler.
func NewNamespaceHandler(s NamespaceService) *NamespaceHandler {
	return &NamespaceHandler{service: s}
}

// ApplyNamespace handles the endpoint for applying namespace.
func (h *NamespaceHandler) ApplyNamespace(c echo.Context) error {
	ctx := c.Request().Context()

	namespace := new(v1.NamespaceApplyConfiguration)
	if err := c.Bind(namespace); err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	n, err := h.service.Apply(ctx, namespace)
	if err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, n)
}
