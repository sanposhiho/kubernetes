package namespace

import (
	"context"

	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	clientset "k8s.io/client-go/kubernetes"
)

// Service manages namespace.
type Service struct {
	client clientset.Interface
}

// NewNamespaceService initializes Service.
func NewNamespaceService(client clientset.Interface) *Service {
	return &Service{
		client: client,
	}
}

func (s *Service) Apply(ctx context.Context, namespace *v1.NamespaceApplyConfiguration) (*corev1.Namespace, error) {
	namespace.WithKind("Namespace")
	namespace.WithAPIVersion("v1")
	newNamespace, err := s.client.CoreV1().Namespaces().Apply(ctx, namespace, metav1.ApplyOptions{Force: true, FieldManager: "simulator"})
	if err != nil {
		return nil, xerrors.Errorf("apply namespaces: %w", err)
	}

	return newNamespace, nil
}
