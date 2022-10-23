package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"

	"github.com/rectcircle/learn-observability/01-opentracing/tracing"
)

type SMSMsgTask struct {
	PhoneNumber        string
	Msg                string
	SpanContextCarrier []byte
}

type SMSHandler struct {
	tracer opentracing.Tracer
	client *tracing.HTTPClient
	queue  chan SMSMsgTask // 模拟一个外部的消息队列
}

func NewSMSHandler(tracer opentracing.Tracer, client *tracing.HTTPClient) *SMSHandler {
	return &SMSHandler{
		tracer: tracer,
		client: client,
		queue:  make(chan SMSMsgTask, 10),
	}
}

func (l *SMSHandler) SendSMSMsg(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := opentracing.SpanFromContext(ctx)
	spanContextCarrier := &bytes.Buffer{}
	l.tracer.Inject(
		span.Context(),
		opentracing.Binary,
		spanContextCarrier)

	phoneNumber := r.URL.Query().Get("PhoneNumber")
	if phoneNumber == "" {
		tracing.RespError(ctx, w, "info", 400, "PhoneNumber not exist.")
		return
	}
	message := r.URL.Query().Get("Message")
	if phoneNumber == "" {
		tracing.RespError(ctx, w, "info", 400, "Message not exist.")
		return
	}

	select {
	case l.queue <- SMSMsgTask{
		PhoneNumber:        phoneNumber,
		Msg:                message,
		SpanContextCarrier: spanContextCarrier.Bytes(),
	}:
		tracing.RespSuccess(w, map[string]string{})
	default:
		tracing.RespError(ctx, w, "info", 500, "Queue has fulled.")
	}
}

func (l *SMSHandler) StartWorker() {
	for t := range l.queue {
		lastSpanContext, err := l.tracer.Extract(
			opentracing.Binary,
			bytes.NewBuffer(t.SpanContextCarrier))
		if err != nil {
			log.Printf("[ERROR] extract span error: %s", err)
		}
		func() {
			span := l.tracer.StartSpan("sms-worker", opentracing.FollowsFrom(lastSpanContext))
			defer span.Finish()
			time.Sleep(time.Duration(rand.Intn(3000)) * time.Millisecond)
			fmt.Printf("=== Send to: %s, Message: %s\n", t.PhoneNumber, t.Msg)
		}()
	}
}
