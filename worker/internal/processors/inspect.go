package processors

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"parcel/worker/internal/models"
)

type InspectProcessor struct{}

func (InspectProcessor) Run(ctx context.Context, s3Client S3GetObjectAPI, job models.Job) (Result, error) {
	out, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(job.Bucket),
		Key:    aws.String(job.Key),
	})
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
