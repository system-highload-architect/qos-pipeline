// services/pipeline/internal/stages/aggregate.go

package stages

import (
	"context"
	"sync"
	"time"

	"github.com/system-highload-architect/go-solutions/math/statistics"
	"github.com/system-highload-architect/qos-pipeline/services/pipeline/internal/domain"
)

// Aggregator хранит промежуточные агрегаты и считает статистики.
type Aggregator struct {
	mu      sync.Mutex
	windows map[string]*aggregateWindow
}

type aggregateWindow struct {
	values []float64
	sum    float64
	count  int64
	min    float64
	max    float64
	window time.Duration
}

func NewAggregator() *Aggregator {
	return &Aggregator{windows: make(map[string]*aggregateWindow)}
}

// AggregateStage добавляет метрику в окно и возвращает готовый агрегат,
// если окно заполнено. Пока использует простой порог по количеству.
func (a *Aggregator) AggregateStage(ctx context.Context, m domain.NormalizedMetric) (domain.AggregatedMetric, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	key := m.Source + ":" + m.Name
	w, ok := a.windows[key]
	if !ok {
		w = &aggregateWindow{
			window: 1 * time.Minute,
			min:    m.Value,
			max:    m.Value,
		}
		a.windows[key] = w
	}

	w.values = append(w.values, m.Value)
	w.sum += m.Value
	w.count++
	if m.Value < w.min {
		w.min = m.Value
	}
	if m.Value > w.max {
		w.max = m.Value
	}

	// Возвращаем агрегат с расчитанными статистиками.
	agg := domain.AggregatedMetric{
		Source: m.Source,
		Name:   m.Name,
		Sum:    w.sum,
		Count:  w.count,
		Min:    w.min,
		Max:    w.max,
		P50:    statistics.Percentile(w.values, 0.5),
		P95:    statistics.Percentile(w.values, 0.95),
		P99:    statistics.Percentile(w.values, 0.99),
		Window: w.window,
	}
	return agg, nil
}
