package tracing

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

func WrapHTTPHandler(tracer opentracing.Tracer, pattern string, handler http.Handler) http.Handler {
	return nethttp.Middleware(tracer, handler, nethttp.OperationNameFunc(func(r *http.Request) string {
		return "HTTP " + r.Method + " " + pattern
	}))
}

func RespError(ctx context.Context, w http.ResponseWriter, logLevel string, status int, err string) {
	span := opentracing.SpanFromContext(ctx)
	result := map[string]interface{}{"Error": err}
	body, _ := json.Marshal(result)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(body)
	span.LogFields(log.String("level", logLevel), log.Message(err))
}

func RespSuccess(w http.ResponseWriter, body interface{}) {
	bodyBytes, _ := json.Marshal(body)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(bodyBytes)
}
