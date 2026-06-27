// services/ingester/cmd/main.go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/system-highload-architect/go-solutions/config"
	"github.com/system-highload-architect/go-solutions/logger"
	"github.com/system-highload-architect/go-solutions/metrics"
	"github.com/system-highload-architect/go-solutions/shutdown"

	ingesterv1 "github.com/system-highload-architect/qos-pipeline/pb/ingester/v1"
	"github.com/system-highload-architect/qos-pipeline/services/ingester/internal/adapters/kafka"
	"github.com/system-highload-architect/qos-pipeline/services/ingester/internal/server"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// AppConfig – корневая конфигурация сервиса.
type AppConfig struct {
	Server  ServerConfig  `yaml:"server"`
	Kafka   KafkaConfig   `yaml:"kafka"`
	Log     LogConfig     `yaml:"log"`
	Metrics MetricsConfig `yaml:"metrics"`
}

type ServerConfig struct {
	Port int `yaml:"port" env:"SERVER_PORT"`
}

type KafkaConfig struct {
	Brokers []string `yaml:"brokers" env:"KAFKA_BROKERS"`
	Topic   string   `yaml:"topic" env:"KAFKA_TOPIC"`
}

type LogConfig struct {
	Level  string `yaml:"level" env:"LOG_LEVEL"`
	Format string `yaml:"format" env:"LOG_FORMAT"`
}

type MetricsConfig struct {
	UseOTLP bool `yaml:"use_otlp" env:"METRICS_USE_OTLP"`
}

func main() {
	// 1. Загрузка конфигурации
	var cfg AppConfig
	if err := config.Load(&cfg, config.WithPath("configs/dev.yaml")); err != nil {
		slog.Error("cannot load config", "error", err)
		os.Exit(1)
	}

	// 2. Инициализация логгера
	log := logger.New(cfg.Log.Level, cfg.Log.Format, slog.String("service", "ingester"))
	log.Info("starting ingester")

	// 3. Инициализация метрик
	if err := metrics.Init(context.Background(), "ingester", cfg.Metrics.UseOTLP); err != nil {
		log.Error("failed to init metrics", "error", err)
		os.Exit(1)
	}
	defer metrics.Shutdown(context.Background())

	// 4. Kafka продюсер
	// publisher, err := kafka.NewPublisher(cfg.Kafka.Brokers, cfg.Kafka.Topic, log)
	// if err != nil {
	// 	log.Error("failed to create kafka publisher", "error", err)
	// 	os.Exit(1)
	// }
	publisher := &kafka.MockPublisher{Logger: log}
	defer publisher.Close()

	// 5. gRPC сервер
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)

	ingesterSrv := server.NewIngesterServer(publisher, log)
	// Регистрация сгенерированного сервиса будет после генерации proto
	// pb.RegisterIngesterServer(grpcServer, ingesterSrv)

	ingesterv1.RegisterIngesterServiceServer(grpcServer, ingesterSrv)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		log.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	go func() {
		log.Info("gRPC server listening", "port", cfg.Server.Port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("gRPC server error", "error", err)
			os.Exit(1)
		}
	}()

	// 6. Graceful shutdown
	shutdownMgr := shutdown.NewManager(30 * time.Second)
	shutdownMgr.SetLogger(log)
	shutdownMgr.Add("gRPC server", 0, func(ctx context.Context) error {
		grpcServer.GracefulStop()
		return nil
	}, 5*time.Second)
	shutdownMgr.Add("kafka publisher", 1, func(ctx context.Context) error {
		return publisher.Close()
	}, 3*time.Second)

	shutdownMgr.Wait()
}
