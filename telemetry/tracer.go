package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/loongkirin/gdk/util"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.opentelemetry.io/otel/trace"
)

// InitTracer 初始化 TracerProvider
func InitTracer(ctx context.Context, cfg TelemetryConfig) (trace.TracerProvider, error) {
	collecteTimeout, err := util.ParseDuration(cfg.CollecteTimeout)
	if err != nil {
		collecteTimeout = time.Second * 10
	}

	enableTLS := cfg.TlsConfig.EnableTLS()

	// 创建 OTLP exporter
	var exporter *otlptrace.Exporter
	switch cfg.CollectorType {
	case "grpc":
		grpcOptions := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(cfg.CollectorURL),
			otlptracegrpc.WithTimeout(collecteTimeout),
		}
		if enableTLS {
			credentials, err := LoadTLSCredentials(cfg.TlsConfig)
			if err != nil {
				enableTLS = false
				fmt.Println("failed to load TLS credentials:", err)
			} else {
				grpcOptions = append(grpcOptions, otlptracegrpc.WithTLSCredentials(credentials))
			}
		}
		if !enableTLS {
			grpcOptions = append(grpcOptions, otlptracegrpc.WithInsecure())
		}
		exporter, err = otlptracegrpc.New(ctx, grpcOptions...)
	case "http":
		httpOptions := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(cfg.CollectorURL),
			otlptracehttp.WithTimeout(collecteTimeout),
		}
		if enableTLS {
			tlsConfig, err := NewTLSConfig(cfg.TlsConfig)
			if err != nil {
				enableTLS = false
				fmt.Println("failed to create TLS config:", err)
			} else {
				httpOptions = append(httpOptions, otlptracehttp.WithTLSClientConfig(tlsConfig))
			}
		}
		if !enableTLS {
			httpOptions = append(httpOptions, otlptracehttp.WithInsecure())
		}

		exporter, err = otlptracehttp.New(ctx, httpOptions...)
	default:
		fmt.Println("unsupported tracer exporter type:", cfg.CollectorType)
		return nil, fmt.Errorf("unsupported tracer exporter type: %s", cfg.CollectorType)
	}
	if err != nil {
		fmt.Println("failed to create tracer exporter:", err)
		return nil, err
	}

	// 创建资源属性
	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(cfg.ServiceName),
		semconv.ServiceVersionKey.String(cfg.ServiceVersion),
		semconv.DeploymentEnvironmentNameKey.String(cfg.ServiceEnvironment),
	)

	// 创建 TracerProvider
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.TraceSample)),
	)

	// 设置全局 TracerProvider
	otel.SetTracerProvider(provider)

	// 设置全局 propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	fmt.Println("tracer provider initialized")
	return provider, nil
}

// ShutdownTracer 关闭 TracerProvider
func ShutdownTracer(ctx context.Context) error {
	if provider, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider); ok {
		return provider.Shutdown(ctx)
	}
	return nil
}

func GetTracer(name string) trace.Tracer {
	return otel.GetTracerProvider().Tracer(name)
}
