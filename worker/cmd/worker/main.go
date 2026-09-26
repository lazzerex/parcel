package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"

	"parcel/worker/internal/awsconfig"
	"parcel/worker/internal/metrics"
	"parcel/worker/internal/models"
	"parcel/worker/internal/notify"
	"parcel/worker/internal/processors"
	"parcel/worker/internal/store"
)

type worker struct {
	s3Client       processors.S3GetObjectAPI
	s3PutClient    processors.S3PutObjectAPI
	dynamodbClient store.DynamoDBAPI
	publisher      *notify.Publisher
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

	start := time.Now()
	result, err := processor.Run(ctx, w.s3Client, job)
	duration := time.Since(start)

	if err != nil {
		log.Error("processing failed", "error", err, "duration_ms", duration.Milliseconds())
		metrics.Record(metrics.ProcessingMetrics{
			Operation: job.Operation, Status: "FAILED", Duration: duration,
		})
		if setErr := store.SetStatus(ctx, w.dynamodbClient, w.table, job.FileID, "FAILED"); setErr != nil {
			log.Error("failed to set failed status", "error", setErr)
		}
		return err
	}

	if err := store.SetCompleted(ctx, w.dynamodbClient, w.table, job.FileID, result.Size, result.SHA256, result.ContentType); err != nil {
		log.Error("failed to write result", "error", err)
		return err
	}

	log.Info("completed", "size", result.Size, "sha256", result.SHA256, "content_type", result.ContentType, "duration_ms", duration.Milliseconds())
	metrics.Record(metrics.ProcessingMetrics{
		Operation: job.Operation, Status: "COMPLETED", Duration: duration, Size: result.Size,
	})

	if w.publisher != nil {
		if pubErr := w.publisher.Publish(ctx, notify.Event{
			JobID: job.JobID, FileID: job.FileID, Bucket: job.Bucket, Key: job.Key,
			Operation: job.Operation, Status: "COMPLETED", Size: result.Size, SHA256: result.SHA256,
		}); pubErr != nil {
			log.Error("failed to publish event", "error", pubErr)
		}
	}

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

	s3Client := s3.NewFromConfig(cfg)
	w := worker{
		s3Client:       s3Client,
		s3PutClient:    s3Client,
		dynamodbClient: dynamodb.NewFromConfig(cfg),
		table:          table,
	}

	if topicARN := os.Getenv("SNS_TOPIC_ARN"); topicARN != "" {
		w.publisher = notify.New(
			sns.NewFromConfig(cfg),
			eventbridge.NewFromConfig(cfg),
			topicARN,
			os.Getenv("EVENTBUS_NAME"),
		)
	}

	slog.Info("worker configured", "region", cfg.Region, "table", table)
	lambda.Start(w.handleRequest)
}
