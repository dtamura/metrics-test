package main

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

var apiCounter metric.Int64Counter

func init() {
	var err error
	apiCounter, err = meter.Int64Counter(
		"dtamura.api.counter",
		metric.WithDescription("Number of API calls."),
		metric.WithUnit("{call}"),
	)
	if err != nil {
		panic(err)
	}
}

func pingHandler(w http.ResponseWriter, r *http.Request) {

	span := trace.SpanFromContext(r.Context())

	msg := ping(r.Context())
	log.WithFields(commonLogFieleds(span)).Info(msg)
	span.SetAttributes(attribute.String("pong", msg))

	// 一定の割合でエラーを返却
	if rand.Float64() < 0.05 {
		span.RecordError(errors.New("エラー"))
		span.SetStatus(codes.Error, "エラー")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"msg": "error"})
		return
	}

	span.SetStatus(codes.Ok, "OK")
	// Response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"msg": msg, "traceId": span.SpanContext().TraceID().String()})
}

func ping(ctx context.Context) string {
	ctx, span := tracer.Start(ctx, "pong", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	apiCounter.Add(ctx, 1)

	// 一定の割合で意図的な遅延
	if rand.Float64() < 0.2 {
		_, childSpan := tracer.Start(ctx, "sleep")
		time.Sleep(time.Millisecond * time.Duration(rand.Int63n(1000)))
		childSpan.End()
	}
	span.End()

	return "Hello World"
}
