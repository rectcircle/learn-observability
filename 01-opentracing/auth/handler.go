package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"sync"

	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"

	"github.com/rectcircle/learn-observability/01-opentracing/tracing"
)

type AuthHandler struct {
	client    *tracing.HTTPClient
	codeCache sync.Map // 模拟存储验证码
}

func NewAuthHandler(client *tracing.HTTPClient) *AuthHandler {
	return &AuthHandler{
		client: client,
	}
}

func (l *AuthHandler) SendSMSCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := opentracing.SpanFromContext(ctx)

	phoneNumber := r.URL.Query().Get("PhoneNumber")
	if phoneNumber == "" {
		tracing.RespError(ctx, w, "info", 400, "PhoneNumber not exist.")
		return
	}

	var (
		smsResult struct {
			Error string
		}
		codeKey string
		code    string
	)
	codeKey = uuid.NewString()
	code = fmt.Sprintf("%04d", rand.Intn(10000))

	span.LogFields(log.String("level", "info"), log.Message(fmt.Sprintf("key=%s, code=%s", codeKey, code)))

	msg := "[XXXX] Your code is: " + code
	err := l.client.GetJSON(ctx, "sms.SendMsg",
		fmt.Sprintf("http://localhost:8082/api/v1/SendSMSMsg?PhoneNumber=%s&Message=%s", url.QueryEscape(phoneNumber), url.QueryEscape(msg)), &smsResult)
	if err != nil {
		tracing.RespError(ctx, w, "info", 500, "Call sms.SendMsg error: "+err.Error())
		return
	}
	l.codeCache.Store(codeKey, code)
	tracing.RespSuccess(w, map[string]string{
		"CodeKey": codeKey,
	})
}
