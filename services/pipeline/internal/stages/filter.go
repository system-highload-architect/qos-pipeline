package stages

import (
	"context"
	"fmt"
	"time"

	"github.com/system-highload-architect/go-solutions/data/timedcache"
	"github.com/system-highload-architect/qos-pipeline/services/pipeline/internal/domain"
)

// FilterStage отбрасывает дубликаты и выбросы.
func FilterStage(ctx context.Context, m domain.NormalizedMetric) (domain.NormalizedMetric, error) {
	// простая проверка: значение не отрицательное
	if m.Value < 0 {
		return domain.NormalizedMetric{}, fmt.Errorf("negative value")
	}
	return m, nil
}

// NewDeduplicator создаёт кэш для дедупликации (TTL = окно агрегации).
func NewDeduplicator(window time.Duration) *timedcache.Cache[string, bool] {
	return timedcache.New[string, bool](window)
}

// IsDuplicate проверяет, был ли уже такой ключ.
func IsDuplicate(cache *timedcache.Cache[string, bool], key string) bool {
	if _, exists := cache.Get(key); exists {
		return true
	}
	cache.Set(key, true)
	return false
}
