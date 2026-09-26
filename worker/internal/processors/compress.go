package processors

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"parcel/worker/internal/models"
)

type CompressProcessor struct {
	PutClient S3PutObjectAPI
}

func (cp CompressProcessor) Run(ctx context.Context, s3Client S3GetObjectAPI, job models.Job) (Result, error) {
	out, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(job.Bucket),
		Key:    aws.String(job.Key),
	})
	if err != nil {
		return Result{}, fmt.Errorf("get object: %w", err)
	}
	defer out.Body.Close()

	raw, err := io.ReadAll(out.Body)
	if err != nil {
		return Result{}, fmt.Errorf("read object: %w", err)
	}

	if len(raw) == 0 {
		return Result{}, fmt.Errorf("cannot compress empty file")
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(raw); err != nil {
		return Result{}, fmt.Errorf("gzip write: %w", err)
	}
	if err := gz.Close(); err != nil {
		return Result{}, fmt.Errorf("gzip close: %w", err)
	}

	compressed := buf.Bytes()

	destKey := "processed/" + strings.TrimPrefix(job.Key, "uploads/")
	destKey = strings.TrimSuffix(destKey, path.Ext(destKey)) + ".gz"

	if cp.PutClient != nil {
		contentType := http.DetectContentType(compressed[:min(len(compressed), 512)])
		if _, err := cp.PutClient.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(job.Bucket),
			Key:         aws.String(destKey),
			Body:        bytes.NewReader(compressed),
			ContentType: aws.String(contentType),
		}); err != nil {
			return Result{}, fmt.Errorf("put compressed: %w", err)
		}
	}

	sum := sha256.Sum256(compressed)
	return Result{
		Size:        int64(len(compressed)),
		SHA256:      hex.EncodeToString(sum[:]),
		ContentType: "application/gzip",
	}, nil
}