package pod

import (
	"context"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
)

type Service struct {
	Client      clientset.Interface
	PodInformer coreinformers.PodInformer
}

func NewPodService(client clientset.Interface, podInformer coreinformers.PodInformer) *Service {
	return &Service{
		Client:      client,
		PodInformer: podInformer,
	}
}

func (s *Service) Create(ctx context.Context) (*v1.Pod, error) {
	basePod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "sample-pod-",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{{
				Name:  "pause",
				Image: "k8s.gcr.io/pause:3.5",
				//              Ports: []v1.ContainerPort{{ContainerPort: 80}},
				//				Resources: v1.ResourceRequirements{
				//					Limits: v1.ResourceList{
				//						v1.ResourceCPU:    resource.MustParse("100m"),
				//						v1.ResourceMemory: resource.MustParse("500Mi"),
				//					},
				//					Requests: v1.ResourceList{
				//						v1.ResourceCPU:    resource.MustParse("100m"),
				//						v1.ResourceMemory: resource.MustParse("500Mi"),
				//					},
				//			},
			}},
		},
	}

	pod, err := s.Client.CoreV1().Pods("default").Create(context.TODO(), basePod, metav1.CreateOptions{})
	// TODO: do retry?
	if err != nil {
		return nil, xerrors.Errorf("create pod: %w", err)
	}
	return pod, nil
}

func (s *Service) List(ctx context.Context) (*v1.PodList, error) {
	return s.Client.CoreV1().Pods("default").List(ctx, metav1.ListOptions{})
}
