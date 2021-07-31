package pod

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	schedulerapi "k8s.io/kube-scheduler/config/v1beta1"

	"k8s.io/kubernetes/cmd/scheduler-simulator/errors"
	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler"
	"k8s.io/kubernetes/cmd/scheduler-simulator/scheduler/plugin/annotation"
	"k8s.io/kubernetes/cmd/scheduler-simulator/schedulerconfig"
	"k8s.io/kubernetes/cmd/scheduler-simulator/util"
	"k8s.io/kubernetes/pkg/scheduler/algorithmprovider"
	original "k8s.io/kubernetes/pkg/scheduler/apis/config"
)

// Service manages pods.
type Service struct {
	client                        clientset.Interface
	schedulerConfigurationService SchedulerConfigurationService
}

type SchedulerConfigurationService interface {
	GetSchedulerConfig(ctx context.Context, k string) (*schedulerapi.KubeSchedulerConfiguration, error)
}

// NewPodService initializes Service.
func NewPodService(client clientset.Interface, scs SchedulerConfigurationService) *Service {
	return &Service{
		client:                        client,
		schedulerConfigurationService: scs,
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

	sc, err := s.schedulerConfigurationService.GetSchedulerConfig(ctx, simulatorID)
	if err != nil {
		if !xerrors.Is(err, errors.ErrNotFound) {
			return xerrors.Errorf("get scheduler config: %w", err)
		}
		sc = schedulerconfig.DefaultSchedulerConfig()
	}

	if err := s.addSchedulerConfiguration(pod, sc); err != nil {
		return xerrors.Errorf("add scheduler configuration on pod annotation: %w", err)
	}

	applyFunc := func() (bool, error) {
		_, err := s.client.CoreV1().Pods(simulatorID).Apply(ctx, pod, metav1.ApplyOptions{Force: true, FieldManager: "simulator"})
		if err == nil || apierrors.IsAlreadyExists(err) {
			return true, nil
		}
		return false, xerrors.Errorf("apply pod: %w", err)
	}

	if err := util.RetryWithExponentialBackOff(applyFunc); err != nil {
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

	if err := util.RetryWithExponentialBackOff(deleteFunc); err != nil {
		return xerrors.Errorf("delete pod with retry: %w", err)
	}

	return nil
}

func (s *Service) addSchedulerConfiguration(pac *v1.PodApplyConfiguration, sc *schedulerapi.KubeSchedulerConfiguration) error {
	if pac.Annotations == nil {
		pac.Annotations = make(map[string]string)
	}

	if pac.Spec == nil || pac.Spec.SchedulerName == nil {
		defaultSchedulerName := scheduler.DefaultSchedulerName
		pac.Spec.SchedulerName = &defaultSchedulerName
	}

	pluginindex := -1
	for i, p := range sc.Profiles {
		if *p.SchedulerName != *pac.Spec.SchedulerName {
			continue
		}
		pluginindex = i
	}
	if pluginindex == -1 {
		return xerrors.New("plugin profile not found")
	}

	// only original.Plugins has Append and Apply method.
	defaultPlugins := algorithmprovider.GetDefaultConfig()
	plugins := &original.Plugins{}
	plugins.Append(defaultPlugins)
	plugins.Apply(schedulerconfig.ConvertToOriginalPluginsType(sc.Profiles[pluginindex].Plugins))

	j, err := json.Marshal(plugins)
	if err != nil {
		return xerrors.Errorf("encode to json: %w", err)
	}
	pac.Annotations[annotation.EnabledPluginsAnnotationKey] = string(j)

	// remove scheduler name to use `default-scheduler`.
	// This simulator has only one scheduler named default-scheduler, and it behaves as if there are multiple schedulers.
	pac.Annotations[annotation.SchedulerNameAnnotationKey] = *pac.Spec.SchedulerName
	pac.Spec.SchedulerName = nil

	return nil
}
