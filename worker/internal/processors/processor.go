package processors

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"

	"parcel/worker/internal/models"
)

type Result struct {
	Size        int64
	SHA256      string
	ContentType string
}

type S3GetObjectAPI interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

type S3PutObjectAPI interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

type Processor interface {
	Run(ctx context.Context, s3Client S3GetObjectAPI, job models.Job) (Result, error)
}

var operations = map[string]Processor{
	"inspect":  InspectProcessor{},
	"validate": ValidateProcessor{},
	"compress": CompressProcessor{},
}

func Dispatch(operation string) (Processor, error) {
	processor, ok := operations[operation]
	if !ok {
		return nil, fmt.Errorf("unknown operation: %q", operation)
	}
	return processor, nil
}
