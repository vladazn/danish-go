package storage

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module(
	"storage",
	fx.Provide(
		NewFirestoreClient,
	),
	fx.Invoke(
		RegisterHooks,
	),
)

type HooksParams struct {
	fx.In
	Logger          *zap.Logger
	Lifecycle       fx.Lifecycle
	FirestoreClient *FirestoreClient
}

func RegisterHooks(p HooksParams) {
	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			err := p.FirestoreClient.Connect(ctx)
			if err != nil {
				return fmt.Errorf("connect firestore client: %w", err)
			}

			return nil
		},
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("Stopping HTTP server...")
			return p.FirestoreClient.Client.Close()
		},
	})
}
