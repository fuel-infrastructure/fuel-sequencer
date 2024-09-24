package prometheus

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type MetricsServer struct {
	log *zap.Logger
	cfg Config
	srv *http.Server
}

func NewMetricsServer(log *zap.Logger, cfg Config) *MetricsServer {
	return &MetricsServer{
		log: log,
		cfg: cfg,
	}
}

func (s *MetricsServer) Start() {

	s.srv = &http.Server{
		Addr: s.cfg.ListenAddress,
		Handler: promhttp.InstrumentMetricHandler(
			prometheus.DefaultRegisterer, promhttp.HandlerFor(
				prometheus.DefaultGatherer,
				promhttp.HandlerOpts{MaxRequestsInFlight: s.cfg.MaxOpenConnections},
			),
		),
		ReadHeaderTimeout: s.cfg.ReadHeaderTimeout,
		WriteTimeout:      s.cfg.WriteTimeout,
	}

	s.log.Info("starting metrics server", zap.String("address", s.cfg.ListenAddress))

	go func() {
		if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
			// Error starting or closing listener:
			s.log.Error("Prometheus HTTP server ListenAndServe", zap.Error(err))
		}
	}()
}

func (s *MetricsServer) Stop() {
	if err := s.srv.Shutdown(context.Background()); err != nil {
		// Error from closing listeners, or context timeout:
		s.log.Error("Prometheus HTTP server Shutdown", zap.Error(err))
	}
}
