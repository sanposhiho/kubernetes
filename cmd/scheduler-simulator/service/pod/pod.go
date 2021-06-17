package pod

import (
	"context"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
)

type Service struct {
	client      clientset.Interface
	podInformer coreinformers.PodInformer
}

func NewPodService(client clientset.Interface, podInformer coreinformers.PodInformer) *Service {
	return &Service{
		client:      client,
		podInformer: podInformer,
	}
}

func (s *Service) Get(ctx context.Context, name string) (*v1.Pod, error) {
	n, err := s.client.CoreV1().Pods("default").Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, xerrors.Errorf("get pod: %w", err)
	}
	return n, nil
}

func (s *Service) List(ctx context.Context) (*v1.PodList, error) {
	pl, err := s.client.CoreV1().Pods("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, xerrors.Errorf("list pods: %w", err)
	}
	return pl, nil
}

func (s *Service) Create(ctx context.Context) (*v1.Pod, error) {
	// TODO: users specify specs.
	basePod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "sample-pod-",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{{
				Name:  "pause",
				Image: "k8s.gcr.io/pause:3.5",
				Ports: []v1.ContainerPort{{ContainerPort: 80}},
				Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("100m"),
						v1.ResourceMemory: resource.MustParse("16Gi"),
					},
					Requests: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("100m"),
						v1.ResourceMemory: resource.MustParse("16Gi"),
					},
				},
			}},
		},
	}

	pod, err := s.client.CoreV1().Pods("default").Create(context.TODO(), basePod, metav1.CreateOptions{})
	// TODO: do retry?
	if err != nil {
		return nil, xerrors.Errorf("create pod: %w", err)
	}
	return pod, nil
}

func (s *Service) Delete(ctx context.Context, name string) error {
	noGrace := int64(0)
	if err := s.client.CoreV1().Pods("default").Delete(ctx, name, metav1.DeleteOptions{
		GracePeriodSeconds: &noGrace,
	}); err != nil {
		return xerrors.Errorf("delete pod: %w", err)
	}
	return nil
}
