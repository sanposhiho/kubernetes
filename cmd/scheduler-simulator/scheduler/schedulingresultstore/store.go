package schedulingresultstore

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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

	client            clientset.Interface
	result            map[key]*result
	scorePluginWeight map[string]int32
}

const (
	// DisabledMessage is used when plugin is disabled.
	DisabledMessage = "(disabled)"
	// DisabledScore means the scoring plugin is disabled.
	DisabledScore = -1

	// PassedFilterMessage is used when node pass the filter plugin.
	PassedFilterMessage = "passed"
)

// result has a scheduling result of pod.
type result struct {
	// node name → plugin name → score(string)
	// When the plugin is disabled, score will be DisabledMessage.
	score map[string]map[string]string

	// node name → plugin name → finalscore(string)
	// This score is normalized and applied weight for each plugins.
	// When the plugin is disabled, score will be DisabledMessage.
	finalscore map[string]map[string]string

	// node name → plugin name → filtering result
	// When node pass the filter, filtering result will be PassedFilterMessage.
	// When node blocked by the filter, filtering result is blocked reason.
	// When the plugin is disabled, score will be DisabledMessage.
	filter map[string]map[string]string
}

func New(informerFactory informers.SharedInformerFactory, client clientset.Interface, scorePluginWeight map[string]int32) *Store {
	s := &Store{
		mu:                new(sync.Mutex),
		client:            client,
		result:            map[key]*result{},
		scorePluginWeight: scorePluginWeight,
	}

	// Store adds scheduling results when pod is updating.
	informerFactory.Core().V1().Pods().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			UpdateFunc: s.addSchedulingResultToPod,
		},
	)

	return s
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
		score:      map[string]map[string]string{},
		finalscore: map[string]map[string]string{},
		filter:     map[string]map[string]string{},
	}
	return d
}

const (
	filterResultAnnotationKey     = "scheduler-simulator/filter-result"
	scoreResultAnnotationKey      = "scheduler-simulator/score-result"
	finalScoreResultAnnotationKey = "scheduler-simulator/finalscore-result"
)

func (s *Store) addSchedulingResultToPod(_, newObj interface{}) {
	ctx := context.Background()

	pod, ok := newObj.(*v1.Pod)
	if !ok {
		klog.ErrorS(nil, "Cannot convert to *v1.Pod", "obj", newObj)
		return
	}

	_, ok = pod.Annotations[scoreResultAnnotationKey]
	_, ok2 := pod.Annotations[finalScoreResultAnnotationKey]
	if ok && ok2 {
		// Pod already have scheduling result
		return
	}

	k := newKey(pod.Namespace, pod.Name)
	if _, ok := s.result[k]; !ok {
		// Store doesn't have scheduling result of pod.
		return
	}

	if err := s.addFilterResultToPod(pod); err != nil {
		klog.Errorf("failed to add filtering result to pod: %v", err)
		return
	}

	if err := s.addScoreResultToPod(pod); err != nil {
		klog.Errorf("failed to add scoring result to pod: %v", err)
		return
	}

	if err := s.addFinalScoreResultToPod(pod); err != nil {
		klog.Errorf("failed to add final score result to pod: %v", err)
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

	// delete data from Store only if data is successfully added on pod's annotations.
	s.DeleteData(k)
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

func (s *Store) addFinalScoreResultToPod(pod *v1.Pod) error {
	k := newKey(pod.Namespace, pod.Name)
	scores, err := json.Marshal(s.result[k].finalscore)
	if err != nil {
		return fmt.Errorf("encode json to record scores: %w", err)
	}

	metav1.SetMetaDataAnnotation(&pod.ObjectMeta, finalScoreResultAnnotationKey, string(scores))
	return nil
}

// AddFilterResult adds filtering result to pod annotation.
func (s *Store) AddFilterResult(namespace, podName, nodeName, pluginName, reason string) {
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

	if _, ok := s.result[k].filter[nodeName]; !ok {
		s.result[k].filter[nodeName] = map[string]string{}
	}

	s.result[k].filter[nodeName][pluginName] = reason
}

// AddScoreResult adds scoring result to pod annotation.
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
		s.result[k].score[nodeName] = map[string]string{}
	}
	if _, ok := s.result[k].finalscore[nodeName]; !ok {
		s.result[k].finalscore[nodeName] = map[string]string{}
	}

	s.result[k].score[nodeName][pluginName] = scoreToString(score)
	s.addFinalScoreResultWithoutLock(namespace, podName, nodeName, pluginName, score)
}

// AddFinalScoreResult adds final score result to pod annotation.
// Final score is calculated with each plugin weight from normalizedScore.
func (s *Store) AddFinalScoreResult(namespace, podName, nodeName, pluginName string, normalizedScore int64) {
	if !strings.HasSuffix(nodeName, namespace) {
		// when suffix of nodeName don't match namespace, this node is not created by a user who creates this pod.
		// So we don't need to record the result in this case.
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.addFinalScoreResultWithoutLock(namespace, podName, nodeName, pluginName, normalizedScore)
}

func (s *Store) addFinalScoreResultWithoutLock(namespace, podName, nodeName, pluginName string, normalizedScore int64) {
	k := newKey(namespace, podName)
	if _, ok := s.result[k]; !ok {
		s.result[k] = newData()
	}

	if _, ok := s.result[k].finalscore[nodeName]; !ok {
		s.result[k].finalscore[nodeName] = map[string]string{}
	}

	// apply weight to calculate final score.
	finalscore := s.applyWeightOnScore(pluginName, normalizedScore)
	s.result[k].finalscore[nodeName][pluginName] = scoreToString(finalscore)
}

func (s *Store) applyWeightOnScore(pluginName string, score int64) int64 {
	weight := s.scorePluginWeight[pluginName]
	return score * int64(weight)
}

// scoreToString convert score(int64) to string.
// It returns DisabledMessage when score is DisabledScore.
func scoreToString(score int64) string {
	if score == DisabledScore {
		return DisabledMessage
	}
	return strconv.FormatInt(score, 10)
}

func (s *Store) DeleteData(k key) {
	delete(s.result, k)
}
