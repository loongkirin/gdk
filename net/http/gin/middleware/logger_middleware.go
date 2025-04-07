package middleware

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	gdklogger "github.com/loongkirin/gdk/logger"
)

func Logger(logger gdklogger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		body, err := c.GetRawData()
		if err != nil {
			logger.Error("Failed to get gin request body raw data", gdklogger.Fields{"error": err.Error})
			body = []byte{}
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// 记录响应体
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		// 处理请求
		c.Next()

		// 记录请求信息
		duration := float64(time.Since(start).Microseconds()) / 1e3
		// ctxLogger := logger.With().Fields(map[string]interface{}{
		// 	"traceId":   traceId,
		// 	"requestId": requestId,
		// }).Logger()

		logger.Info("HTTP Request", gdklogger.Fields{
			"traceId":          GetTraceID(c),
			"requestId":        GetRequestId(c),
			"method":           c.Request.Method,
			"path":             c.Request.URL.Path,
			"status":           c.Writer.Status(),
			"duration":         fmt.Sprintf("%.3fms", duration),
			"clientIp":         c.ClientIP(),
			"userAgent":        c.Request.UserAgent(),
			"requestSize":      c.Request.ContentLength,
			"requestBody":      string(body),
			"headers":          c.Request.Header,
			"responseBody":     writer.body.String(),
			"requestStartTime": start.Format(time.RFC3339Nano),
			"requestEndTime":   time.Now().Format(time.RFC3339Nano),
		})
	}
}
