package main

// JAEGER_AGENT_HOST=localhost go run ./auth

import (
	"net/http"
	"time"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/rectcircle/learn-observability/01-opentracing/tracing"
)

func main() {
	tracer, err := tracing.NewTracer("sms")
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	client := &tracing.HTTPClient{
		Tracer: tracer,
		Client: &http.Client{Transport: &nethttp.Transport{}, Timeout: 30 * time.Second},
	}
	smsHandler := NewSMSHandler(tracer, client)
	go smsHandler.StartWorker()
	mux.Handle("/api/v1/SendSMSMsg", tracing.WrapHTTPHandler(tracer, "/api/v1/SendSMSMsg", http.HandlerFunc(smsHandler.SendSMSMsg)))
	http.ListenAndServe(":8082", mux)
}
