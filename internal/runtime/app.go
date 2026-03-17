package runtime

import (
	"context"
	"fmt"
	"net/http"

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
			Addr:    addr,
			Handler: httpSrv.Handler(),
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
		_ = a.http.Shutdown(context.Background())
		return nil
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}
