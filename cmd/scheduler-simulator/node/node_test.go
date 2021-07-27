package node_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	applymetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"k8s.io/kubernetes/cmd/scheduler-simulator/node"
	"k8s.io/kubernetes/cmd/scheduler-simulator/node/mock_node"
)

func TestService_Delete(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                    string
		nodeName                string
		simulatorID             string
		preparePodServiceMockFn func(m *mock_node.MockPodService)
		prepareFakeClientSetFn  func() *fake.Clientset
		wantErr                 bool
	}{
		{
			name:        "delete node and pods on node",
			nodeName:    "node1",
			simulatorID: "simuid",
			preparePodServiceMockFn: func(m *mock_node.MockPodService) {
				m.EXPECT().List(gomock.Any(), "simuid").Return(&corev1.PodList{
					Items: []corev1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pod1",
							},
							Spec: corev1.PodSpec{
								NodeName: "node1",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pod2",
							},
							Spec: corev1.PodSpec{
								NodeName: "node1",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "this-pod-will-not-be-deleted",
							},
							Spec: corev1.PodSpec{
								NodeName: "other-node",
							},
						},
					},
				}, nil)
				m.EXPECT().Delete(gomock.Any(), "pod1", "simuid").Return(nil)
				m.EXPECT().Delete(gomock.Any(), "pod2", "simuid").Return(nil)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				// add test data.
				c.CoreV1().Nodes().Create(context.Background(), &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
					},
				}, metav1.CreateOptions{})
				return c
			},
			wantErr: false,
		},
		{
			name:        "one of deleting pods fail",
			nodeName:    "node1",
			simulatorID: "simuid",
			preparePodServiceMockFn: func(m *mock_node.MockPodService) {
				m.EXPECT().List(gomock.Any(), "simuid").Return(&corev1.PodList{
					Items: []corev1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pod1",
							},
							Spec: corev1.PodSpec{
								NodeName: "node1",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pod2",
							},
							Spec: corev1.PodSpec{
								NodeName: "node1",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "this-pod-will-not-be-deleted",
							},
							Spec: corev1.PodSpec{
								NodeName: "other-node",
							},
						},
					},
				}, nil)
				m.EXPECT().Delete(gomock.Any(), "pod1", "simuid").Return(nil)
				m.EXPECT().Delete(gomock.Any(), "pod2", "simuid").Return(errors.New("error"))
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			wantErr: true,
		},
		{
			name:        "delete node with no pods",
			nodeName:    "node1",
			simulatorID: "simuid",
			preparePodServiceMockFn: func(m *mock_node.MockPodService) {
				m.EXPECT().List(gomock.Any(), "simuid").Return(&corev1.PodList{Items: []corev1.Pod{}}, nil)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				// add test data.
				c.CoreV1().Nodes().Create(context.Background(), &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
					},
				}, metav1.CreateOptions{})
				return c
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
			fakeclientset := tt.prepareFakeClientSetFn()

			s := node.NewNodeService(fakeclientset, mockPodService)
			if err := s.Delete(context.Background(), tt.nodeName, tt.simulatorID); (err != nil) != tt.wantErr {
				t.Fatalf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_Apply(t *testing.T) {
	t.Parallel()
	nodeNameWithoutSuffix := "node1"
	tests := []struct {
		name                   string
		simulatorID            string
		nac                    *v1.NodeApplyConfiguration
		prepareFakeClientSetFn func() *fake.Clientset
		wantNodeName           string
		wantErr                bool
	}{
		{
			name:        "apply node whose name doesn't have suffix(=simulatorID)",
			simulatorID: "simulator1",
			nac: &v1.NodeApplyConfiguration{
				ObjectMetaApplyConfiguration: &applymetav1.ObjectMetaApplyConfiguration{Name: &nodeNameWithoutSuffix},
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			wantNodeName: "node1-simulator1",
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockPodService := mock_node.NewMockPodService(ctrl)
			fakeclientset := tt.prepareFakeClientSetFn()

			s := node.NewNodeService(fakeclientset, mockPodService)
			if err := s.Apply(context.Background(), tt.simulatorID, tt.nac); (err != nil) != tt.wantErr {
				t.Fatalf("Apply() error = %v, wantErr %v", err, tt.wantErr)
			}

			n, err := fakeclientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
			if err != nil {
				t.Fatalf("failed to get nodes from fake clientset: %v", err)
			}
			if len(n.Items) != 1 {
				t.Fatalf("the number of nodes on fake clientset must be 1")
			}
			if n.Items[0].Name != tt.wantNodeName {
				t.Fatalf("applied node name %s doesn't match wantNodeName %s", n.Items[0].Name, tt.wantNodeName)
			}
		})
	}
}
