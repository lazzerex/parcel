variable "aws_region" {
  description = "AWS region. Any value works against Floci."
  type        = string
  default     = "us-east-1"
}

variable "aws_endpoint_url" {
  description = "Single endpoint for all AWS services. Set to Floci for local development; leave empty to target real AWS."
  type        = string
  default     = "http://localhost:4566"
}

variable "aws_access_key" {
  description = "Access key. Floci accepts any non-empty value."
  type        = string
  default     = "test"
}

variable "aws_secret_key" {
  description = "Secret key. Floci accepts any non-empty value."
  type        = string
  default     = "test"
}
