package ports

import "context"

// IngesterPort — интерфейс для отправки метрик в Ingester.
type IngesterPort interface {
	IngestMetrics(ctx context.Context, source string, metrics []Metric) error
}

// Metric — упрощённая структура метрики.
type Metric struct {
	Name      string
	Value     float64
	Labels    map[string]string
	Timestamp int64
}
