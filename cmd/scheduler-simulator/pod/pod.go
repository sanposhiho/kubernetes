package pod

import (
	"context"
	"fmt"
	"time"

	"k8s.io/kubernetes/cmd/scheduler-simulator/node"

	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
)

// Service manages pods.
type Service struct {
	client      clientset.Interface
	podInformer coreinformers.PodInformer
}

// NewPodService initializes Service.
func NewPodService(client clientset.Interface, podInformer coreinformers.PodInformer) *Service {
	return &Service{
		client:      client,
		podInformer: podInformer,
	}
}

// Get returns the pod has given name.
// use simulatorID as namespace.
func (s *Service) Get(ctx context.Context, name string, simulatorID string) (*corev1.Pod, error) {
	n, err := s.client.CoreV1().Pods(simulatorID).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, xerrors.Errorf("get pod: %w", err)
	}
	return n, nil
}

// List list all pods.
// use simulatorID as namespace.
func (s *Service) List(ctx context.Context, simulatorID string) (*corev1.PodList, error) {
	pl, err := s.client.CoreV1().Pods(simulatorID).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, xerrors.Errorf("list pods: %w", err)
	}
	return pl, nil
}

// Apply applies the pod.
// use simulatorID as namespace.
func (s *Service) Apply(ctx context.Context, simulatorID string, pod *v1.PodApplyConfiguration) error {
	pod.WithKind("Pod")
	pod.WithAPIVersion("v1")

	addSimulatorIDNodeSelector(pod, simulatorID)

	applyFunc := func() (bool, error) {
		_, err := s.client.CoreV1().Pods(simulatorID).Apply(ctx, pod, metav1.ApplyOptions{Force: true, FieldManager: "simulator"})
		if err == nil || apierrors.IsAlreadyExists(err) {
			return true, nil
		}
		return false, xerrors.Errorf("apply pod: %w", err)
	}

	if err := RetryWithExponentialBackOff(applyFunc); err != nil {
		return xerrors.Errorf("apply pod with retry: %w", err)
	}

	return nil
}

// Delete deletes the pod has given name.
// use simulatorID as namespace.
func (s *Service) Delete(ctx context.Context, name string, simulatorID string) error {
	noGrace := int64(0)
	deleteFunc := func() (bool, error) {
		err := s.client.CoreV1().Pods(simulatorID).Delete(ctx, name, metav1.DeleteOptions{
			// need to use noGrace to avoid waiting kubelet checking.
			// > When a force deletion is performed, the API server does not wait for confirmation from the kubelet that
			//   the Pod has been terminated on the node it was running on.
			// https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination-forced
			GracePeriodSeconds: &noGrace,
		})
		if err == nil || apierrors.IsNotFound(err) {
			return true, nil
		}
		return false, fmt.Errorf("delete pod: %w", err)
	}

	if err := RetryWithExponentialBackOff(deleteFunc); err != nil {
		return xerrors.Errorf("delete pod with retry: %w", err)
	}

	return nil
}

const (
	// Parameters for retrying with exponential backoff.
	retryBackoffInitialDuration = 100 * time.Millisecond
	retryBackoffFactor          = 3
	retryBackoffJitter          = 0
	retryBackoffSteps           = 6
)

// RetryWithExponentialBackOff is the utility for retrying the given function with exponential backoff.
func RetryWithExponentialBackOff(fn wait.ConditionFunc) error {
	backoff := wait.Backoff{
		Duration: retryBackoffInitialDuration,
		Factor:   retryBackoffFactor,
		Jitter:   retryBackoffJitter,
		Steps:    retryBackoffSteps,
	}
	return wait.ExponentialBackoff(backoff, fn)
}

func addSimulatorIDNodeSelector(pac *v1.PodApplyConfiguration, simulatorID string) {
	pac.Spec.NodeSelector = map[string]string{
		node.SimulatorIDLabelKey: simulatorID,
	}
}
