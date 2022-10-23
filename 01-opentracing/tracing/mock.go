package tracing

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-lib/metrics"
)

type mockJaegerLogger struct{}

func NewJaegerLogger() jaeger.Logger {
	return &mockJaegerLogger{}
}

// Error implements jaeger.Logger
func (*mockJaegerLogger) Error(msg string) {
	log.Printf("<mock-jaeger-logger> [ERROR] %s", msg)
}

// Infof implements jaeger.Logger
func (*mockJaegerLogger) Infof(msg string, args ...interface{}) {
	log.Printf("<mock-jaeger-logger> [INFO] %s", fmt.Sprintf(msg, args...))
}

type mockJaegerMetricsFactory struct {
	Name string
	Tags map[string]string
}

func NewJaegerMetricsFactory() metrics.Factory {
	return &mockJaegerMetricsFactory{}
}

// Counter implements metrics.Factory
func (f *mockJaegerMetricsFactory) Counter(metric metrics.Options) metrics.Counter {
	return &mockJaegerMetrics{
		FactoryName: f.Name,
		FactoryTags: f.Tags,
		Name:        metric.Name,
		Tags:        metric.Tags,
		Help:        metric.Help,
	}
}

// Gauge implements metrics.Factory
func (f *mockJaegerMetricsFactory) Gauge(metric metrics.Options) metrics.Gauge {
	return &mockJaegerMetrics{
		FactoryName: f.Name,
		FactoryTags: f.Tags,
		Name:        metric.Name,
		Tags:        metric.Tags,
		Help:        metric.Help,
	}
}

// Histogram implements metrics.Factory
func (f *mockJaegerMetricsFactory) Histogram(metric metrics.HistogramOptions) metrics.Histogram {
	return &mockJaegerMetrics{
		FactoryName:      f.Name,
		FactoryTags:      f.Tags,
		Name:             metric.Name,
		Tags:             metric.Tags,
		Help:             metric.Help,
		HistogramBuckets: metric.Buckets,
	}
}

// Namespace implements metrics.Factory
func (*mockJaegerMetricsFactory) Namespace(scope metrics.NSOptions) metrics.Factory {
	return &mockJaegerMetricsFactory{
		Name: scope.Name,
		Tags: scope.Tags,
	}
}

// Timer implements metrics.Factory
func (f *mockJaegerMetricsFactory) Timer(metric metrics.TimerOptions) metrics.Timer {
	return &mockJaegerTimerMetrics{
		FactoryName:  f.Name,
		FactoryTags:  f.Tags,
		Name:         metric.Name,
		Tags:         metric.Tags,
		Help:         metric.Help,
		TimerBuckets: metric.Buckets,
	}
}

type mockJaegerMetrics struct {
	FactoryName      string
	FactoryTags      map[string]string
	Name             string
	Tags             map[string]string
	Help             string
	HistogramBuckets []float64
}

func (m *mockJaegerMetrics) printMetric(typ string, val string) {
	FactoryTagsArr := []string{}
	for k, v := range m.FactoryTags {
		FactoryTagsArr = append(FactoryTagsArr, fmt.Sprintf("%s=%s", k, v))
	}
	TagsArr := []string{}
	for k, v := range m.Tags {
		TagsArr = append(TagsArr, fmt.Sprintf("%s=%s", k, v))
	}
	HistogramBucketsArr := []string{}
	for _, v := range m.HistogramBuckets {
		HistogramBucketsArr = append(HistogramBucketsArr, fmt.Sprint(v))
	}

	log.Printf("<mock-jaeger-metric> %s.%s=%s (FactoryName:%s, FactoryTags:%s, Tags:%s, Help:%s, HistogramBuckets:%s)",
		m.Name, typ, val,
		m.FactoryName,
		strings.Join(FactoryTagsArr, ","),
		strings.Join(TagsArr, ","),
		m.Help,
		strings.Join(HistogramBucketsArr, ","),
	)
}

// Inc implements metrics.Counter
func (m *mockJaegerMetrics) Inc(v int64) {
	m.printMetric("Counter", fmt.Sprint(v))
}

// Update implements metrics.Gauge
func (m *mockJaegerMetrics) Update(v int64) {
	m.printMetric("Gauge", fmt.Sprint(v))
}

// Record implements metrics.Histogram
func (m *mockJaegerMetrics) Record(v float64) {
	m.printMetric("Histogram", fmt.Sprint(v))
}

type mockJaegerTimerMetrics struct {
	FactoryName  string
	FactoryTags  map[string]string
	Name         string
	Tags         map[string]string
	Help         string
	TimerBuckets []time.Duration
}

// Record implements metrics.Timer
func (m *mockJaegerTimerMetrics) Record(v time.Duration) {
	typ := "Timer"
	val := fmt.Sprintf("%fs", v.Seconds())
	FactoryTagsArr := []string{}
	for k, v := range m.FactoryTags {
		FactoryTagsArr = append(FactoryTagsArr, fmt.Sprintf("%s=%s", k, v))
	}
	TagsArr := []string{}
	for k, v := range m.Tags {
		TagsArr = append(TagsArr, fmt.Sprintf("%s=%s", k, v))
	}
	TimerBucketsArr := []string{}
	for _, v := range m.TimerBuckets {
		TimerBucketsArr = append(TimerBucketsArr, fmt.Sprint(v))
	}

	log.Printf("<mock-jaeger-metric> %s.%s=%s (FactoryName:%s, FactoryTags:%s, Tags:%s, Help:%s, TimerBuckets:%s)",
		m.Name, typ, val,
		m.FactoryName,
		strings.Join(FactoryTagsArr, ","),
		strings.Join(TagsArr, ","),
		m.Help,
		strings.Join(TimerBucketsArr, ","),
	)
}
