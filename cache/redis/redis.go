package redis

import (
	"context"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type RedisClient struct {
	RedisConfig *RedisConfig
	master      *redis.Client
	slaves      []*redis.Client
	lock        sync.RWMutex
	current     int
	tracer      trace.Tracer
	meter       metric.Meter
}

func NewRedisClient(cfg *RedisConfig) (*RedisClient, error) {
	// 创建 master 客户端
	master := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Master.Host, cfg.Master.Port),
		Password: cfg.Master.Password,
		DB:       cfg.Master.DB,
		PoolSize: cfg.PoolSize,
	})

	if err := master.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to master redis: %w", err)
	}

	var meter metric.Meter
	var tracer trace.Tracer
	if cfg.EnableMetrics {
		// 创建 meter 和 tracer
		meter = otel.Meter("redis-client")
		tracer = otel.Tracer("redis-client")
		reportRedisMetrics(master, meter, "master", cfg.Master)
	}

	var slaves []*redis.Client
	for i, slaveCfg := range cfg.Slaves {
		slave := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", slaveCfg.Host, slaveCfg.Port),
			Password: slaveCfg.Password,
			DB:       slaveCfg.DB,
			PoolSize: cfg.PoolSize,
		})

		if err := slave.Ping(context.Background()).Err(); err != nil {
			return nil, fmt.Errorf("failed to connect to slave redis: %w", err)
		}

		if cfg.EnableMetrics {
			reportRedisMetrics(slave, meter, fmt.Sprintf("slave_%d", i), slaveCfg)
		}
		slaves = append(slaves, slave)
	}

	return &RedisClient{
		RedisConfig: cfg,
		master:      master,
		slaves:      slaves,
		tracer:      tracer,
		meter:       meter,
	}, nil
}

// 添加指标报告函数
func reportRedisMetrics(client *redis.Client, meter metric.Meter, dbRole string, dbConntection RedisConnection) {
	// 创建指标
	poolSize, _ := meter.Int64ObservableGauge(
		"redis.pool.size",
		metric.WithDescription("Number of connections in the pool"),
	)
	idleConns, _ := meter.Int64ObservableGauge(
		"redis.pool.idle_connections",
		metric.WithDescription("Number of idle connections"),
	)
	totalConns, _ := meter.Int64ObservableGauge(
		"redis.pool.total_connections",
		metric.WithDescription("Total number of connections"),
	)

	opts := []metric.ObserveOption{
		metric.WithAttributes(
			attribute.String("redis.role", dbRole),
			attribute.String("redis.host", dbConntection.Host),
			attribute.Int("redis.port", dbConntection.Port),
			attribute.Int("redis.db", dbConntection.DB),
		),
	}

	// 注册回调
	meter.RegisterCallback(
		func(ctx context.Context, o metric.Observer) error {
			stats := client.PoolStats()
			o.ObserveInt64(poolSize, int64(stats.TotalConns), opts...)
			o.ObserveInt64(idleConns, int64(stats.IdleConns), opts...)
			o.ObserveInt64(totalConns, int64(stats.TotalConns), opts...)
			return nil
		},
		poolSize,
		idleConns,
		totalConns,
	)
}

// 添加健康检查方法
func (r *RedisClient) HealthCheck() error {
	// 检查 master
	if err := r.master.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("master health check failed: %w", err)
	}

	// 检查 slaves
	for i, slave := range r.slaves {
		if err := slave.Ping(context.Background()).Err(); err != nil {
			return fmt.Errorf("slave_%d health check failed: %w", i, err)
		}
	}
	return nil
}

// 添加关闭方法
func (r *RedisClient) Close() error {
	if err := r.master.Close(); err != nil {
		return fmt.Errorf("failed to close master: %w", err)
	}
	for i, slave := range r.slaves {
		if err := slave.Close(); err != nil {
			return fmt.Errorf("failed to close slave_%d: %w", i, err)
		}
	}
	return nil
}

// 添加带追踪的命令包装器
func (r *RedisClient) WithTrace(ctx context.Context, cmd string) (context.Context, trace.Span) {
	return r.tracer.Start(ctx, "redis."+cmd,
		trace.WithAttributes(
			attribute.String("redis.command", cmd),
		),
	)
}

func (r *RedisClient) GetMasterDb() *redis.Client {
	return r.master
}

func (r *RedisClient) GetSlaveDb() *redis.Client {
	if len(r.slaves) == 0 {
		return r.master
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	r.current = (r.current + 1) % len(r.slaves)
	return r.slaves[r.current]
}
