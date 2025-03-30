package middleware

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/loongkirin/gdk/logger"
	"github.com/loongkirin/gdk/net/http/response"
	"github.com/loongkirin/gdk/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/time/rate"
)

var (
	breakRateLimitExceededDef = telemetry.MetricDefinition[int64]{
		Name:        "http_request_break_rate_limit_exceeded_total",
		Description: "Total number of HTTP requests break rate limit exceeded",
		Unit:        "1",
		Kind:        telemetry.KindCounter,
	}
)

type BreakRateLimiterConfig struct {
	Limit  rate.Limit                     // 限流速率 次/秒
	Burst  int                            // 限流器桶大小
	Meter  *telemetry.DynamicMeter[int64] // 指标
	Logger logger.Logger                  // 日志
}

// RateLimiter 基于 source 的限流器
type IPBreakRateLimiter struct {
	sources map[string]*rate.Limiter
	lock    *sync.RWMutex
	config  BreakRateLimiterConfig
}

// NewRateLimiter 创建一个新的限流器
func NewIPBreakRateLimiter(config BreakRateLimiterConfig) *IPBreakRateLimiter {
	return &IPBreakRateLimiter{
		sources: make(map[string]*rate.Limiter),
		lock:    &sync.RWMutex{},
		config:  config,
	}
}

// 添加 Source 到限流器
func (i *IPBreakRateLimiter) AddSource(source string) *rate.Limiter {
	i.lock.Lock()
	defer i.lock.Unlock()

	limiter := rate.NewLimiter(i.config.Limit, i.config.Burst)
	i.sources[source] = limiter
	return limiter
}

// GetLimiter 获取 IP 对应的限流器
func (i *IPBreakRateLimiter) GetLimiter(source string) *rate.Limiter {
	i.lock.Lock()
	limiter, exists := i.sources[source]

	if !exists {
		i.lock.Unlock()
		return i.AddSource(source)
	}

	i.lock.Unlock()
	return limiter
}

// RateLimitMiddleware 限流中间件
func BreakRateLimiter(limiter *IPBreakRateLimiter) gin.HandlerFunc {
	if limiter.config.Meter != nil {
		if err := initBreakRateLimitExceededMetrics(limiter.config.Meter); err != nil {
			panic(err)
		}
	}

	return func(c *gin.Context) {
		source := fmt.Sprintf("%s:%s:%s", c.ClientIP(), c.Request.Method, c.Request.URL.Path)
		rateLimiter := limiter.GetLimiter(source)
		if !rateLimiter.Allow() {
			if limiter.config.Meter != nil {
				limiter.config.Meter.RecordMetric(c, telemetry.MetricValue[int64]{
					Name:  breakRateLimitExceededDef.Name,
					Value: 1,
					Attributes: []attribute.KeyValue{
						attribute.String("source", source),
					},
				})
			}

			if limiter.config.Logger != nil {
				limiter.config.Logger.Error(fmt.Sprintf("source: %s, rate limit exceeded", source))
			}

			c.JSON(http.StatusTooManyRequests, response.NewResponse(response.ERROR, "Too many requests"))
			c.Abort()
			return
		}
		c.Next()
	}
}

// initBreakRateLimitExceededMetrics 初始化指标
func initBreakRateLimitExceededMetrics(meter *telemetry.DynamicMeter[int64]) error {
	metrics := []telemetry.MetricDefinition[int64]{
		breakRateLimitExceededDef,
	}

	for _, metric := range metrics {
		if _, err := meter.GetOrCreateMetric(metric); err != nil {
			return err
		}
	}
	return nil
}
