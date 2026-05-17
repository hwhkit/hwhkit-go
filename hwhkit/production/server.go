package production

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hwhkit/hwhkit-go/core"
	"github.com/hwhkit/hwhkit-go/hwhkit/middleware"
	"github.com/hwhkit/hwhkit-go/observability"
	"github.com/prometheus/client_golang/prometheus"
)

func Serve(ctx context.Context, built *core.BuiltApplication, logger *slog.Logger) error {
	if logger == nil {
		logger = slog.Default()
	}
	cfg := built.Config()

	if reg, ok := built.MetricsRegistry().(*prometheus.Registry); !ok || reg == nil {
		reg := observability.InitMetrics()
		observability.RegisterHTTPMetrics(reg)
		built.SetMetricsRegistry(reg)
	} else {
		observability.RegisterHTTPMetrics(reg)
	}

	root := chi.NewRouter()
	root.Use(middleware.RequestID)
	root.Use(middleware.Recover(logger))
	root.Use(middleware.AccessLog(logger))
	root.Use(middleware.CORS(middleware.DefaultCORS()))
	root.Use(middleware.Compress(5))
	root.Use(middleware.BodyLimit(cfg.Server.BodyLimitBytes))
	if h := cfg.Server.HandlerTimeout(); h > 0 {
		root.Use(middleware.Timeout(h))
	}
	root.Use(middleware.HTTPMetrics())

	mountEndpoints(root, built)

	root.Mount("/", built.Router())
	built.SetRouter(root)

	srv := &http.Server{
		Addr:         cfg.Server.BindAddr,
		Handler:      root,
		ReadTimeout:  cfg.Server.ReadTimeout(),
		WriteTimeout: cfg.Server.WriteTimeout(),
		IdleTimeout:  cfg.Server.IdleTimeout(),
		BaseContext:  func(net.Listener) context.Context { return ctx },
	}

	signalCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server starting", "addr", cfg.Server.BindAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("listen: %w", err)
	case <-signalCtx.Done():
		logger.Info("shutdown signal received")
	}

	built.Shutdown().Cancel()

	drain := cfg.Server.DrainTimeout()
	if drain <= 0 {
		drain = 30 * time.Second
	}
	shutCtx, cancel := context.WithTimeout(context.Background(), drain)
	defer cancel()

	if err := srv.Shutdown(shutCtx); err != nil {
		logger.Error("http shutdown failed", "err", err)
	}

	providers := built.Providers()
	for i := len(providers) - 1; i >= 0; i-- {
		p := providers[i]
		key := p.Key()
		pctx, pcancel := context.WithTimeout(context.Background(), drain)
		if err := p.Shutdown(pctx, built.AppContext()); err != nil {
			logger.Error("provider shutdown failed", "key", key, "err", err)
		} else {
			logger.Info("provider shutdown clean", "key", key)
		}
		pcancel()
	}

	logger.Info("server stopped cleanly")
	return nil
}
