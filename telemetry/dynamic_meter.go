package telemetry

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// MetricKind 定义指标种类
type MetricKind int

const (
	KindCounter MetricKind = iota
	KindUpDownCounter
	KindHistogram
	KindGauge
)

// MetricDefinition 指标定义
type MetricDefinition[T ~int64 | ~float64] struct {
	Name        string
	Description string
	Unit        string
	Kind        MetricKind
}

// MetricValue 指标值
type MetricValue[T ~int64 | ~float64] struct {
	Name       string
	Value      T
	Attributes []attribute.KeyValue
}

// Metric 接口定义通用指标操作
type Metric[T ~int64 | ~float64] interface {
	Record(ctx context.Context, value MetricValue[T]) error
}

// DynamicMeter 动态指标
type DynamicMeter[T ~int64 | ~float64] struct {
	meter   metric.Meter
	metrics map[string]Metric[T]
	lock    sync.RWMutex
}

// NewDynamicMeter 创建特定类型的 DynamicMeter
func NewDynamicMeter[T ~int64 | ~float64](meter metric.Meter) *DynamicMeter[T] {
	return &DynamicMeter[T]{
		meter:   meter,
		metrics: make(map[string]Metric[T], 20),
		lock:    sync.RWMutex{},
	}
}

// metricWrapper 包装不同类型的指标
type metricWrapper[T ~int64 | ~float64] struct {
	counter       metric.Float64Counter
	upDownCounter metric.Float64UpDownCounter
	histogram     metric.Float64Histogram
	gauge         metric.Float64Gauge
	kind          MetricKind
}

func (m *metricWrapper[T]) Record(ctx context.Context, value MetricValue[T]) error {
	v := float64(value.Value)
	switch m.kind {
	case KindCounter:
		m.counter.Add(ctx, v, metric.WithAttributes(value.Attributes...))
	case KindUpDownCounter:
		m.upDownCounter.Add(ctx, v, metric.WithAttributes(value.Attributes...))
	case KindHistogram:
		m.histogram.Record(ctx, v, metric.WithAttributes(value.Attributes...))
	case KindGauge:
		m.gauge.Record(ctx, v, metric.WithAttributes(value.Attributes...))
	default:
		return fmt.Errorf("unknown metric kind: %v", m.kind)
	}
	return nil
}

// GetOrCreateMetric 获取或创建指标
func (dm *DynamicMeter[T]) GetOrCreateMetric(def MetricDefinition[T]) (Metric[T], error) {
	dm.lock.RLock()
	if metric, exists := dm.metrics[def.Name]; exists {
		dm.lock.RUnlock()
		return metric, nil
	}
	dm.lock.RUnlock()

	dm.lock.Lock()
	defer dm.lock.Unlock()

	// 双重检查
	if metric, exists := dm.metrics[def.Name]; exists {
		return metric, nil
	}

	var wrapper metricWrapper[T]
	var err error

	switch def.Kind {
	case KindCounter:
		wrapper.counter, err = dm.meter.Float64Counter(
			def.Name,
			metric.WithDescription(def.Description),
			metric.WithUnit(def.Unit),
		)
	case KindUpDownCounter:
		wrapper.upDownCounter, err = dm.meter.Float64UpDownCounter(
			def.Name,
			metric.WithDescription(def.Description),
			metric.WithUnit(def.Unit),
		)
	case KindHistogram:
		wrapper.histogram, err = dm.meter.Float64Histogram(
			def.Name,
			metric.WithDescription(def.Description),
			metric.WithUnit(def.Unit),
		)
	case KindGauge:
		wrapper.gauge, err = dm.meter.Float64Gauge(
			def.Name,
			metric.WithDescription(def.Description),
			metric.WithUnit(def.Unit),
		)
	default:
		return nil, fmt.Errorf("unsupported metric kind: %v", def.Kind)
	}

	if err != nil {
		return nil, err
	}

	wrapper.kind = def.Kind
	dm.metrics[def.Name] = &wrapper
	return &wrapper, nil
}

// RecordMetric 记录单个指标
func (dm *DynamicMeter[T]) RecordMetric(ctx context.Context, value MetricValue[T]) error {
	dm.lock.RLock()
	dmMetric, exists := dm.metrics[value.Name]
	dm.lock.RUnlock()

	if !exists {
		return fmt.Errorf("metric not found: %s", value.Name)
	}

	return dmMetric.Record(ctx, value)
}

// RecordBatch 记录多个指标
func (dm *DynamicMeter[T]) RecordBatch(ctx context.Context, values []MetricValue[T]) error {
	for _, value := range values {
		if err := dm.RecordMetric(ctx, value); err != nil {
			return err
		}
	}
	return nil
}

// 使用示例
type Int64DynamicMeter = DynamicMeter[int64]
type Float64DynamicMeter = DynamicMeter[float64]
