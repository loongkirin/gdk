package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	gdklogger "github.com/loongkirin/gdk/logger"
	"github.com/loongkirin/gdk/util"
)

func Logger(logger gdklogger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		traceId := GetTraceID(c)
		if len(traceId) == 0 {
			traceId = util.GenerateId()
		}
		SetTraceID(c, traceId)

		requestId := GetRequestId(c)
		if len(requestId) == 0 {
			requestId = util.GenerateId()
		}
		SetRequestId(c, requestId)
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
		duration := time.Since(start).Microseconds()
		// ctxLogger := logger.With().Fields(map[string]interface{}{
		// 	"traceId":   traceId,
		// 	"requestId": requestId,
		// }).Logger()

		logger.Info("HTTP Request", gdklogger.Fields{
			"traceId":      traceId,
			"requestId":    requestId,
			"method":       c.Request.Method,
			"path":         c.Request.URL.Path,
			"status":       c.Writer.Status(),
			"duration":     duration,
			"clientIp":     c.ClientIP(),
			"userAgent":    c.Request.UserAgent(),
			"requestSize":  c.Request.ContentLength,
			"requestBody":  string(body),
			"headers":      c.Request.Header,
			"responseBody": writer.body.String(),
		})
	}
}
