package handler

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/klog/v2"

	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/plugin/annotation"
)

// PodHandler is handler for manage pod.
type PodHandler struct {
	service PodService
}

// PodService represents service for manage Pods.
type PodService interface {
	Get(ctx context.Context, name string, simulatorID string) (*corev1.Pod, error)
	List(ctx context.Context, simulatorID string) (*corev1.PodList, error)
	Apply(ctx context.Context, simulatorID string, pod *v1.PodApplyConfiguration) error
	Delete(ctx context.Context, name string, simulatorID string) error
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
		klog.Errorf("failed to bind apply pod request: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if err := h.service.Apply(ctx, id, pod); err != nil {
		klog.Errorf("failed to apply pod: %+v", err)
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
		klog.Errorf("failed to get pod: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if err := addSchedulerNameToPod(p); err != nil {
		klog.Errorf("failed to add scheduler name to pod: %+v", err)
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
		klog.Errorf("failed to list pods: %+v", err)
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
		klog.Errorf("failed to delete pod: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

// addSchedulerNameToPod adds scheduler name on .spec.schedulerName
// so that users can see what schedulerName they have used.
// When simulator creates pods, it removes .spec.schedulerName to use `default-scheduler`.
// This simulator has only one scheduler named default-scheduler,
// and it behaves as if there are multiple schedulers.
// Annotation of annotation.SchedulerNameAnnotationKey has scheduler name that the user specified.
func addSchedulerNameToPod(pod *corev1.Pod) error {
	schedulerName, ok := pod.Annotations[annotation.SchedulerNameAnnotationKey]
	if !ok {
		return xerrors.New("pod doesn't have scheduler name on annotation")
	}
	pod.Spec.SchedulerName = schedulerName
	return nil
}
