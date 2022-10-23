package tracing

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/opentracing/opentracing-go"
)

type TextMapCarrier struct {
	m map[string]string
}

func (c *TextMapCarrier) Set(k, v string) {
	c.m[k] = v
}

func (c *TextMapCarrier) ToMap() map[string]string {
	return c.m
}

func TestOpentracing(t *testing.T) {
	tracer, err := NewTracer("testing")
	if err != nil {
		t.Fatal(err)
	}
	rootSpan := tracer.StartSpan("root")

	headerCarrier := http.Header{}
	tracer.Inject(rootSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(headerCarrier))
	fmt.Printf("header: %v\n", headerCarrier)

	textMapCarrier := &TextMapCarrier{m: map[string]string{}}
	tracer.Inject(rootSpan.Context(), opentracing.TextMap, textMapCarrier)
	fmt.Printf("map: %v\n", textMapCarrier.ToMap())
}
