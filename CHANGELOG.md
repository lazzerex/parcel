# Changelog

All notable changes to Parcel will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

## [v2.0.0] - 2026-09-26

## [1.0.0] - 2026-09-26

First release. The core workflow is complete and verified.

### Added

- Presigned S3 upload URL generation via `POST /files/upload-url`
- File metadata storage in DynamoDB
- SQS job queue for async processing
- Go worker: SHA-256, file size, content-type detection
- `GET /files`, `GET /files/{id}`, `DELETE /files/{id}`
- Idempotent worker processing with status-based dedup
- Full Terraform provisioning: S3, DynamoDB, SQS+DLQ, API Gateway, 2 Lambdas, 2 IAM roles
- Infrastructure verification script (`scripts/verify-infra.sh`)
- CI workflow: pytest, go vet, go test, go build, terraform fmt
- Release workflow with semver bumping and GitHub Releases

[Unreleased]: https://github.com/lazzerex/parcel/compare/v2.0.0...HEAD
[v2.0.0]: https://github.com/lazzerex/parcel/releases/tag/v2.0.0
[1.0.0]: https://github.com/lazzerex/parcel/releases/tag/v1.0.0
