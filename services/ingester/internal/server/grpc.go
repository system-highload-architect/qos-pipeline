package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	ingesterv1 "github.com/system-highload-architect/qos-pipeline/pb/ingester/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Publisher — интерфейс для отправки сообщений в шину.
type Publisher interface {
	Publish(ctx context.Context, key, value []byte) error
}

type IngesterServer struct {
	ingesterv1.UnimplementedIngesterServiceServer
	publisher Publisher
	logger    *slog.Logger
}

func NewIngesterServer(publisher Publisher, logger *slog.Logger) *IngesterServer {
	return &IngesterServer{
		publisher: publisher,
		logger:    logger,
	}
}

func (s *IngesterServer) IngestMetrics(ctx context.Context, req *ingesterv1.IngestRequest) (*ingesterv1.IngestResponse, error) {
	if req.Source == "" {
		return nil, status.Error(codes.InvalidArgument, "source is required")
	}
	if len(req.Metrics) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one metric is required")
	}

	// Преобразуем метрики в JSON для отправки в Kafka
	payload := struct {
		Source    string               `json:"source"`
		Metrics   []*ingesterv1.Metric `json:"metrics"`
		Timestamp time.Time            `json:"ingested_at"`
	}{
		Source:    req.Source,
		Metrics:   req.Metrics,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		s.logger.Error("failed to marshal metrics", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to marshal metrics: %v", err)
	}

	// Публикуем в Kafka (ключ – источник)
	if err := s.publisher.Publish(ctx, []byte(req.Source), data); err != nil {
		s.logger.Error("failed to publish to kafka", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to publish to kafka: %v", err)
	}

	s.logger.Info("metrics ingested", "source", req.Source, "count", len(req.Metrics))

	return &ingesterv1.IngestResponse{
		Accepted: true,
		Count:    int32(len(req.Metrics)),
	}, nil
}
