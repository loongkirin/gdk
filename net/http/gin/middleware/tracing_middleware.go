package middleware

import (
	"bytes"
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
			),
		)
		defer span.End()

		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		// 将 span 上下文传递给后续处理器
		c.Request = c.Request.WithContext(spanCtx)

		// 记录响应体
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		span.AddEvent("http.request.start", trace.WithAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.trace_id", GetTraceID(c)),
			attribute.String("http.request_id", GetRequestId(c)),
		))

		c.Next()

		duration := time.Since(start).Milliseconds()
		span.AddEvent("http.request.end", trace.WithAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.Int64("http.response_size", int64(c.Writer.Size())),
			attribute.String("http.response_body", writer.body.String()),
			attribute.Int64("http.duration", duration),
		))

		// 记录响应状态
		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
		)
	}
}
