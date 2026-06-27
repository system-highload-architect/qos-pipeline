// services/pipeline/internal/stages/write.go

package stages

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/system-highload-architect/go-solutions/net/breaker"
	"github.com/system-highload-architect/qos-pipeline/services/pipeline/internal/domain"
)

var (
	// dbBreaker — глобальный Circuit Breaker для записи в PostgreSQL.
	// В реальном проекте он создаётся один раз в main и передаётся сюда.
	dbBreaker = breaker.New("postgres-write", 3, 10*time.Second)
)

// WriteStage записывает агрегированную метрику в PostgreSQL, защищаясь Circuit Breaker'ом.
func WriteStage(ctx context.Context, agg domain.AggregatedMetric) error {
	err := dbBreaker.Execute(ctx, func() error {
		// Здесь должен быть реальный вызов PostgreSQL.
		// Пока эмулируем успешную запись.
		slog.Info("writing aggregate to PostgreSQL",
			"source", agg.Source,
			"name", agg.Name,
			"sum", agg.Sum,
			"count", agg.Count,
			"p95", agg.P95,
		)
		return nil
	})
	if errors.Is(err, breaker.ErrCircuitOpen) {
		slog.Warn("circuit breaker open, skipping write")
		return nil // не фатально для демо
	}
	return err
}
