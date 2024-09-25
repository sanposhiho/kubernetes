/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	"time"

	"k8s.io/component-base/metrics"
)

// MetricRecorder represents a metric recorder which takes action when the
// metric Inc(), Dec() and Clear()
type MetricRecorder interface {
	Inc()
	Dec()
	Clear()
}

var _ MetricRecorder = &PendingPodsRecorder{}

// PendingPodsRecorder is an implementation of MetricRecorder
type PendingPodsRecorder struct {
	recorder metrics.GaugeMetric
}

// NewActivePodsRecorder returns ActivePods in a Prometheus metric fashion
func NewActivePodsRecorder() *PendingPodsRecorder {
	return &PendingPodsRecorder{
		recorder: ActivePods(),
	}
}

// NewUnschedulablePodsRecorder returns UnschedulablePods in a Prometheus metric fashion
func NewUnschedulablePodsRecorder() *PendingPodsRecorder {
	return &PendingPodsRecorder{
		recorder: UnschedulablePods(),
	}
}

// NewBackoffPodsRecorder returns BackoffPods in a Prometheus metric fashion
func NewBackoffPodsRecorder() *PendingPodsRecorder {
	return &PendingPodsRecorder{
		recorder: BackoffPods(),
	}
}

// NewGatedPodsRecorder returns GatedPods in a Prometheus metric fashion
func NewGatedPodsRecorder() *PendingPodsRecorder {
	return &PendingPodsRecorder{
		recorder: GatedPods(),
	}
}

// Inc increases a metric counter by 1, in an atomic way
func (r *PendingPodsRecorder) Inc() {
	r.recorder.Inc()
}

// Dec decreases a metric counter by 1, in an atomic way
func (r *PendingPodsRecorder) Dec() {
	r.recorder.Dec()
}

// Clear set a metric counter to 0, in an atomic way
func (r *PendingPodsRecorder) Clear() {
	r.recorder.Set(float64(0))
}

// histogramVecMetric is the data structure passed in the buffer channel between the main framework thread
// and the metricsRecorder goroutine.
type histogramVecMetric struct {
	metric      *metrics.HistogramVec
	labelValues []string
	value       float64
}

type counterVecMetric struct {
	metric      *metrics.CounterVec
	labelValues []string
	valueToAdd  float64
}

type gaugeVecMetric struct {
	metric      *metrics.GaugeVec
	labelValues []string
	valueToAdd  float64
}

type gaugeVecMetricKey struct {
	metricName string
	labelValue string
}

// MetricAsyncRecorder records metric in a separate goroutine to avoid overhead in the critical path.
type MetricAsyncRecorder struct {
	// bufferCh is a channel that serves as a metrics buffer before the metricsRecorder goroutine reports it.
	bufferCh chan any // *histogramVecMetric, *counterVecMetric, or *gaugeVecMetric
	// if bufferSize is reached, incoming metrics will be discarded.
	bufferSize int
	// how often the recorder runs to flush the metrics.
	interval time.Duration

	// aggregatedInflightEventMetric is only to record InFlightEvents metric asynchronously.
	// It's a map from gaugeVecMetricKey to the aggregated value
	// and the aggregated value is flushed to Prometheus every time the interval is reached.
	// Note that we don't lock the map deliberately because we assume the queue takes lock before updating the in-flight events.
	aggregatedInflightEventMetric              map[gaugeVecMetricKey]int
	aggregatedInflightEventMetricLastFlushTime time.Time

	// stopCh is used to stop the goroutine which periodically flushes metrics.
	stopCh <-chan struct{}
	// IsStoppedCh indicates whether the goroutine is stopped. It's used in tests only to make sure
	// the metric flushing goroutine is stopped so that tests can collect metrics for verification.
	IsStoppedCh chan struct{}
}

func NewMetricsAsyncRecorder(bufferSize int, interval time.Duration, stopCh <-chan struct{}) *MetricAsyncRecorder {
	recorder := &MetricAsyncRecorder{
		bufferCh:                      make(chan any, bufferSize),
		bufferSize:                    bufferSize,
		interval:                      interval,
		stopCh:                        stopCh,
		aggregatedInflightEventMetric: make(map[gaugeVecMetricKey]int),
		aggregatedInflightEventMetricLastFlushTime: time.Now(),
		IsStoppedCh: make(chan struct{}),
	}
	go recorder.run()
	return recorder
}

