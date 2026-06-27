package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/system-highload-architect/go-solutions/config"
	"github.com/system-highload-architect/go-solutions/net/backpressure"
	"github.com/system-highload-architect/go-solutions/shutdown"
	"github.com/system-highload-architect/qos-pipeline/services/pipeline/internal/adapters/kafka"
	"github.com/system-highload-architect/qos-pipeline/services/pipeline/internal/domain"
	"github.com/system-highload-architect/qos-pipeline/services/pipeline/internal/stages"
)

// AppConfig – корневая конфигурация сервиса Pipeline.
type AppConfig struct {
	Kafka KafkaConfig `yaml:"kafka"`
	Log   LogConfig   `yaml:"log"`
}

type KafkaConfig struct {
	Brokers []string `yaml:"brokers" env:"KAFKA_BROKERS"`
	Topic   string   `yaml:"topic" env:"KAFKA_TOPIC"`
}

type LogConfig struct {
	Level  string `yaml:"level" env:"LOG_LEVEL"`
	Format string `yaml:"format" env:"LOG_FORMAT"`
}

func main() {
	// Загружаем конфигурацию
	var cfg AppConfig
	if err := config.Load(&cfg, config.WithPath("configs/dev.yaml")); err != nil {
		slog.Error("cannot load config", "error", err)
		os.Exit(1)
	}

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	log.Info("starting pipeline")

	// Стадии
	normalizeStage := func(ctx context.Context, raw []byte) (domain.NormalizedMetric, error) {
		return stages.NormalizeStage(ctx, raw)
	}
	filterStage := func(ctx context.Context, m domain.NormalizedMetric) (domain.NormalizedMetric, error) {
		return stages.FilterStage(ctx, m)
	}
	aggregator := stages.NewAggregator()
	aggregateStage := func(ctx context.Context, m domain.NormalizedMetric) (domain.AggregatedMetric, error) {
		return aggregator.AggregateStage(ctx, m)
	}
	writeStage := func(ctx context.Context, agg domain.AggregatedMetric) error {
		return stages.WriteStage(ctx, agg)
	}

	// Конвейер
	pipe := backpressure.NewPipeline[[]byte](
		[]backpressure.Stage[[]byte]{
			func(ctx context.Context, raw []byte) error {
				m, err := normalizeStage(ctx, raw)
				if err != nil {
					return err
				}
				m2, err := filterStage(ctx, m)
				if err != nil {
					return err
				}
				agg, err := aggregateStage(ctx, m2)
				if err != nil {
					return err
				}
				return writeStage(ctx, agg)
			},
		},
		backpressure.WithWorkers[[]byte](2),
		backpressure.WithBufferSize[[]byte](100),
	)

	input, output := pipe.Run()
	defer close(input)

	// Потребитель Kafka
	// consumer := kafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.Topic, "pipeline-group", log)
	consumer := &kafka.MockConsumer{
		Logger: log,
		Data: [][]byte{
			[]byte(`{"source":"web-1","metrics":[{"name":"http_requests_total","value":150,"labels":{"method":"GET"},"timestamp_unix":1719500000}]}`),
			[]byte(`{"source":"web-1","metrics":[{"name":"http_requests_total","value":120,"labels":{"method":"POST"},"timestamp_unix":1719500001}]}`),
			[]byte(`{"source":"web-2","metrics":[{"name":"http_requests_total","value":200,"labels":{"method":"GET"},"timestamp_unix":1719500002}]}`),
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	consumer.Start(ctx, input)

	go func() {
		for range output {
		}
	}()

	// Graceful shutdown
	shutdownMgr := shutdown.NewManager(10 * time.Second)
	shutdownMgr.SetLogger(log)
	shutdownMgr.Add("pipeline", 0, func(ctx context.Context) error {
		pipe.Wait()
		consumer.Close()
		return nil
	}, 5*time.Second)

	shutdownMgr.Wait()
	log.Info("pipeline stopped")
}
