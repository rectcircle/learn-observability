package main

// 本例仅用来展示 Prometheus Go SDK 的用法，不可用于生产。
// 1. http middleware 可以直接使用 github.com/prometheus/client_golang/prometheus/promhttp 包。
// 2. go runtime 可以直接使用 github.com/prometheus/client_golang/prometheus/collectors 包。

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
)

type HTTPMetricsMiddleware struct {
	reg                    *prometheus.Registry
	httpRequestsTotal      *prometheus.CounterVec
	goMemstatsAllocBytes   prometheus.Gauge
	httpDurations          *prometheus.SummaryVec
	httpDurationsHistogram *prometheus.HistogramVec
}

func NewHTTPMetrics(reg *prometheus.Registry, normMean, normDomain float64) *HTTPMetricsMiddleware {
	// 一些进程粒度的标签，比如 pod name 之类的，这里使用 pid 模拟。
	ConstLabels := map[string]string{
		"pid": fmt.Sprint(os.Getpid()),
	}
	httpLabelNames := []string{"handler", "method", "status_code"}
	m := &HTTPMetricsMiddleware{
		reg: reg,
		// 创建一个 Counter 类型的指标：每个请求会增加 1。
		// 下文， SummaryVec 或者 httpDurationsHistogram 会自动上报该指标，这里仅做演示。
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "http_requests_total",
				Help:        "HTTP request total.",
				ConstLabels: ConstLabels,
			},
			httpLabelNames,
		),
		// 创建一个 Gauge 类型的指标：统计当前时刻的 go runtime memstats alloc。
		// 下文， SummaryVec 或者 httpDurationsHistogram 会自动上报该指标，这里仅做演示。
		goMemstatsAllocBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "go_memstats_alloc_bytes",
			Help:        "HTTP request total.",
			ConstLabels: ConstLabels,
		}),
		// 创建一个 SummaryVec 类型的指标：按照 handler 标签，计算请求耗时的 50% 90% 99% 分位数。
		httpDurations: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:        "http_durations_seconds",
				Help:        "HTTP latency distributions.",
				Objectives:  map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
				ConstLabels: ConstLabels,
			},
			httpLabelNames,
		),
		// 和上面的 httpDurations 类似，但是类型为 Histogram
		// Histogram 分为 20 个桶，桶的划分为：
		//   * 区间 [normMean-5*normDomain, normMean+0.5*normDomain]
		//   * 步长为 0.5*normDomain
		// 举个例子，当 normMean = 1, normDomain = 0.2 时，桶划分为： {0, 0.1, 0.2, ..., 1, ..., 1.8, 1.9}
		httpDurationsHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:                        "http_durations_histogram_seconds",
				Help:                        "HTTP latency distributions.",
				Buckets:                     prometheus.LinearBuckets(normMean-5*normDomain, .5*normDomain, 20),
				NativeHistogramBucketFactor: 1.1,
				ConstLabels:                 ConstLabels,
			},
			httpLabelNames,
		),
	}
	reg.MustRegister(m.httpRequestsTotal)
	reg.MustRegister(m.goMemstatsAllocBytes)
	reg.MustRegister(m.httpDurations)
	reg.MustRegister(m.httpDurationsHistogram)
	return m
}

func (m *HTTPMetricsMiddleware) MetricsHandler() http.Handler {
	return promhttp.HandlerFor(m.reg, promhttp.HandlerOpts{
		// Opt into OpenMetrics to support exemplars.
		EnableOpenMetrics: true,
		// Pass custom registry
		Registry: m.reg,
	})
}

func (m *HTTPMetricsMiddleware) WrapHandler(handlerName string, handler http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		ww := &metricsHTTPResponseWrapper{
			ResponseWriter: w,
			statusCode:     0,
		}
		handler.ServeHTTP(ww, r)
		duration := float64(time.Since(startTime)) / float64(time.Second)
		statusCode := fmt.Sprint(ww.statusCode)
		m.httpRequestsTotal.WithLabelValues(handlerName, r.Method, statusCode).Add(1)
		m.httpDurations.WithLabelValues(handlerName, r.Method, statusCode).Observe(duration)
		m.httpDurationsHistogram.WithLabelValues(handlerName, r.Method, statusCode).Observe(duration)
	})
}

func (m *HTTPMetricsMiddleware) StartBackgroundReportGoCollector(interval time.Duration) {
	// 这只是例子，想要统计 go runtime 相关的指标，可以直接使用 go collector，参见：https://github.com/prometheus/client_golang/blob/main/examples/gocollector/main.go
	// https://gist.github.com/j33ty/79e8b736141be19687f565ea4c6f4226
	go func() {
		for {
			var stat runtime.MemStats
			runtime.ReadMemStats(&stat)
			m.goMemstatsAllocBytes.Set(float64(stat.Alloc))
			time.Sleep(interval)
		}
	}()
}

func (m *HTTPMetricsMiddleware) StartMetricsPush(interval time.Duration) {
	go func() {
		for {
			push.New("http://localhost:9091", "demo_by_pushgateway").Gatherer(m.reg).Push()
			time.Sleep(interval)
		}
	}()
}

type metricsHTTPResponseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *metricsHTTPResponseWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
