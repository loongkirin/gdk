package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/loongkirin/gdk/logger"
	"github.com/loongkirin/gdk/net/http/response"
	"github.com/loongkirin/gdk/telemetry"
	breaker "github.com/sony/gobreaker/v2"
	"go.opentelemetry.io/otel/attribute"
)

var (
	// 定义指标
	circuitBreakerStateDef = telemetry.MetricDefinition[float64]{
		Name:        "http_request_circuit_breaker_state",
		Description: "Number of HTTP requests circuit breaker state",
		Unit:        "1",
		Kind:        telemetry.KindGauge,
	}

	circuitBreakerFailuresDef = telemetry.MetricDefinition[float64]{
		Name:        "http_request_circuit_breaker_failures_total",
		Description: "Total number of HTTP requests circuit breaker failures",
		Unit:        "1",
		Kind:        telemetry.KindCounter,
	}
)

// CircuitBreakerConfig 断路器配置
type CircuitBreakerConfig struct {
	Name         string
	MaxRequests  uint32                           // 半开状态下允许的最大请求数
	Interval     time.Duration                    // 统计时间窗口
	Timeout      time.Duration                    // 断路器打开后，多久后尝试半开
	FailureRatio float64                          // 触发断路器的失败率阈值
	MinRequests  uint32                           // 最小请求数阈值
	Meter        *telemetry.DynamicMeter[float64] // 指标
	Logger       logger.Logger                    // 日志ß
}

// DefaultCircuitBreakerConfig 默认配置
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		Name:         "default",
		MaxRequests:  100,
		Interval:     10 * time.Second,
		Timeout:      60 * time.Second,
		FailureRatio: 0.6,
		MinRequests:  10,
		Meter:        nil,
		Logger:       nil,
	}
}

// CircuitBreakerMiddleware 创建断路器中间件
func CircuitBreaker(config *CircuitBreakerConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}

	if config.Meter != nil {
		if err := initCircuitBreakerMetrics(config.Meter); err != nil {
			panic(err)
		}
	}

	cb := breaker.NewCircuitBreaker[interface{}](breaker.Settings{
		Name:        config.Name,
		MaxRequests: config.MaxRequests,
		Interval:    config.Interval,
		Timeout:     config.Timeout,
		ReadyToTrip: func(counts breaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= config.MinRequests && failureRatio >= config.FailureRatio
		},
		OnStateChange: func(name string, from breaker.State, to breaker.State) {
			// 更新 Prometheus 指标
			state := float64(0) // Closed
			if to == breaker.StateHalfOpen {
				state = 1
			} else if to == breaker.StateOpen {
				state = 2
			}
			if config.Meter != nil {
				config.Meter.RecordMetric(context.Background(), telemetry.MetricValue[float64]{
					Name:  circuitBreakerStateDef.Name,
					Value: state,
					Attributes: []attribute.KeyValue{
						attribute.String("name", name),
					},
				})
			}

			if config.Logger != nil {
				config.Logger.Info(fmt.Sprintf("Circuit Breaker '%s' state changed from %s to %s", name, from, to))
			}
		},
	})

	return func(c *gin.Context) {
		result, err := cb.Execute(func() (interface{}, error) {
			ch := make(chan struct {
				err error
			}, 1)

			go func() {
				c.Next()
				var err error
				if len(c.Errors) > 0 {
					err = c.Errors.Last()
					// 记录失败次数
					if config.Meter != nil {
						config.Meter.RecordMetric(c, telemetry.MetricValue[float64]{
							Name:  circuitBreakerFailuresDef.Name,
							Value: 1,
							Attributes: []attribute.KeyValue{
								attribute.String("name", config.Name),
							},
						})
					}

					if config.Logger != nil {
						config.Logger.Error(fmt.Sprintf("Circuit Breaker '%s' failed, error: %s", config.Name, err))
					}
				}
				ch <- struct{ err error }{err: err}
			}()

			select {
			case result := <-ch:
				return nil, result.err
			case <-time.After(30 * time.Second):
				if config.Meter != nil {
					config.Meter.RecordMetric(c, telemetry.MetricValue[float64]{
						Name:  circuitBreakerFailuresDef.Name,
						Value: 1,
						Attributes: []attribute.KeyValue{
							attribute.String("name", config.Name),
						},
					})
				}

				if config.Logger != nil {
					config.Logger.Error(fmt.Sprintf("Circuit Breaker '%s' timeout", config.Name))
				}
				return nil, fmt.Errorf("request timeout")
			}
		})

		if err != nil {
			if err == breaker.ErrOpenState {
				c.AbortWithStatusJSON(http.StatusServiceUnavailable, response.NewResponse(response.ERROR, "Service is unavailable"))
				return
			}

			c.AbortWithStatusJSON(http.StatusInternalServerError, response.NewResponse(response.ERROR, err.Error()))
			return
		}

		if c.Writer.Written() {
			return
		}

		if result != nil {
			c.JSON(http.StatusOK, result)
		}
	}
}

