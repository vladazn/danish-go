package logger

import (
	"strings"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/vladazn/danish/config"
)

type Params struct {
	fx.In
	Cfg *config.LogConfig
}

func NewLogger(params Params) (*zap.Logger, error) {
	if strings.ToLower(params.Cfg.Level) == "debug" {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}
