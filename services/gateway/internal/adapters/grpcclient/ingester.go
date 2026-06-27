package grpcclient

import (
	"context"
	"fmt"

	ingesterv1 "github.com/system-highload-architect/qos-pipeline/pb/ingester/v1"
	"github.com/system-highload-architect/qos-pipeline/services/gateway/internal/ports"
)

type ingesterAdapter struct {
	client ingesterv1.IngesterServiceClient
}

func NewIngesterPort(client ingesterv1.IngesterServiceClient) ports.IngesterPort {
	return &ingesterAdapter{client: client}
}

func (a *ingesterAdapter) IngestMetrics(ctx context.Context, source string, metrics []ports.Metric) error {
	pbMetrics := make([]*ingesterv1.Metric, len(metrics))
	for i, m := range metrics {
		pbMetrics[i] = &ingesterv1.Metric{
			Name:          m.Name,
			Value:         m.Value,
			Labels:        m.Labels,
			TimestampUnix: m.Timestamp,
		}
	}
	_, err := a.client.IngestMetrics(ctx, &ingesterv1.IngestRequest{
		Source:  source,
		Metrics: pbMetrics,
	})
	if err != nil {
		return fmt.Errorf("ingester: %w", err)
	}
	return nil
}