// ObservePluginDurationAsync observes the plugin_execution_duration_seconds metric.
// The metric will be flushed to Prometheus asynchronously.
func (r *MetricAsyncRecorder) ObservePluginDurationAsync(extensionPoint, pluginName, status string, value float64) {
	r.observeHistogramMetricAsync(PluginExecutionDuration, value, pluginName, extensionPoint, status)
}

// ObserveQueueingHintExecution observes the queueing_hint_execution_duration_seconds metric and the queueing_hint_evaluation_total metric.
// The metric will be flushed to Prometheus asynchronously.
func (r *MetricAsyncRecorder) ObserveQueueingHintExecution(pluginName, event, hint string, value float64) {
	r.observeHistogramMetricAsync(queueingHintExecutionDuration, value, pluginName, event, hint)
	r.addCounterMetricAsync(queueingHintEvaluationTotal, 1, pluginName, event, hint)
}

// ObserveInFlightEventsAsync observes the in_flight_events metric.
//
// Note that this function is not goroutine-safe;
// we don't lock the map deliberately for the performance reason and we assume the queue (i.e., the caller) takes lock before updating the in-flight events.
func (r *MetricAsyncRecorder) ObserveInFlightEventsAsync(eventLabel string, valueToAdd float64, forceFlush bool) {
	r.aggregatedInflightEventMetric[gaugeVecMetricKey{metricName: InFlightEvents.Name, labelValue: eventLabel}] += int(valueToAdd)

	// Only flush the metric to the channel if the interval is reached.
	// The values are flushed to Prometheus in the run() function, which runs once the interval time.
	// Note: we implement this flushing here, not in FlushMetrics, because, if we did so, we would need to implement a lock for the map, which we want to avoid.
	if forceFlush || time.Since(r.aggregatedInflightEventMetricLastFlushTime) > r.interval {
		for key, value := range r.aggregatedInflightEventMetric {
			newMetric := &gaugeVecMetric{
				metric:      InFlightEvents,
				labelValues: []string{key.labelValue},
				valueToAdd:  float64(value),
			}
			select {
			case r.bufferCh <- newMetric:
			default:
			}
		}
		r.aggregatedInflightEventMetricLastFlushTime = time.Now()
		// reset
		r.aggregatedInflightEventMetric = make(map[gaugeVecMetricKey]int)
	}
}

func (r *MetricAsyncRecorder) observeHistogramMetricAsync(m *metrics.HistogramVec, value float64, labelsValues ...string) {
	newMetric := &histogramVecMetric{
		metric:      m,
		labelValues: labelsValues,
		value:       value,
	}
	select {
	case r.bufferCh <- newMetric:
	default:
	}
}

func (r *MetricAsyncRecorder) addCounterMetricAsync(m *metrics.CounterVec, valueToAdd float64, labelsValues ...string) {
	newMetric := &counterVecMetric{
		metric:      m,
		labelValues: labelsValues,
		valueToAdd:  valueToAdd,
	}
	select {
	case r.bufferCh <- newMetric:
	default:
	}
}

// run flushes buffered metrics into Prometheus every second.
func (r *MetricAsyncRecorder) run() {
	for {
		select {
		case <-r.stopCh:
			close(r.IsStoppedCh)
			return
		default:
		}
		r.FlushMetrics()
		time.Sleep(r.interval)
	}
}

// FlushMetrics tries to clean up the bufferCh by reading at most bufferSize metrics.
func (r *MetricAsyncRecorder) FlushMetrics() {
	for i := 0; i < r.bufferSize; i++ {
		select {
		case m := <-r.bufferCh:
			switch received := m.(type) {
			case *histogramVecMetric:
				received.metric.WithLabelValues(received.labelValues...).Observe(received.value)
			case *gaugeVecMetric:
				received.metric.WithLabelValues(received.labelValues...).Add(received.valueToAdd)
			case *counterVecMetric:
				received.metric.WithLabelValues(received.labelValues...).Add(received.valueToAdd)
			}
		default:
			// no more value
		}
	}
}
