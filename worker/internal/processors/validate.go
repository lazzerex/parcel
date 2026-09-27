package processors

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"parcel/worker/internal/models"
)

const (
	maxFileSize  = 10 * 1024 * 1024
	maxSniffSize = 512
)

var allowedContentTypes = []string{
	"image/jpeg",
	"image/png",
	"image/gif",
	"image/webp",
	"application/pdf",
	"text/plain",
	"application/json",
	"text/csv",
}

type ValidateProcessor struct{}

func (ValidateProcessor) Run(ctx context.Context, s3Client S3GetObjectAPI, job models.Job) (Result, error) {
	out, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(job.Bucket),
		Key:    aws.String(job.Key),
	})
	if err != nil {
		return Result{}, fmt.Errorf("get object: %w", err)
	}
	defer out.Body.Close()

	sniff := make([]byte, maxSniffSize)
	n, err := io.ReadFull(out.Body, sniff)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return Result{}, fmt.Errorf("read object: %w", err)
	}
	sniff = sniff[:n]

	if n == 0 {
		return Result{}, fmt.Errorf("validation failed: file is empty")
	}

	hasher := sha256.New()
	hasher.Write(sniff)
	size := int64(n)

	written, err := io.Copy(hasher, out.Body)
	if err != nil {
		return Result{}, fmt.Errorf("read object: %w", err)
	}
	size += written

	if size > maxFileSize {
		return Result{}, fmt.Errorf("validation failed: file size %d exceeds maximum %d", size, maxFileSize)
	}

	contentType := http.DetectContentType(sniff)
	if !contentTypeAllowed(contentType) {
		return Result{}, fmt.Errorf("validation failed: content type %q is not allowed", contentType)
	}

	return Result{
		Size:        size,
		SHA256:      hex.EncodeToString(hasher.Sum(nil)),
		ContentType: contentType,
	}, nil
}

func contentTypeAllowed(detected string) bool {
	for _, prefix := range allowedContentTypes {
		if strings.HasPrefix(detected, prefix) {
			return true
		}
	}
	return false
}