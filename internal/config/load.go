package config

import (
	"context"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

func Load(cfg any) error {
	godotenv.Load()
	if err := envconfig.Process(context.Background(), cfg); err != nil {
		return err
	}
	return nil
}
