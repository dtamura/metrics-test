package main

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

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

	// create http request
	client := &http.Client{}

	target := os.Getenv("PING_TARGET_URL")
	req, err := http.NewRequest("GET", target+"/greeting", nil)
	if err != nil {
		log.Error(err)
		return ""
	}
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	req.Header.Add("Content-Type", "application/json")

	// Start Request
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		span.End()
		return ""
	}
	defer resp.Body.Close()

	// 一定の割合で意図的な遅延
	if rand.Float64() < 0.2 {
		_, childSpan := tracer.Start(ctx, "sleep")
		time.Sleep(time.Millisecond * time.Duration(rand.Int63n(1000)))
		childSpan.End()
	}
	span.End()

	var data struct {
		Id      int
		Content string
		TraceId string
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Error(err)
		span.RecordError(err)
		span.End()
		return ""
	}

	return data.Content
}
