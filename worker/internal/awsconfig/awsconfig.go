package awsconfig

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

func Load(ctx context.Context) (aws.Config, error) {
	var opts []func(*config.LoadOptions) error

	if endpoint := os.Getenv("AWS_ENDPOINT_URL"); endpoint != "" {
		opts = append(opts, config.WithBaseEndpoint(endpoint))
	}

	return config.LoadDefaultConfig(ctx, opts...)
}
