package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

func loggingHandler(c *gin.Context) {

	if !ignoreTracingRequest(c.Request) {
		c.Next()
		return
	}

	// Start timer
	start := time.Now()
	path := c.Request.URL.Path
	raw := c.Request.URL.RawQuery

	span := trace.SpanFromContext(c.Request.Context())

	c.Next()

	now := time.Now()
	latency := float64(now.Sub(start).Nanoseconds()) / 1000000.0
	clientIP := c.ClientIP()
	method := c.Request.Method
	statusCode := c.Writer.Status()
	errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()
	userAgent := c.Request.UserAgent()
	proto := c.Request.Proto

	log.WithFields(log.Fields{
		"http": log.Fields{
			"status":     statusCode,
			"client":     clientIP,
			"method":     method,
			"path":       path,
			"query":      raw,
			"user-agent": userAgent,
			"proto":      proto,
			"headers":    headersFromRequest(c.Request),
			"latency":    latency,
			"error":      errorMessage,
		},
	}).WithFields(commonLogFieleds(span)).Info()

}

func commonLogFieleds(span trace.Span) log.Fields {
	return log.Fields{
		"service":  os.Getenv("OTEL_SERVICE_NAME"),
		"version":  os.Getenv("VERSION"),
		"env":      os.Getenv("ENV"),
		"trace_id": span.SpanContext().TraceID().String(),
		"span_id":  span.SpanContext().SpanID().String(),
	}
}

func headersFromRequest(r *http.Request) log.Fields {
	ipHeaders := []string{
		"x-forwarded-for",
		"x-real-ip",
		"x-client-ip",
		"x-forwarded",
		"x-cluster-client-ip",
		"forwarded-for",
		"forwarded",
		"via",
		"true-client-ip",
	}
	var headers []string
	var ips []string
	for _, hdr := range ipHeaders {
		if v := r.Header.Get(hdr); v != "" {
			headers = append(headers, hdr)
			ips = append(ips, v)
		}
	}

	result := log.Fields{}
	for i := range headers {
		result[headers[i]] = ips[i]
	}
	return result
}
