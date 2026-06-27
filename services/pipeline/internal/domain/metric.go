// services/pipeline/internal/domain/metric.go
package domain

import "time"

// NormalizedMetric – метрика после стадии нормализации.
type NormalizedMetric struct {
	Source    string
	Name      string
	Value     float64
	Labels    map[string]string
	Timestamp time.Time
}

// AggregatedMetric – агрегированная метрика за окно.
type AggregatedMetric struct {
	Source string
	Name   string
	Sum    float64
	Count  int64
	Min    float64
	Max    float64
	P50    float64
	P95    float64
	P99    float64
	Window time.Duration
}
