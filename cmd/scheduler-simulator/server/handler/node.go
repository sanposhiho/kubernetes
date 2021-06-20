package handler

import (
	"log"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"

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

// ApplyNode handles the endpoint for applying node.
func (h *NodeHandler) ApplyNode(c echo.Context) error {
	ctx := c.Request().Context()

	node := new(v1.NodeApplyConfiguration)
	// TODO: Allow only certain fields and make sure that no disallowed fields are requested.
	if err := c.Bind(node); err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	// FIXME: delete this
	name := "sample-node"
	phase := corev1.NodeRunning
	ready := corev1.NodeReady
	conditionTrue := corev1.ConditionTrue
	node = v1.Node(name)
	node.Status = &v1.NodeStatusApplyConfiguration{
		Capacity: &corev1.ResourceList{
			corev1.ResourcePods:   *resource.NewQuantity(110, resource.DecimalSI),
			corev1.ResourceCPU:    resource.MustParse("4"),
			corev1.ResourceMemory: resource.MustParse("32Gi"),
		},
		Phase: &phase,
		Conditions: []v1.NodeConditionApplyConfiguration{
			{Type: &ready, Status: &conditionTrue},
		},
	}

	n, err := h.service.Apply(ctx, node)
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
