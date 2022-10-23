package tracing

import (
	"fmt"

	"github.com/uber/jaeger-client-go/rpcmetrics"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
)

// Init creates a new instance of Jaeger tracer.
func NewTracer(serviceName string) (opentracing.Tracer, error) {
	cfg := &config.Configuration{
		// 服务名
		ServiceName: serviceName,
		// 采样配置
		// 以创建 tracing 的节点的配置有效，如 A -> B -> C，则 A 采样策略生效，B、C 遵循 A 的决定。
		Sampler: &config.SamplerConfig{
			// 更多参见：https://www.jaegertracing.io/docs/1.38/sampling/#client-sampling-configuration
			// const: Param 为 1 表示全部采样（全部上报），为 0 关闭采样（永远不上报），
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			// 将 span 提交日志，上报到外部日志服务
			LogSpans: true,
		},
	}
	_, err := cfg.FromEnv()
	if err != nil {
		return nil, fmt.Errorf("cannot parse Jaeger env vars: %s", err)
	}

	metricsFactory := NewJaegerMetricsFactory()
	tracer, _, err := cfg.NewTracer(
		// 用来记录 Jaeger 自身的一些错误以及 Span 提交（需启用 Reporter.LogSpans），到外部日志服务。
		config.Logger(NewJaegerLogger()),
		// 用来上报 Span 的一些统计指标到外部 Metrics 服务。
		config.Metrics(metricsFactory),
		// 用来观察 Span 创建的事件。
		config.Observer(rpcmetrics.NewObserver(metricsFactory, rpcmetrics.DefaultNameNormalizer)),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize Jaeger Tracer: %s", err)
	}
	return tracer, nil
}
