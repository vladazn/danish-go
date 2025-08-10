package config

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"go.uber.org/fx"
)

type Config struct {
	fx.Out
	FirebaseConfig   FirebaseConfig   `envPrefix:"FIREBASE_"`
	LogConfig        LogConfig        `envPrefix:"LOG_"`
	HttpServerConfig HttpServerConfig `envPrefix:"HTTP_SERVER_"`
}

type Result struct {
	fx.Out
	FirebaseConfig   *FirebaseConfig
	LogConfig        *LogConfig
	HttpServerConfig *HttpServerConfig
}

type HttpServerConfig struct {
	Port int `env:"PORT" envDefault:"8080"`
}

type LogConfig struct {
	Level string `env:"LEVEL" envDefault:"info"`
}

type FirebaseConfig struct {
	AuthJSONBase64 string `env:"AUTH_JSON_BASE64"`
	AuthJSON       []byte
	ProjectId      string `env:"PROJECT_ID"`
}

func NewConfig() (Result, error) {
	cfg, err := env.ParseAs[Config]()
	_ = cfg
	if err != nil {
		return Result{}, fmt.Errorf("could not parse config: %w", err)
	}

	fmt.Println(os.Getenv("FIREBASE_PROJECT_ID"))

	fmt.Println(cfg)

	err = parseAuthJson(&cfg)
	if err != nil {
		return Result{}, err
	}

	return Result{
		FirebaseConfig:   &cfg.FirebaseConfig,
		LogConfig:        &cfg.LogConfig,
		HttpServerConfig: &cfg.HttpServerConfig,
	}, nil
}

func parseAuthJson(cfg *Config) error {
	data, err := base64.StdEncoding.DecodeString(cfg.FirebaseConfig.AuthJSONBase64)
	if err != nil {
		return fmt.Errorf("firebase auth json base64 decoding failed: %w", err)
	}

	cfg.FirebaseConfig.AuthJSON = data

	return nil
}
