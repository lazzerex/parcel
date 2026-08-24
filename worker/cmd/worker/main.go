package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"parcel/worker/internal/awsconfig"
	"parcel/worker/internal/models"
	"parcel/worker/internal/processors"
	"parcel/worker/internal/store"
)

type worker struct {
	s3Client       *s3.Client
	dynamodbClient *dynamodb.Client
	table          string
}

func (w worker) handleRecord(ctx context.Context, record events.SQSMessage) error {
	job, err := models.DecodeJob(record.Body)
	if err != nil {
		slog.Error("invalid job", "message_id", record.MessageId, "error", err)
		return err
	}

	log := slog.With("job_id", job.JobID, "file_id", job.FileID)

	processor, err := processors.Dispatch(job.Operation)
	if err != nil {
		log.Error("unknown operation", "operation", job.Operation, "error", err)
		return err
	}

	status, err := store.GetStatus(ctx, w.dynamodbClient, w.table, job.FileID)
	if err != nil {
		log.Error("failed to read status", "error", err)
		return err
	}
	if status == "COMPLETED" {
		log.Info("already completed, skipping")
		return nil
	}

	if err := store.SetStatus(ctx, w.dynamodbClient, w.table, job.FileID, "PROCESSING"); err != nil {
		log.Error("failed to set processing status", "error", err)
		return err
	}

	result, err := processor.Run(ctx, w.s3Client, job)
	if err != nil {
		log.Error("processing failed", "error", err)
		if setErr := store.SetStatus(ctx, w.dynamodbClient, w.table, job.FileID, "FAILED"); setErr != nil {
			log.Error("failed to set failed status", "error", setErr)
		}
		return err
	}

	if err := store.SetCompleted(ctx, w.dynamodbClient, w.table, job.FileID, result.Size, result.SHA256, result.ContentType); err != nil {
		log.Error("failed to write result", "error", err)
		return err
	}

	log.Info("completed", "size", result.Size, "sha256", result.SHA256, "content_type", result.ContentType)
	return nil
}

func (w worker) handleRequest(ctx context.Context, event events.SQSEvent) error {
	for _, record := range event.Records {
		if err := w.handleRecord(ctx, record); err != nil {
			return err
		}
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

	table := os.Getenv("DYNAMODB_TABLE")
	if table == "" {
		slog.Error("DYNAMODB_TABLE is not set")
		os.Exit(1)
	}

	w := worker{
		s3Client:       s3.NewFromConfig(cfg),
		dynamodbClient: dynamodb.NewFromConfig(cfg),
		table:          table,
	}

	slog.Info("worker configured", "region", cfg.Region, "table", table)
	lambda.Start(w.handleRequest)
}
