package schedulingresultstore

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"k8s.io/kubernetes/cmd/scheduler-simulator/util"
)

type Store struct {
	mu *sync.Mutex

	client clientset.Interface
	result map[key]*result
}

type result struct {
	// node name → plugin name → score
	score map[string]map[string]int64

	// node name → plugin name → normalizedScore
	normalizedScore map[string]map[string]int64

	// node name → plugin name → filtering result
	// - When node pass the filter, filtering result is "pass".
	// - When node blocked by the filter, filtering result is blocked reason.
	filter map[string]map[string]string
}

// key is the key of result map on Store.
// key is created from namespace and podName.
type key string

// newKey creates key with namespace and podName.
func newKey(namespace, podName string) key {
	k := namespace + "/" + podName
	return key(k)
}

func newData() *result {
	d := &result{
		score:           map[string]map[string]int64{},
		normalizedScore: map[string]map[string]int64{},
		filter:          map[string]map[string]string{},
	}
	return d
}

func New(informerFactory informers.SharedInformerFactory, client clientset.Interface) *Store {
	s := &Store{
		mu:     new(sync.Mutex),
		client: client,
		result: map[key]*result{},
	}

	informerFactory.Core().V1().Pods().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			UpdateFunc: s.addSchedulingResultToPod,
		},
	)

	return s
}

const (
	filterResultAnnotationKey          = "scheduler-simulator/filter-result"
	scoreResultAnnotationKey           = "scheduler-simulator/score-result"
	normalizedScoreResultAnnotationKey = "scheduler-simulator/normalizedscore-result"
)

func (s *Store) addSchedulingResultToPod(oldObj, newObj interface{}) {
	ctx := context.Background()

	pod, ok := newObj.(*v1.Pod)
	if !ok {
		klog.ErrorS(nil, "Cannot convert to *v1.Pod", "obj", newObj)
		return
	}

	_, ok = pod.Annotations[scoreResultAnnotationKey]
	_, ok2 := pod.Annotations[normalizedScoreResultAnnotationKey]
	if ok && ok2 {
		// already have scheduling result
		return
	}

	k := newKey(pod.Namespace, pod.Name)
	if _, ok := s.result[k]; !ok {
		// not scheduled yet
		return
	}
	defer func() {
		s.DeleteData(k)
	}()

	if err := s.addFilterResultToPod(pod); err != nil {
		klog.Errorf("failed to add filtering result to pod: %v", err)
		return
	}

	if err := s.addScoreResultToPod(pod); err != nil {
		klog.Errorf("failed to add scoring result to pod: %v", err)
		return
	}

	if err := s.addNormalizedScoreResultToPod(pod); err != nil {
		klog.Errorf("failed to add normalized score result to pod: %v", err)
		return
	}

	updateFunc := func() (bool, error) {
		_, err := s.client.CoreV1().Pods(pod.Namespace).Update(ctx, pod, metav1.UpdateOptions{})
		if err != nil {
			return false, xerrors.Errorf("update pod: %v", err)
		}

		return true, nil
	}
	if err := util.RetryWithExponentialBackOff(updateFunc); err != nil {
		klog.Errorf("failed to update pod with retry to record score: %v", err)
		return
	}
}

func (s *Store) addFilterResultToPod(pod *v1.Pod) error {
	k := newKey(pod.Namespace, pod.Name)
	scores, err := json.Marshal(s.result[k].filter)
	if err != nil {
		return fmt.Errorf("encode json to record scores: %w", err)
	}

	metav1.SetMetaDataAnnotation(&pod.ObjectMeta, filterResultAnnotationKey, string(scores))
	return nil
}

func (s *Store) addScoreResultToPod(pod *v1.Pod) error {
	k := newKey(pod.Namespace, pod.Name)
	scores, err := json.Marshal(s.result[k].score)
	if err != nil {
		return fmt.Errorf("encode json to record scores: %w", err)
	}

	metav1.SetMetaDataAnnotation(&pod.ObjectMeta, scoreResultAnnotationKey, string(scores))
	return nil
}

func (s *Store) addNormalizedScoreResultToPod(pod *v1.Pod) error {
	k := newKey(pod.Namespace, pod.Name)
	scores, err := json.Marshal(s.result[k].normalizedScore)
	if err != nil {
		return fmt.Errorf("encode json to record scores: %w", err)
	}

	metav1.SetMetaDataAnnotation(&pod.ObjectMeta, normalizedScoreResultAnnotationKey, string(scores))
	return nil
}

func (s *Store) AddFilterResult(namespace, podName, nodeName, pluginName, reason string) {
	if !strings.HasSuffix(nodeName, namespace) {
		// when suffix of nodeName don't match namespace, this node is not created by a user who creates this pod.
		// So we don't need to record the result in this case.
		return
	}
	if reason == "" {
		// empty reason means the node passed the filter.
		reason = "pass"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	k := newKey(namespace, podName)
	if _, ok := s.result[k]; !ok {
		s.result[k] = newData()
	}

	if _, ok := s.result[k].filter[nodeName]; !ok {
		s.result[k].filter[nodeName] = map[string]string{}
	}

	s.result[k].filter[nodeName][pluginName] = reason
}

// AddScoreResult adds scoring result to pod annotation.
// When score is -1, it means the plugin is disabled.
func (s *Store) AddScoreResult(namespace, podName, nodeName, pluginName string, score int64) {
	if !strings.HasSuffix(nodeName, namespace) {
		// when suffix of nodeName don't match namespace, this node is not created by a user who creates this pod.
		// So we don't need to record the result in this case.
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	k := newKey(namespace, podName)
	if _, ok := s.result[k]; !ok {
		s.result[k] = newData()
	}

	if _, ok := s.result[k].score[nodeName]; !ok {
		s.result[k].score[nodeName] = map[string]int64{}
	}
	if _, ok := s.result[k].normalizedScore[nodeName]; !ok {
		s.result[k].normalizedScore[nodeName] = map[string]int64{}
	}

	s.result[k].score[nodeName][pluginName] = score
	s.result[k].normalizedScore[nodeName][pluginName] = score
}

// AddNormalizeScoreResult adds normalized score result to pod annotation.
// When normalizedScore is -1, it means the plugin is disabled.
func (s *Store) AddNormalizeScoreResult(namespace, podName, nodeName, pluginName string, normalizedScore int64) {
	if !strings.HasSuffix(nodeName, namespace) {
		// when suffix of nodeName don't match namespace, this node is not created by a user who creates this pod.
		// So we don't need to record the result in this case.
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	k := newKey(namespace, podName)
	if _, ok := s.result[k]; !ok {
		s.result[k] = newData()
	}

	if _, ok := s.result[k].normalizedScore[nodeName]; !ok {
		s.result[k].normalizedScore[nodeName] = map[string]int64{}
	}

	s.result[k].normalizedScore[nodeName][pluginName] = normalizedScore
}

func (s *Store) DeleteData(k key) {
	delete(s.result, k)
}
