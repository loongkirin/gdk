package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/loongkirin/gdk/logger"
	"github.com/loongkirin/gdk/net/http/response"
	"github.com/loongkirin/gdk/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/ratelimit"
)

var (
	// 指标定义
	blockRateLimitExceededDef = telemetry.MetricDefinition[int64]{
		Name:        "http_request_block_rate_limit_exceeded_total",
		Description: "Total number of HTTP requests block rate limit exceeded",
		Unit:        "1",
		Kind:        telemetry.KindCounter,
	}
)

// IPBlockRateLimiter 基于 source 的限流器
type IPBlockRateLimiter struct {
	sources map[string]ratelimit.Limiter
	lock    *sync.RWMutex
	config  RateLimiterConfig
}

type RateLimiterConfig struct {
	Limit   int                            // 限流速率 次/秒
	Timeout time.Duration                  // 超时时间
	Meter   *telemetry.DynamicMeter[int64] // 指标
	Logger  logger.Logger                  // 日志
}

// NewRateLimiter 创建一个新的限流器
func NewIPBlockRateLimiter(config RateLimiterConfig) *IPBlockRateLimiter {
	return &IPBlockRateLimiter{
		sources: make(map[string]ratelimit.Limiter, 1000),
		lock:    &sync.RWMutex{},
		config:  config,
	}
}

// GetLimiter 获取 source 对应的限流器
func (i *IPBlockRateLimiter) GetLimiter(source string) ratelimit.Limiter {
	// 先使用读锁检查
	i.lock.RLock()
	limiter, exists := i.sources[source]
	i.lock.RUnlock()

	if exists {
		return limiter
	}

	// 不存在时使用写锁创建
	i.lock.Lock()
	defer i.lock.Unlock()

	// 双重检查
	if limiter, exists = i.sources[source]; exists {
		return limiter
	}

	limiter = ratelimit.New(i.config.Limit, ratelimit.Per(time.Second*1))
	i.sources[source] = limiter
	return limiter
}

// RateLimitMiddleware 限流中间件
func BlockRateLimiter(limiter *IPBlockRateLimiter) gin.HandlerFunc {
	if limiter.config.Meter != nil {
		if err := initBlockRateLimitExceededMetrics(limiter.config.Meter); err != nil {
			panic(err)
		}
	}
	return func(c *gin.Context) {
		source := fmt.Sprintf("%s:%s:%s", c.ClientIP(), c.Request.Method, c.Request.URL.Path)
		rateLimiter := limiter.GetLimiter(source)

		// 添加超时控制
		done := make(chan struct{})
		go func() {
			rateLimiter.Take()
			close(done)
		}()

		select {
		case <-done:
			c.Next()
		case <-time.After(limiter.config.Timeout):
			if limiter.config.Meter != nil {
				limiter.config.Meter.RecordMetric(c, telemetry.MetricValue[int64]{
					Name:  blockRateLimitExceededDef.Name,
					Value: 1,
					Attributes: []attribute.KeyValue{
						attribute.String("source", source),
					},
				})
			}
			if limiter.config.Logger != nil {
				limiter.config.Logger.Error(fmt.Sprintf("source: %s, block rate limit exceeded, timeout: %s", source, limiter.config.Timeout))
			}
			c.AbortWithStatusJSON(http.StatusTooManyRequests, response.NewResponse(response.ERROR, fmt.Sprintf("source: %s, block rate limit exceeded, timeout: %s", source, limiter.config.Timeout)))
		}
	}
}

// initBlockRateLimitExceededMetrics 初始化指标
func initBlockRateLimitExceededMetrics(meter *telemetry.DynamicMeter[int64]) error {
	metrics := []telemetry.MetricDefinition[int64]{
		blockRateLimitExceededDef,
	}
	for _, metric := range metrics {
		_, err := meter.GetOrCreateMetric(metric)
		if err != nil {
			return err
		}
	}
	return nil
}
