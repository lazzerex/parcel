terraform {
  required_version = ">= 1.10"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }
}

# Every AWS endpoint routes through one variable. Setting aws_endpoint_url to ""
# makes the provider fall back to real AWS resolution, which is the whole
# migration path: no resource definition mentions Floci.
provider "aws" {
  region     = var.aws_region
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key

  # Floci issues no real credentials and serves no metadata endpoint.
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true

  # Path-style avoids depending on wildcard DNS for bucket subdomains.
  s3_use_path_style = true

  endpoints {
    apigateway = var.aws_endpoint_url
    dynamodb   = var.aws_endpoint_url
    iam        = var.aws_endpoint_url
    lambda     = var.aws_endpoint_url
    s3         = var.aws_endpoint_url
    sqs        = var.aws_endpoint_url
    sts        = var.aws_endpoint_url
  }
}
