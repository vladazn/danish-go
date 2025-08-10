package server

import (
	"context"

	firebase "firebase.google.com/go/v4"
	"go.uber.org/fx"
	"google.golang.org/api/option"

	"github.com/vladazn/danish/config"
)

type FirebaseParams struct {
	fx.In
	Cfg *config.FirebaseConfig
}

type FirebaseClient struct {
	cfg *config.FirebaseConfig
	app *firebase.App
}

func (c *FirebaseClient) Connect(ctx context.Context) error {
	app, err := firebase.NewApp(ctx,
		&firebase.Config{ProjectID: c.cfg.ProjectId},
		option.WithCredentialsJSON(c.cfg.AuthJSON),
	)
	if err != nil {
		return err
	}
	c.app = app

	return nil
}

func NewFirebaseClient(p FirebaseParams) *FirebaseClient {
	return &FirebaseClient{
		cfg: p.Cfg,
	}
}
