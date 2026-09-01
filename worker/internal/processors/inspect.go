package processors

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"parcel/worker/internal/models"
)

type InspectProcessor struct{}

// The API Lambda enqueues the job as soon as it hands back the presigned
// upload URL, before the client has actually PUT the object to S3. A short
// bounded retry absorbs that race instead of failing the job and waiting out
// the SQS visibility timeout for redelivery.
var objectNotFoundRetryDelays = []time.Duration{200 * time.Millisecond, 400 * time.Millisecond, 800 * time.Millisecond}

func (InspectProcessor) Run(ctx context.Context, s3Client S3GetObjectAPI, job models.Job) (Result, error) {
	out, err := getObjectWithRetry(ctx, s3Client, job)
	if err != nil {
		return Result{}, err
	}
	defer out.Body.Close()

	sniff := make([]byte, 512)
	n, err := io.ReadFull(out.Body, sniff)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return Result{}, err
	}
	sniff = sniff[:n]

	hasher := sha256.New()
	hasher.Write(sniff)
	size := int64(n)

	written, err := io.Copy(hasher, out.Body)
	if err != nil {
		return Result{}, err
	}
	size += written

	return Result{
		Size:        size,
		SHA256:      hex.EncodeToString(hasher.Sum(nil)),
		ContentType: http.DetectContentType(sniff),
	}, nil
}

func getObjectWithRetry(ctx context.Context, s3Client S3GetObjectAPI, job models.Job) (*s3.GetObjectOutput, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(job.Bucket),
		Key:    aws.String(job.Key),
	}

	for attempt := 0; ; attempt++ {
		out, err := s3Client.GetObject(ctx, input)
		if err == nil {
			return out, nil
		}

		var notFound *types.NoSuchKey
		if !errors.As(err, &notFound) || attempt >= len(objectNotFoundRetryDelays) {
			return nil, err
		}

		select {
		case <-time.After(objectNotFoundRetryDelays[attempt]):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}
