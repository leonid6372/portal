package main

import (
	"portal/internal/config"
	"portal/internal/storage/postgres"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	storage, err := postgres.New(cfg.StoragePath)
}
