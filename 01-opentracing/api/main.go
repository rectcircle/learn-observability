package main

// JAEGER_AGENT_HOST=localhost go run ./sms

import (
	"net/http"
	"time"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/rectcircle/learn-observability/01-opentracing/tracing"
)

func main() {
	tracer, err := tracing.NewTracer("api")
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	client := &tracing.HTTPClient{
		Tracer: tracer,
		Client: &http.Client{Transport: &nethttp.Transport{}, Timeout: 30 * time.Second},
	}
	loginHandler := NewLoginHandler(client)
	mux.Handle("/api/v1/SendSMSCode", tracing.WrapHTTPHandler(tracer, "/api/v1/SendSMSCode", http.HandlerFunc(loginHandler.SendSMSCode)))
	http.ListenAndServe(":8080", mux)
}
