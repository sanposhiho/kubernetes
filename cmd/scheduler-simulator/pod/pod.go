package pod

import (
	"context"

	v1 "k8s.io/client-go/applyconfigurations/core/v1"

	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
)

const defaultNameSpace = "default"

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
func (s *Service) Get(ctx context.Context, name string) (*corev1.Pod, error) {
	n, err := s.client.CoreV1().Pods(defaultNameSpace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, xerrors.Errorf("get pod: %w", err)
	}
	return n, nil
}

// List list all pods.
func (s *Service) List(ctx context.Context) (*corev1.PodList, error) {
	pl, err := s.client.CoreV1().Pods(defaultNameSpace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, xerrors.Errorf("list pods: %w", err)
	}
	return pl, nil
}

func (s *Service) Apply(ctx context.Context, pod *v1.PodApplyConfiguration) (*corev1.Pod, error) {
	newPod, err := s.client.CoreV1().Pods(defaultNameSpace).Apply(ctx, pod, metav1.ApplyOptions{Force: true, FieldManager: "simulator"})
	if err != nil {
		return nil, xerrors.Errorf("apply pods: %w", err)
	}

	return newPod, nil
}

// Delete deletes the pod has given name.
func (s *Service) Delete(ctx context.Context, name string) error {
	noGrace := int64(0)
	if err := s.client.CoreV1().Pods(defaultNameSpace).Delete(ctx, name, metav1.DeleteOptions{
		GracePeriodSeconds: &noGrace,
	}); err != nil {
		return xerrors.Errorf("delete pod: %w", err)
	}
	return nil
}
