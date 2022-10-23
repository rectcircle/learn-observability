package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/rectcircle/learn-observability/01-opentracing/tracing"
)

type LoginHandler struct {
	client *tracing.HTTPClient
}

func NewLoginHandler(client *tracing.HTTPClient) *LoginHandler {
	return &LoginHandler{
		client: client,
	}
}

func (l *LoginHandler) SendSMSCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := opentracing.SpanFromContext(ctx)

	phoneNumber := r.URL.Query().Get("PhoneNumber")
	if phoneNumber == "" {
		tracing.RespError(ctx, w, "info", 400, "PhoneNumber not exist.")
		return
	}

	var authResult struct {
		Error   string
		CodeKey string
	}
	err := l.client.GetJSON(ctx, "auth.SendSMSCode", "http://localhost:8081/api/v1/SendSMSCode?PhoneNumber="+url.QueryEscape(phoneNumber), &authResult)
	if err != nil {
		tracing.RespError(ctx, w, "info", 500, "Call auth.SendSMSCode error: "+err.Error())
		return
	}
	if authResult.Error != "" {
		// TODO 这里应该有错误包装逻辑
		tracing.RespError(ctx, w, "info", 400, authResult.Error)
	}
	tracing.RespSuccess(w, map[string]string{
		"CodeKey": authResult.CodeKey,
	})
	span.LogFields(log.String("level", "info"), log.Message(fmt.Sprintf("key=%s", authResult.CodeKey)))
	return
}
