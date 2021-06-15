package handler

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

type PodHandler struct {
	service PodService
}

func NewPodHandler(s PodService) *PodHandler {
	return &PodHandler{service: s}
}

func (h *PodHandler) CreatePod(c echo.Context) error {
	ctx := c.Request().Context()

	p, err := h.service.Create(ctx)
	if err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, p)
}

func (h *PodHandler) ListPod(c echo.Context) error {
	ctx := c.Request().Context()

	ps, err := h.service.List(ctx)
	if err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, ps)
}
