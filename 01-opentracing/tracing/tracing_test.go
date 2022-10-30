package tracing

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

func TestExtractAndInject(t *testing.T) {
	tracer, err := NewTracer("test")
	if err != nil {
		t.Fatal(err)
	}
	rootSpan := tracer.StartSpan("root")
	rootSpan.SetBaggageItem("BaggageRoot", "456")

	textMapCarrier := opentracing.TextMapCarrier{}
	tracer.Inject(rootSpan.Context(), opentracing.TextMap, textMapCarrier)
	fmt.Printf("=== textMapCarrier: %v\n", textMapCarrier)
	httpHeaderCarrier := opentracing.HTTPHeadersCarrier(http.Header{})
	tracer.Inject(rootSpan.Context(), opentracing.HTTPHeaders, httpHeaderCarrier)
	fmt.Printf("=== httpHeaderCarrier: %v\n", httpHeaderCarrier)
	binaryCarrier := &bytes.Buffer{}
	tracer.Inject(rootSpan.Context(), opentracing.Binary, binaryCarrier)
	fmt.Printf("=== binaryCarrier: %v\n", binaryCarrier)

	root1SpanContext, err := tracer.Extract(opentracing.TextMap, textMapCarrier)
	if err != nil {
		panic(err)
	}
	root2SpanContext, err := tracer.Extract(opentracing.HTTPHeaders, httpHeaderCarrier)
	if err != nil {
		panic(err)
	}
	root3SpanContext, err := tracer.Extract(opentracing.Binary, binaryCarrier)
	if err != nil {
		panic(err)
	}
	child1Span := tracer.StartSpan("child", opentracing.ChildOf(root1SpanContext))
	fmt.Printf("=== child1Span BaggageRoot: %v\n", child1Span.BaggageItem(strings.ToLower("BaggageRoot"))) // 使用 opentracing.TextMap 像是个 bug，不区分大小写。
	child2Span := tracer.StartSpan("child", opentracing.ChildOf(root2SpanContext))
	fmt.Printf("=== child2Span BaggageRoot: %v\n", child2Span.BaggageItem(strings.ToLower("BaggageRoot"))) // 使用 opentracing.HTTPHeaders 像是个 bug，不区分大小写。
	child3Span := tracer.StartSpan("child", opentracing.ChildOf(root3SpanContext))
	fmt.Printf("=== child3Span BaggageRoot: %v\n", child3Span.BaggageItem("BaggageRoot"))
}

func Service2B(tracer2 opentracing.Tracer, httpHeader http.Header) {
	// 准备 Span 一个，一般在中间件中实现，反序列化 SpanContext
	var BSpan opentracing.Span
	tags := opentracing.Tags{"b": 2}
	previousContext, err := tracer2.Extract(opentracing.HTTPHeaders, httpHeader)
	if err == nil {
		BSpan = tracer2.StartSpan("B", tags, opentracing.ChildOf(previousContext))
	} else {
		BSpan = tracer2.StartSpan("B", tags)
	}
	defer BSpan.Finish()

	// 业务逻辑
	BSpan.LogFields(
		log.String("message", "Service2B called"),
		log.String("BaggageA", BSpan.BaggageItem("BaggageA")),
	)
}

func Service1A(tracer1, tracer2 opentracing.Tracer) {
	// 准备 Span 一个，一般在中间件中实现
	ASpan := tracer1.StartSpan("A", opentracing.Tags{"a": 1})
	defer ASpan.Finish()

	// 业务逻辑
	// 在 span 中记录日志
	ASpan.LogFields(log.String("message", "Service1A called"))
	// 设置 Baggage
	ASpan.SetBaggageItem("BaggageA", "123")
	// 模拟调用 Service2 的 B 函数，序列化 SpanContext
	headerCarrier := http.Header{}
	tracer1.Inject(ASpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(headerCarrier))
	Service2B(tracer2, headerCarrier)
	ASpan.LogFields(log.String("message", "call Service2B success"))
}

func TestOpentracing(t *testing.T) {
	tracer1, err := NewTracer("service1")
	if err != nil {
		t.Fatal(err)
	}
	tracer2, err := NewTracer("service2")
	if err != nil {
		t.Fatal(err)
	}

	// 模拟调用 Service1 的 方法 A
	Service1A(tracer1, tracer2)
}
