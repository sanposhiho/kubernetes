package namespace

import (
	"context"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

type Service struct {
	client clientset.Interface
}

func NewNamespaceService(client clientset.Interface) *Service {
	return &Service{
		client: client,
	}
}

func (s *Service) Get(ctx context.Context, name string) (*v1.Namespace, error) {
	n, err := s.client.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, xerrors.Errorf("get namespace: %w", err)
	}
	return n, nil
}

func (s *Service) List(ctx context.Context) (*v1.NamespaceList, error) {
	pl, err := s.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, xerrors.Errorf("list namespaces: %w", err)
	}
	return pl, nil
}

func (s *Service) Create(ctx context.Context) (*v1.Namespace, error) {
	ns := &v1.Namespace{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       v1.NamespaceSpec{},
		Status:     v1.NamespaceStatus{},
	}

	namespace, err := s.client.CoreV1().Namespaces().Create(context.TODO(), basePod, metav1.CreateOptions{})
	if err != nil {
		return nil, xerrors.Errorf("create namespace: %w", err)
	}
	return namespace, nil
}

func (s *Service) Delete(ctx context.Context, name string) error {
	noGrace := int64(0)
	if err := s.client.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{
		GracePeriodSeconds: &noGrace,
	}); err != nil {
		return xerrors.Errorf("delete namespace: %w", err)
	}
	return nil
}
