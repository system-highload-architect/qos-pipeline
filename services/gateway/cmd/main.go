package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/system-highload-architect/go-solutions/config"
	"github.com/system-highload-architect/go-solutions/logger"
	"github.com/system-highload-architect/go-solutions/shutdown"
	ingesterv1 "github.com/system-highload-architect/qos-pipeline/pb/ingester/v1"
	"github.com/system-highload-architect/qos-pipeline/services/gateway/internal/adapters/grpcclient"
	"github.com/system-highload-architect/qos-pipeline/services/gateway/internal/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AppConfig struct {
	Server   ServerConfig   `yaml:"server"`
	GRPC     GRPCConfig     `yaml:"grpc"`
	Security SecurityConfig `yaml:"security"`
	Log      LogConfig      `yaml:"log"`
}

type ServerConfig struct {
	Port int `yaml:"port" env:"SERVER_PORT"`
}

type GRPCConfig struct {
	Ingester   string `yaml:"ingester" env:"GRPC_INGESTER"`
	Aggregator string `yaml:"aggregator" env:"GRPC_AGGREGATOR"`
}

type SecurityConfig struct {
	RateLimit      float64       `yaml:"rate_limit"`
	RateBurst      float64       `yaml:"rate_burst"`
	IdempotencyTTL time.Duration `yaml:"idempotency_ttl"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

func main() {
	var cfg AppConfig
	if err := config.Load(&cfg, config.WithPath("configs/dev.yaml")); err != nil {
		slog.Error("cannot load config", "error", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Log.Level, cfg.Log.Format, slog.String("service", "gateway"))
	log.Info("starting gateway")

	// gRPC-клиент к Ingester
	conn, err := grpc.NewClient(cfg.GRPC.Ingester, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("cannot connect to ingester", "error", err)
		os.Exit(1)
	}
	defer conn.Close()
	ingesterClient := ingesterv1.NewIngesterServiceClient(conn)
	ingesterPort := grpcclient.NewIngesterPort(ingesterClient)

	// HTTP-сервер
	srv := server.NewHTTPServer(
		server.WithPort(cfg.Server.Port),
		server.WithLogger(log),
		server.WithIngesterPort(ingesterPort),
		server.WithAggregatorURL("http://aggregator:9002"),
	)

	go func() {
		log.Info("gateway listening", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	shutdownMgr := shutdown.NewManager(30 * time.Second)
	shutdownMgr.SetLogger(log)
	shutdownMgr.Add("http_server", 0, func(ctx context.Context) error {
		return srv.Shutdown(ctx)
	}, 5*time.Second)
	shutdownMgr.Wait()
}