// CircuitBreakerByPath 为不同路径创建不同的断路器
func CircuitBreakerByPath(configs map[string]*CircuitBreakerConfig) gin.HandlerFunc {
	breakers := make(map[string]*breaker.CircuitBreaker[interface{}])
	defaultConfig := DefaultCircuitBreakerConfig()

	return func(c *gin.Context) {
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		config := defaultConfig
		if pathConfig, ok := configs[path]; ok {
			config = pathConfig
		}

		if config.Meter != nil {
			if err := initCircuitBreakerMetrics(config.Meter); err != nil {
				panic(err)
			}
		}

		cb, exists := breakers[path]
		if !exists {
			cb = breaker.NewCircuitBreaker[interface{}](breaker.Settings{
				Name:        path,
				MaxRequests: config.MaxRequests,
				Interval:    config.Interval,
				Timeout:     config.Timeout,
				ReadyToTrip: func(counts breaker.Counts) bool {
					failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
					return counts.Requests >= config.MinRequests && failureRatio >= config.FailureRatio
				},
				OnStateChange: func(name string, from breaker.State, to breaker.State) {
					state := float64(0)
					if to == breaker.StateHalfOpen {
						state = 1
					} else if to == breaker.StateOpen {
						state = 2
					}
					if config.Meter != nil {
						config.Meter.RecordMetric(context.Background(), telemetry.MetricValue[float64]{
							Name:  circuitBreakerStateDef.Name,
							Value: state,
							Attributes: []attribute.KeyValue{
								attribute.String("name", name),
							},
						})
					}

					fmt.Printf("Circuit Breaker '%s' state changed from %s to %s\n", name, from, to)
				},
			})
			breakers[path] = cb
		}

		result, err := cb.Execute(func() (interface{}, error) {
			ch := make(chan struct {
				err error
			}, 1)

			go func() {
				c.Next()
				var err error
				if len(c.Errors) > 0 {
					err = c.Errors.Last()
					if config.Meter != nil {
						config.Meter.RecordMetric(c, telemetry.MetricValue[float64]{
							Name:  circuitBreakerFailuresDef.Name,
							Value: 1,
							Attributes: []attribute.KeyValue{
								attribute.String("name", path),
							},
						})
					}

					if config.Logger != nil {
						config.Logger.Error(fmt.Sprintf("Circuit Breaker '%s' failed, error: %s", path, err))
					}
				}
				ch <- struct{ err error }{err: err}
			}()

			select {
			case result := <-ch:
				return nil, result.err
			case <-time.After(30 * time.Second):
				if config.Meter != nil {
					config.Meter.RecordMetric(c, telemetry.MetricValue[float64]{
						Name:  circuitBreakerFailuresDef.Name,
						Value: 1,
						Attributes: []attribute.KeyValue{
							attribute.String("name", path),
						},
					})
				}

				if config.Logger != nil {
					config.Logger.Error(fmt.Sprintf("Circuit Breaker '%s' timeout", path))
				}
				return nil, fmt.Errorf("request timeout")
			}
		})

		if err != nil {
			if err == breaker.ErrOpenState {
				c.AbortWithStatusJSON(http.StatusServiceUnavailable, response.NewResponse(response.ERROR, "Service is unavailable"))
				return
			}

			c.AbortWithStatusJSON(http.StatusInternalServerError, response.NewResponse(response.ERROR, err.Error()))
			return
		}

		if c.Writer.Written() {
			return
		}

		if result != nil {
			response.Ok(c, "success", result)
		}
	}
}

// initMetrics 初始化所有指标
func initCircuitBreakerMetrics(dynamicMeter *telemetry.DynamicMeter[float64]) error {
	metrics := []telemetry.MetricDefinition[float64]{
		circuitBreakerStateDef,
		circuitBreakerFailuresDef,
	}

	for _, def := range metrics {
		if _, err := dynamicMeter.GetOrCreateMetric(def); err != nil {
			return err
		}
	}
	return nil
}
