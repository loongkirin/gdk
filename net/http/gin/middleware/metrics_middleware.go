package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/loongkirin/gdk/telemetry"
	"go.opentelemetry.io/otel/attribute"
)

var (
	// 定义指标
	httpRequestsInFlightDef = telemetry.MetricDefinition[float64]{
		Name:        "http_requests_in_flight",
		Description: "Number of HTTP requests currently being served",
		Unit:        "1",
		Kind:        telemetry.KindGauge,
	}

	httpRequestDurationDef = telemetry.MetricDefinition[float64]{
		Name:        "http_request_duration_seconds",
		Description: "HTTP request duration in seconds",
		Unit:        "s",
		Kind:        telemetry.KindHistogram,
	}

	httpRequestsTotalDef = telemetry.MetricDefinition[float64]{
		Name:        "http_requests_total",
		Description: "Total number of HTTP requests",
		Unit:        "1",
		Kind:        telemetry.KindCounter,
	}
)

// Metrics 中间件用于收集 HTTP 请求的指标
func Metrics(dynamicMeter *telemetry.DynamicMeter[float64]) gin.HandlerFunc {
	// 初始化指标
	if err := initHttpRequestsMetrics(dynamicMeter); err != nil {
		panic(err)
	}

	return func(c *gin.Context) {
		start := time.Now()
		ctx := c.Request.Context()

		// 记录当前请求数
		dynamicMeter.RecordMetric(ctx, telemetry.MetricValue[float64]{
			Name:  httpRequestsInFlightDef.Name,
			Value: 1,
			Attributes: []attribute.KeyValue{
				attribute.String("method", c.Request.Method),
				attribute.String("path", c.FullPath()),
			},
		})

		// 处理请求
		c.Next()

		// 记录请求持续时间
		duration := time.Since(start).Seconds()
		dynamicMeter.RecordMetric(ctx, telemetry.MetricValue[float64]{
			Name:  httpRequestDurationDef.Name,
			Value: duration,
			Attributes: []attribute.KeyValue{
				attribute.String("method", c.Request.Method),
				attribute.String("path", c.FullPath()),
			},
		})

		// 记录请求总数
		dynamicMeter.RecordMetric(ctx, telemetry.MetricValue[float64]{
			Name:  httpRequestsTotalDef.Name,
			Value: 1,
			Attributes: []attribute.KeyValue{
				attribute.String("method", c.Request.Method),
				attribute.String("path", c.FullPath()),
				attribute.String("status", strconv.Itoa(c.Writer.Status())),
			},
		})

		// 减少当前请求数
		dynamicMeter.RecordMetric(ctx, telemetry.MetricValue[float64]{
			Name:  httpRequestsInFlightDef.Name,
			Value: -1,
			Attributes: []attribute.KeyValue{
				attribute.String("method", c.Request.Method),
				attribute.String("path", c.FullPath()),
			},
		})
	}
}

// initMetrics 初始化所有指标
func initHttpRequestsMetrics(dynamicMeter *telemetry.DynamicMeter[float64]) error {
	metrics := []telemetry.MetricDefinition[float64]{
		httpRequestsInFlightDef,
		httpRequestDurationDef,
		httpRequestsTotalDef,
	}

	for _, def := range metrics {
		if _, err := dynamicMeter.GetOrCreateMetric(def); err != nil {
			return err
		}
	}
	return nil
}
