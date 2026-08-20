package main

import (
	"context"
	"log/slog"
	"os"

	"parcel/worker/internal/awsconfig"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := awsconfig.Load(context.Background())
	if err != nil {
		slog.Error("failed to load AWS config", "error", err)
		os.Exit(1)
	}

	endpoint := "default"
	if cfg.BaseEndpoint != nil {
		endpoint = *cfg.BaseEndpoint
	}

	slog.Info("worker configured", "region", cfg.Region, "endpoint", endpoint)
}
