package middleware

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func Tracing(tracer trace.Tracer) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		ctx := c.Request.Context()
		var body []byte
		var err error
		if c.Request.Body != nil {
			body, err = io.ReadAll(c.Request.Body)
			if err != nil {
				body = []byte{}
			}
		}

		spanCtx, span := tracer.Start(ctx, c.Request.URL.Path,
			trace.WithAttributes(
				attribute.String("http.method", c.Request.Method),
				attribute.String("http.url", c.Request.URL.String()),
				attribute.String("http.trace_id", GetTraceID(c)),
				attribute.String("http.request_id", GetRequestId(c)),
				attribute.Int64("http.request_size", c.Request.ContentLength),
				attribute.String("http.client_ip", c.ClientIP()),
				attribute.String("http.user_agent", c.Request.UserAgent()),
				attribute.String("http.request_body", string(body)),
				attribute.String("http.request_start_time", start.Format(time.RFC3339Nano)),
			),
		)
		defer span.End()

		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		// 将 span 上下文传递给后续处理器
		c.Request = c.Request.WithContext(spanCtx)

		span.AddEvent("http.request_start", trace.WithAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.trace_id", GetTraceID(c)),
			attribute.String("http.request_id", GetRequestId(c)),
			attribute.Int64("http.request_size", c.Request.ContentLength),
			attribute.String("http.client_ip", c.ClientIP()),
			attribute.String("http.user_agent", c.Request.UserAgent()),
			attribute.String("http.request_body", string(body)),
		))

		// 记录响应体
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		c.Next()

		duration := float64(time.Since(start).Microseconds()) / 1e3
		span.AddEvent("http.request_end", trace.WithAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.Int64("http.response_size", int64(c.Writer.Size())),
			attribute.String("http.response_body", writer.body.String()),
			attribute.String("http.duration", fmt.Sprintf("%.3fms", duration)),
		))

		// 记录响应状态
		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.Int64("http.response_size", int64(c.Writer.Size())),
			attribute.String("http.response_body", writer.body.String()),
			attribute.String("http.duration", fmt.Sprintf("%.3fms", duration)),
			attribute.String("http.request_end_time", time.Now().Format(time.RFC3339Nano)),
		)
	}
}
