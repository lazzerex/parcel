package awsconfig

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

const defaultRegion = "us-east-1"

// Load builds the AWS config for every client the worker creates. Leaving
// AWS_ENDPOINT_URL unset restores normal endpoint resolution, which is the
// whole switch from Floci to real AWS.
func Load(ctx context.Context) (aws.Config, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(Region()),
	}

	if endpoint := os.Getenv("AWS_ENDPOINT_URL"); endpoint != "" {
		opts = append(opts, config.WithBaseEndpoint(endpoint))
	}

	return config.LoadDefaultConfig(ctx, opts...)
}

func Region() string {
	if region := os.Getenv("AWS_DEFAULT_REGION"); region != "" {
		return region
	}
	return defaultRegion
}
