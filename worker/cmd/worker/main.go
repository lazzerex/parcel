package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"parcel/worker/internal/awsconfig"
)

func handleRequest(ctx context.Context, event events.SQSEvent) error {
	for _, record := range event.Records {
		// Processing (Phase 6/7) is not implemented yet; this confirms the
		// SQS -> Lambda wiring and gives visibility while that lands.
		slog.Info("received job", "message_id", record.MessageId, "body", record.Body)
	}
	return nil
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := awsconfig.Load(context.Background())
	if err != nil {
		slog.Error("failed to load AWS config", "error", err)
		os.Exit(1)
	}
	slog.Info("worker configured", "region", cfg.Region)

	lambda.Start(handleRequest)
}
