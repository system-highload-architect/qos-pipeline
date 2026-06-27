package postgres

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/system-highload-architect/go-solutions/net/breaker"
	"github.com/system-highload-architect/qos-pipeline/services/aggregator/internal/domain"
)

// AggregateStore сохраняет и читает агрегированные метрики.
type AggregateStore struct {
	pool    *pgxpool.Pool
	breaker *breaker.Breaker
}

// NewAggregateStore создаёт пул и выполняет миграцию.
func NewAggregateStore(ctx context.Context, dsn string, cb *breaker.Breaker) (*AggregateStore, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("postgres ping: %w", err)
	}
	s := &AggregateStore{pool: pool, breaker: cb}
	if err := s.migrate(ctx); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *AggregateStore) migrate(ctx context.Context) error {
	migration, err := os.ReadFile("migrations/001_init.sql")
	if err != nil {
		return fmt.Errorf("read migration: %w", err)
	}
	_, err = s.pool.Exec(ctx, string(migration))
	return err
}

// InsertAggregate сохраняет агрегат в таблицу metrics.
func (s *AggregateStore) InsertAggregate(ctx context.Context, agg domain.AggregatedMetric) error {
	return s.breaker.Execute(ctx, func() error {
		_, err := s.pool.Exec(ctx,
			`INSERT INTO metrics (source, name, sum, count, min, max, p50, p95, p99, window_seconds)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			agg.Source, agg.Name, agg.Sum, agg.Count,
			agg.Min, agg.Max, agg.P50, agg.P95, agg.P99,
			int64(agg.Window.Seconds()),
		)
		return err
	})
}

// Close освобождает пул соединений.
func (s *AggregateStore) Close() {
	s.pool.Close()
}
