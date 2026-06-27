package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/system-highload-architect/go-solutions/config"
	"github.com/system-highload-architect/go-solutions/logger"
	"github.com/system-highload-architect/go-solutions/shutdown"
	"github.com/system-highload-architect/qos-pipeline/services/aggregator/internal/adapters/postgres"
	"github.com/system-highload-architect/qos-pipeline/services/aggregator/internal/server"
)

type AppConfig struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Log      LogConfig      `yaml:"log"`
}

type ServerConfig struct {
	Port int `yaml:"port" env:"SERVER_PORT"`
}

type DatabaseConfig struct {
	DSN string `yaml:"dsn" env:"DATABASE_DSN"`
}

type LogConfig struct {
	Level  string `yaml:"level" env:"LOG_LEVEL"`
	Format string `yaml:"format" env:"LOG_FORMAT"`
}

func main() {
	var cfg AppConfig
	if err := config.Load(&cfg, config.WithPath("configs/dev.yaml")); err != nil {
		slog.Error("cannot load config", "error", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Log.Level, cfg.Log.Format, slog.String("service", "aggregator"))
	log.Info("starting aggregator")

	// Подключаемся к PostgreSQL, при неудаче используем заглушку
	var store *postgres.AggregateStore
	if cfg.Database.DSN != "" {
		var err error
		store, err = postgres.NewAggregateStore(context.Background(), cfg.Database.DSN, nil)
		if err != nil {
			log.Error("failed to connect to postgres, using stub", "error", err)
			store = nil
		}
	} else {
		log.Info("no DSN provided, using stub")
	}

	graphqlHandler, err := server.NewGraphQLHandler(store)
	if err != nil {
		log.Error("failed to create graphql handler", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.Handle("/graphql", graphqlHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: mux,
	}

	go func() {
		log.Info("aggregator listening", "port", cfg.Server.Port)
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
