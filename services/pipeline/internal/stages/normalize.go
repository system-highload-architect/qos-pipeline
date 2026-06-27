package stages

import (
	"context"
	"encoding/json"
	"time"

	"github.com/system-highload-architect/qos-pipeline/services/pipeline/internal/domain"
)

// NormalizeStage принимает сырые байты и возвращает NormalizedMetric.
func NormalizeStage(ctx context.Context, raw []byte) (domain.NormalizedMetric, error) {
	var payload struct {
		Source  string `json:"source"`
		Metrics []struct {
			Name      string            `json:"name"`
			Value     float64           `json:"value"`
			Labels    map[string]string `json:"labels"`
			Timestamp int64             `json:"timestamp_unix"`
		} `json:"metrics"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return domain.NormalizedMetric{}, err
	}
	// пока берём первую метрику для простоты (позже можно итерироваться)
	if len(payload.Metrics) == 0 {
		return domain.NormalizedMetric{}, nil
	}
	m := payload.Metrics[0]
	return domain.NormalizedMetric{
		Source:    payload.Source,
		Name:      m.Name,
		Value:     m.Value,
		Labels:    m.Labels,
		Timestamp: time.Unix(m.Timestamp, 0),
	}, nil
}
