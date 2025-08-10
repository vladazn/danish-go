package app

import (
	"go.uber.org/fx"

	"github.com/vladazn/danish/app/classroom"
	"github.com/vladazn/danish/app/server"
	"github.com/vladazn/danish/app/storage"
	"github.com/vladazn/danish/common/logger"
	"github.com/vladazn/danish/common/rand"
	"github.com/vladazn/danish/config"
)

func Run() {
	fx.New(
		config.Module,
		fx.Provide(
			rand.New,
			logger.NewLogger,
		),
		storage.Module,
		classroom.Module,
		server.Module,
	).Run()
}
