package handler

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
)

// PodHandler is handler for manage pod.
type PodHandler struct {
	service PodService
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
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if err := h.service.Apply(ctx, id, pod); err != nil {
		log.Println(err)
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
		log.Println(err)
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
		log.Println(err)
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
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}
