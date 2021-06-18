package node_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"k8s.io/kubernetes/cmd/scheduler-simulator/node"
	mock_kubernetes "k8s.io/kubernetes/cmd/scheduler-simulator/node/mock_clientset"
	"k8s.io/kubernetes/cmd/scheduler-simulator/node/mock_node"
)

func TestService_Delete(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                    string
		nodeName                string
		preparePodServiceMockFn func(m *mock_node.MockPodService)
		prepareClientSetMockFn  func(m *mock_kubernetes.MockInterface)
		wantErr                 bool
	}{
		{
			name:     "delete node and pods on node",
			nodeName: "node1",
			preparePodServiceMockFn: func(m *mock_node.MockPodService) {
				m.EXPECT().List(gomock.Any()).Return(&v1.PodList{
					Items: []v1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pod1",
							},
							Spec: v1.PodSpec{
								NodeName: "node1",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pod2",
							},
							Spec: v1.PodSpec{
								NodeName: "node1",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "this-pod-will-not-be-deleted",
							},
							Spec: v1.PodSpec{
								NodeName: "other-node",
							},
						},
					},
				}, nil)
				m.EXPECT().Delete(gomock.Any(), "pod1").Return(nil)
				m.EXPECT().Delete(gomock.Any(), "pod2").Return(nil)
			},
			prepareClientSetMockFn: func(m *mock_kubernetes.MockInterface) {
				m.EXPECT().CoreV1().Times(1).Return((&fake.Clientset{}).CoreV1())
			},
			wantErr: false,
		},
		{
			name:     "one of deleting pods fail",
			nodeName: "node1",
			preparePodServiceMockFn: func(m *mock_node.MockPodService) {
				m.EXPECT().List(gomock.Any()).Return(&v1.PodList{
					Items: []v1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pod1",
							},
							Spec: v1.PodSpec{
								NodeName: "node1",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pod2",
							},
							Spec: v1.PodSpec{
								NodeName: "node1",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "this-pod-will-not-be-deleted",
							},
							Spec: v1.PodSpec{
								NodeName: "other-node",
							},
						},
					},
				}, nil)
				m.EXPECT().Delete(gomock.Any(), "pod1").Return(nil)
				m.EXPECT().Delete(gomock.Any(), "pod2").Return(errors.New("error"))
			},
			prepareClientSetMockFn: func(m *mock_kubernetes.MockInterface) {},
			wantErr:                true,
		},
		{
			name:     "delete node with no pods",
			nodeName: "node1",
			preparePodServiceMockFn: func(m *mock_node.MockPodService) {
				m.EXPECT().List(gomock.Any()).Return(&v1.PodList{Items: []v1.Pod{}}, nil)
			},
			prepareClientSetMockFn: func(m *mock_kubernetes.MockInterface) {
				m.EXPECT().CoreV1().Times(1).Return((&fake.Clientset{}).CoreV1())
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockPodService := mock_node.NewMockPodService(ctrl)
			tt.preparePodServiceMockFn(mockPodService)

			mockClientSet := mock_kubernetes.NewMockInterface(ctrl)
			tt.prepareClientSetMockFn(mockClientSet)

			s := node.NewNodeService(mockClientSet, mockPodService)
			if err := s.Delete(context.Background(), tt.nodeName); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
