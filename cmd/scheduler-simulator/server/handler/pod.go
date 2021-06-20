package handler

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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

	pod := new(v1.PodApplyConfiguration)
	// TODO: Allow only certain fields and make sure that no disallowed fields are requested.
	if err := c.Bind(pod); err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	// FIXME: delete this
	name := "sample-pod"
	containerName := "pause"
	containerImage := "k8s.gcr.io/pause:3.5"
	pod = v1.Pod(name, "default")
	pod.Spec = &v1.PodSpecApplyConfiguration{
		Containers: []v1.ContainerApplyConfiguration{{
			Name:  &containerName,
			Image: &containerImage,
			Resources: &v1.ResourceRequirementsApplyConfiguration{
				Limits: &corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("16Gi"),
				},
				Requests: &corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("16Gi"),
				},
			},
		}},
	}

	n, err := h.service.Apply(ctx, pod)
	if err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, n)
}

// GetPod handles the endpoint for getting pod.
func (h *PodHandler) GetPod(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.Param("name")

	p, err := h.service.Get(ctx, name)
	if err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, p)
}

// ListPod handles the endpoint for listing pod.
func (h *PodHandler) ListPod(c echo.Context) error {
	ctx := c.Request().Context()

	ps, err := h.service.List(ctx)
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

	if err := h.service.Delete(ctx, name); err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}
