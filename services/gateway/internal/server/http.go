package server

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/goccy/go-json"
	"github.com/system-highload-architect/qos-pipeline/services/gateway/internal/ports"
)

type HTTPServer struct {
	server        *http.Server
	logger        *slog.Logger
	ingesterPort  ports.IngesterPort
	aggregatorURL string
	port          int
}

type Option func(*HTTPServer)

func WithPort(port int) Option {
	return func(s *HTTPServer) { s.port = port }
}

func WithLogger(l *slog.Logger) Option {
	return func(s *HTTPServer) { s.logger = l }
}

func WithIngesterPort(p ports.IngesterPort) Option {
	return func(s *HTTPServer) { s.ingesterPort = p }
}

func WithAggregatorURL(rawURL string) Option {
	return func(s *HTTPServer) { s.aggregatorURL = rawURL }
}

func NewHTTPServer(opts ...Option) *HTTPServer {
	s := &HTTPServer{}
	for _, opt := range opts {
		opt(s)
	}

	mux := http.NewServeMux()
	// Приём метрик от клиентов
	mux.HandleFunc("/api/v1/metrics", s.handleMetrics)
	// Проксирование GraphQL к Aggregator
	if s.aggregatorURL != "" {
		u, _ := url.Parse(s.aggregatorURL)
		mux.Handle("/graphql", httputil.NewSingleHostReverseProxy(u))
	}

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	return s
}

func (s *HTTPServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}
	var req struct {
		Source  string         `json:"source"`
		Metrics []ports.Metric `json:"metrics"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := s.ingesterPort.IngestMetrics(r.Context(), req.Source, req.Metrics); err != nil {
		s.logger.Error("ingest metrics error", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"accepted"}`))
}

func (s *HTTPServer) ListenAndServe() error {
	return s.server.ListenAndServe()
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
