package storage

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"go.uber.org/fx"
	"google.golang.org/api/option"

	"github.com/vladazn/danish/config"
)

type NewFirestoreClientParams struct {
	fx.In
	Cfg *config.FirebaseConfig
}

type FirestoreClient struct {
	cfg    *config.FirebaseConfig
	Client *firestore.Client
}

func (fc *FirestoreClient) Connect(ctx context.Context) error {
	client, err := firestore.NewClient(ctx, fc.cfg.ProjectId, option.WithCredentialsJSON(fc.cfg.AuthJSON))
	if err != nil {
		return fmt.Errorf("failed connect firestore: %w", err)
	}

	fc.Client = client

	return nil
}

func NewFirestoreClient(p NewFirestoreClientParams) *FirestoreClient {
	return &FirestoreClient{
		cfg: p.Cfg,
	}
}
