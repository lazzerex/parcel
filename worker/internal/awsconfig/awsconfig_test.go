package awsconfig

import (
	"context"
	"testing"
)

func TestRegionDefaults(t *testing.T) {
	t.Setenv("AWS_DEFAULT_REGION", "")

	if got := Region(); got != defaultRegion {
		t.Fatalf("Region() = %q, want %q", got, defaultRegion)
	}
}

func TestRegionFromEnvironment(t *testing.T) {
	t.Setenv("AWS_DEFAULT_REGION", "eu-west-1")

	if got := Region(); got != "eu-west-1" {
		t.Fatalf("Region() = %q, want %q", got, "eu-west-1")
	}
}

func TestLoadAppliesEndpoint(t *testing.T) {
	t.Setenv("AWS_ENDPOINT_URL", "http://localhost:4566")
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	cfg, err := Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BaseEndpoint == nil {
		t.Fatal("BaseEndpoint = nil, want the configured endpoint")
	}
	if *cfg.BaseEndpoint != "http://localhost:4566" {
		t.Fatalf("BaseEndpoint = %q, want %q", *cfg.BaseEndpoint, "http://localhost:4566")
	}
}

func TestLoadWithoutEndpointLeavesResolutionToSDK(t *testing.T) {
	t.Setenv("AWS_ENDPOINT_URL", "")
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	cfg, err := Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BaseEndpoint != nil {
		t.Fatalf("BaseEndpoint = %q, want nil", *cfg.BaseEndpoint)
	}
}
