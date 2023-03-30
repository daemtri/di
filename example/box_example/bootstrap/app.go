package bootstrap

import (
	"context"
	"time"

	"github.com/daemtri/di/example/box_example/contract"
	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"
)

type App struct {
	servers []contract.Server
	ctx     context.Context
	logger  *slog.Logger
}

func NewApp(ctx context.Context, servers []contract.Server, logger *slog.Logger) (*App, error) {
	return &App{servers: servers, ctx: ctx, logger: logger}, nil
}

func (app *App) Run() error {
	group, ctx := errgroup.WithContext(app.ctx)
	for _, server := range app.servers {
		s := server
		group.Go(func() error {
			return s.ListenAndServe()
		})
	}
	go func() {
		<-ctx.Done()
		for _, server := range app.servers {
			sCtx, cancel := context.WithTimeout(app.ctx, 5*time.Second)
			defer cancel()
			if err := server.Shutdown(sCtx); err != nil {
				app.logger.Warn("app shutdown cause error", "error", err)
			}
		}
	}()

	return group.Wait()
}
