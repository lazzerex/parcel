package awsconfig

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// Load builds the AWS config for every client the worker creates. Region and
// credentials resolve through the SDK's own environment chain, and leaving
// AWS_ENDPOINT_URL unset restores normal endpoint resolution, which is the
// whole switch from the emulator to real AWS.
func Load(ctx context.Context) (aws.Config, error) {
	var opts []func(*config.LoadOptions) error

	if endpoint := os.Getenv("AWS_ENDPOINT_URL"); endpoint != "" {
		opts = append(opts, config.WithBaseEndpoint(endpoint))
	}

	return config.LoadDefaultConfig(ctx, opts...)
}
