package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/loongkirin/gdk/util"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

// InitMetrics 初始化指标收集
func InitMetrics(ctx context.Context, cfg TelemetryConfig) (metric.MeterProvider, error) {
	// 创建资源
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(cfg.ServiceVersion),
			semconv.DeploymentEnvironmentNameKey.String(cfg.ServiceEnvironment),
		),
	)
	if err != nil {
		fmt.Println("failed to create metrics resource:", err)
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	collecteInterval, err := util.ParseDuration(cfg.CollecteInterval)
	if err != nil {
		collecteInterval = time.Second * 1
	}

	collecteTimeout, err := util.ParseDuration(cfg.CollecteTimeout)
	if err != nil {
		collecteTimeout = time.Second * 10
	}

	enableTLS := cfg.TlsConfig.EnableTLS()

	// 创建指标导出器
	var exporter sdkmetric.Exporter
	switch cfg.CollectorType {
	case "grpc":
		grpcOptions := []otlpmetricgrpc.Option{
			otlpmetricgrpc.WithEndpoint(cfg.CollectorURL),
			otlpmetricgrpc.WithTimeout(collecteTimeout),
		}
		if enableTLS {
			credentials, err := LoadTLSCredentials(cfg.TlsConfig)
			if err != nil {
				enableTLS = false
				fmt.Println("failed to load TLS credentials:", err)
			} else {
				grpcOptions = append(grpcOptions, otlpmetricgrpc.WithTLSCredentials(credentials))
			}
		}
		if !enableTLS {
			grpcOptions = append(grpcOptions, otlpmetricgrpc.WithInsecure())
		}
		exporter, err = otlpmetricgrpc.New(ctx, grpcOptions...)
	case "http":
		httpOptions := []otlpmetrichttp.Option{
			otlpmetrichttp.WithEndpoint(cfg.CollectorURL),
			otlpmetrichttp.WithTimeout(collecteTimeout),
		}
		if enableTLS {
			tlsConfig, err := NewTLSConfig(cfg.TlsConfig)
			if err != nil {
				enableTLS = false
				fmt.Println("failed to create TLS config:", err)
			} else {
				httpOptions = append(httpOptions, otlpmetrichttp.WithTLSClientConfig(tlsConfig))
			}
		}
		if !enableTLS {
			httpOptions = append(httpOptions, otlpmetrichttp.WithInsecure())
		}
		exporter, err = otlpmetrichttp.New(ctx, httpOptions...)
	default:
		fmt.Println("unsupported metrics exporter type:", cfg.CollectorType)
		return nil, fmt.Errorf("unsupported metrics exporter type: %s", cfg.CollectorType)
	}
	if err != nil {
		fmt.Println("failed to create metrics exporter:", err)
		return nil, fmt.Errorf("failed to create metrics exporter: %w", err)
	}

	// 创建指标处理器
	processor := sdkmetric.NewPeriodicReader(exporter,
		sdkmetric.WithInterval(collecteInterval),
	)

	// 创建指标提供者
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(processor),
	)

	// 设置全局指标提供者
	otel.SetMeterProvider(provider)

	fmt.Println("metrics provider initialized")
	return provider, nil
}

// ShutdownMetrics 关闭指标收集
func ShutdownMetrics(ctx context.Context) error {
	if provider, ok := otel.GetMeterProvider().(*sdkmetric.MeterProvider); ok {
		return provider.Shutdown(ctx)
	}
	return nil
}

// GetMeter 获取指标
func GetMeter(name string) metric.Meter {
	return otel.GetMeterProvider().Meter(name)
}
