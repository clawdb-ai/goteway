package runtime

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/mac/goteway/internal/config"
	"github.com/mac/goteway/internal/observability"
	"github.com/mac/goteway/internal/transport/httpapi"
)

const version = "go-gateway/0.1.0"

// App wires transport + core logic.
type App struct {
	cfg    config.Config
	http   *http.Server
	logic  *Service
	health *observability.HealthReporter
}

func NewAppFromEnv() (*App, error) {
	cfg, err := config.FromEnv()
	if err != nil {
		return nil, err
	}
	logic := NewService(cfg)
	health := observability.NewHealthReporter(version)
	httpSrv := httpapi.New(logic, health)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	app := &App{
		cfg:    cfg,
		logic:  logic,
		health: health,
		http: &http.Server{
			Addr:              addr,
			Handler:           httpSrv.Handler(),
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       120 * time.Second,
		},
	}
	return app, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- a.http.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(a.cfg.ShutdownSec)*time.Second)
		defer cancel()
		_ = a.http.Shutdown(shutdownCtx)
		return nil
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}
