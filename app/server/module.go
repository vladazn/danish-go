package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/vladazn/danish/config"
)

var Module = fx.Module(
	"server",
	fx.Provide(
		NewFirebaseClient,
		NewRouter,
	),
	fx.Invoke(
		RegisterHooks,
	),
)

type ServerParams struct {
	fx.In

	Lifecycle      fx.Lifecycle
	Cfg            *config.HttpServerConfig
	FirebaseClient *FirebaseClient
	Handler        *chi.Mux
	Logger         *zap.Logger
}

func RegisterHooks(p ServerParams) {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", p.Cfg.Port),
		Handler: p.Handler,
	}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			err := p.FirebaseClient.Connect(ctx)
			if err != nil {
				return err
			}
			go func() {
				p.Logger.Info("Starting server on " + srv.Addr)
				err = srv.ListenAndServe()
				if !errors.Is(err, http.ErrServerClosed) {
					p.Logger.Error("Error starting server", zap.Error(err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("Stopping HTTP server...")
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			return srv.Shutdown(ctx)
		},
	})
}
