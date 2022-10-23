package main

// JAEGER_AGENT_HOST=localhost go run ./auth

import (
	"net/http"
	"time"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/rectcircle/learn-observability/01-opentracing/tracing"
)

func main() {
	tracer, err := tracing.NewTracer("auth")
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	client := &tracing.HTTPClient{
		Tracer: tracer,
		Client: &http.Client{Transport: &nethttp.Transport{}, Timeout: 30 * time.Second},
	}
	loginHandler := NewAuthHandler(client)
	mux.Handle("/api/v1/SendSMSCode", tracing.WrapHTTPHandler(tracer, "/api/v1/SendSMSCode", http.HandlerFunc(loginHandler.SendSMSCode)))
	http.ListenAndServe(":8081", mux)
}
