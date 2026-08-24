<h1 align="center">Parcel</h1>

<p align="center">
  <strong>Event-driven file processing on AWS, running entirely on your laptop.</strong><br />
  No cloud account. No bill. Python orchestrates, Go processes, Terraform provisions.
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Python-3.12-3776AB?logo=python&logoColor=white" alt="Python 3.12" />
  <img src="https://img.shields.io/badge/Go-1.27-00ADD8?logo=go&logoColor=white" alt="Go 1.27" />
  <img src="https://img.shields.io/badge/Terraform-1.15-7B42BC?logo=terraform&logoColor=white" alt="Terraform 1.15" />
  <img src="https://img.shields.io/badge/Docker-29.1-2496ED?logo=docker&logoColor=white" alt="Docker 29.1" />
  <img src="https://img.shields.io/badge/Floci-1.7.0-4B9CD3?logo=icloud&logoColor=white" alt="Floci 1.7.0" />
</p>

<p align="center">
  <img src="https://icon.icepanel.io/AWS/svg/Storage/Simple-Storage-Service.svg" height="42" alt="Amazon S3" title="Amazon S3" />
  &nbsp;&nbsp;
  <img src="https://icon.icepanel.io/AWS/svg/Compute/Lambda.svg" height="42" alt="AWS Lambda" title="AWS Lambda" />
  &nbsp;&nbsp;
  <img src="https://icon.icepanel.io/AWS/svg/Database/DynamoDB.svg" height="42" alt="Amazon DynamoDB" title="Amazon DynamoDB" />
  &nbsp;&nbsp;
  <img src="https://icon.icepanel.io/AWS/svg/App-Integration/Simple-Queue-Service.svg" height="42" alt="Amazon SQS" title="Amazon SQS" />
  &nbsp;&nbsp;
  <img src="https://icon.icepanel.io/AWS/svg/App-Integration/API-Gateway.svg" height="42" alt="Amazon API Gateway" title="Amazon API Gateway" />
  &nbsp;&nbsp;
  <img src="https://icon.icepanel.io/AWS/svg/Security-Identity-Compliance/Identity-and-Access-Management.svg" height="42" alt="AWS IAM" title="AWS IAM" />
</p>

<p align="center">
  <sub>S3 · Lambda · DynamoDB · SQS · API Gateway · IAM</sub>
</p>

<p align="center">
  <a href="#architecture">Architecture</a> ·
  <a href="#why-two-languages">Why two languages</a> ·
  <a href="#running-locally">Running locally</a> ·
  <a href="#example-workflow">Example workflow</a> ·
  <a href="#testing">Testing</a> ·
  <a href="#floci-notes">Floci notes</a>
</p>

---

Parcel is a cloud-native, event-driven file processing platform, built as a
hands-on learning project for AWS architecture. A client requests a presigned
URL, uploads straight to S3, and a queued job wakes a Go worker that inspects
the file and writes results back. It is the same shape a real pipeline takes,
small enough to hold in your head.

The whole system runs on a local AWS emulator, so there is no cloud account and
no bill. The goal is understanding how cloud services fit together into a
distributed, asynchronous system, not shipping a production file host.

## Architecture

```text
Client -> API Gateway -> Python Lambda --> DynamoDB (metadata)
                              |
                              +---------> S3 (presigned upload)
                                            |
                                            v
                                           SQS
                                            |
                                            v
                                       Go Worker --> S3 + DynamoDB
```

A file moves through `PENDING -> UPLOADED -> QUEUED -> PROCESSING -> COMPLETED`,
with `FAILED` reachable from `PROCESSING`. State lives in DynamoDB.

## Why two languages

Each language has a deliberate role, and they communicate only through AWS
service contracts, never by calling each other directly.

- **Python** handles lightweight, I/O-bound orchestration: HTTP handlers,
  presigned URL generation, metadata CRUD, publishing SQS jobs.
- **Go** handles the processing worker: streaming S3 objects, SHA-256 hashing,
  MIME detection, and other CPU- or I/O-intensive work.

## Stack

| Component | Technology |
|---|---|
| Local AWS | Floci 1.7.0 (port 4566) |
| Infrastructure | Terraform 1.15+ |
| API | Python 3.12, boto3 |
| Worker | Go 1.27, aws-sdk-go-v2 |
| Services | API Gateway, Lambda, S3, DynamoDB, SQS, IAM |

## Repository layout

```text
api/                  Python API Lambda
  handler.py          API Gateway routing
  files.py            upload-url / list / get / delete
  metadata.py         DynamoDB CRUD
  storage.py          presigned URLs
  jobs.py             SQS job publisher
  ids.py              file ID and S3 key generation
  models.py           FileMetadata
  config.py           AWS client construction
  log.py              structured JSON logging
  tests/
worker/               Go processing worker
  cmd/worker/         entrypoint: decode, dispatch, idempotency, status writes
  internal/awsconfig/ AWS config loading
  internal/models/    Job decoding
  internal/processors/ Processor interface, inspect (SHA-256, size, MIME sniff)
  internal/store/     DynamoDB status/result writes
terraform/            infrastructure as code
  s3.tf               file bucket
  dynamodb.tf         metadata table
  sqs.tf              job queue and dead-letter queue
  lambda.tf           API and worker Lambdas, worker's SQS event source
  api_gateway.tf      HTTP API routes
  iam.tf              per-role scoped policies
scripts/
  verify-infra.sh     asserts live infrastructure matches the config
docker-compose.yml    Floci
```

## Prerequisites

Docker, Go 1.24+, Python 3.12+, Terraform 1.10+, AWS CLI v2, and the
[Floci CLI](https://github.com/floci-io/floci-cli).

## Running locally

Start the emulator:

```bash
docker compose up -d
floci wait
floci doctor
```

Point your shell at it:

```bash
cp .env.example .env
export $(grep -v '^#' .env | xargs)
```

Provision infrastructure:

```bash
cd terraform
cp local.auto.tfvars.example local.auto.tfvars
terraform init
terraform apply
```

`local.auto.tfvars` is gitignored. It sets `local_emulator = true`, which relaxes
the provider checks a local emulator cannot satisfy. Committed defaults stay
safe for real AWS, so that file is required for local work and must never be
committed.

## Example workflow

`aws_apigatewayv2_stage.default.invoke_url` resolves to a real-looking
`*.execute-api.us-east-1.amazonaws.com` hostname that Floci cannot serve — it
DNS-resolves publicly and refuses the connection. Locally, invoke the Lambda
directly instead:

```bash
echo '{"requestContext":{"http":{"method":"POST"}},"rawPath":"/files/upload-url","body":"{\"filename\":\"photo.jpg\",\"content_type\":\"image/jpeg\"}"}' \
  > event.json

aws lambda invoke --function-name parcel-api \
  --payload file://event.json --cli-binary-format raw-in-base64-out out.json
cat out.json
```

```json
{"file_id": "854347d8...", "s3_key": "uploads/854347d8.../photo.jpg", "upload_url": "http://localhost.floci.io:4566/parcel-files/..."}
```

The client then `PUT`s the file straight to `upload_url`. `create_upload`
already published the processing job and moved status to `QUEUED`:

```json
{"job_id": "45081af8...", "file_id": "854347d8...", "bucket": "parcel-files", "key": "uploads/854347d8.../photo.jpg", "operation": "inspect"}
```

The worker Lambda's SQS event source mapping picks it up automatically. Once
it runs, `GET /files/{id}` (invoked the same way, with `pathParameters`)
returns the completed record:

```json
{
  "id": "854347d8...",
  "filename": "photo.jpg",
  "status": "COMPLETED",
  "size": 4821931,
  "content_type": "image/jpeg",
  "sha256": "e666e3c8...",
  "processed_at": "2026-08-24T09:51:37.139287893Z"
}
```

If the job is redelivered (SQS retry, at-least-once delivery), the worker
reads the current status first and skips reprocessing once it is `COMPLETED`.

## Testing

Python:

```bash
cd api
python3 -m venv .venv
.venv/bin/pip install -r requirements-dev.txt
.venv/bin/python -m pytest
```

Go:

```bash
cd worker
go vet ./...
go test ./...
```

Infrastructure, after `terraform apply`:

```bash
./scripts/verify-infra.sh
```

It reads resource names from Terraform outputs and asserts them against live
Floci: the bucket blocks public access, the table has one hash key and no
secondary indexes, the queue redrives to the dead-letter queue after three
receives, neither IAM policy grants `Resource: "*"`, and an object round-trips
through `uploads/`.

## Configuration

Region, credentials, and endpoint come from the standard AWS environment chain
(`AWS_ENDPOINT_URL`, `AWS_DEFAULT_REGION`, `AWS_ACCESS_KEY_ID`,
`AWS_SECRET_ACCESS_KEY`). Nothing is hardcoded and no credential is committed.

`AWS_ENDPOINT_URL` is the single switch: set it and everything targets the local
emulator, unset it and every component resolves real AWS the normal way. The
application code reads it in exactly two places, `api/config.py` and
`worker/internal/awsconfig`, and passes `None` when it is absent so each SDK
falls back to its own resolution. Terraform does not read it at all; the AWS
provider honours the variable natively.

There are deliberately no default values for region or credentials. A missing
one surfaces as an error rather than silently resolving somewhere unintended.

## Floci notes

Behaviour worth knowing before it looks like a bug:

- The state volume must mount at `/app/data`. Floci's entrypoint drops from root
  to an unprivileged user and only chowns that path; a volume mounted elsewhere
  makes the server fail at boot with `AccessDeniedException`.
- SQS is slow. Creating a queue takes roughly 28s and deleting one roughly 43s,
  against ~3s for S3. Terraform applies touching SQS are not hanging.
- Storage defaults to in-memory. `docker-compose.yml` sets `hybrid` so buckets,
  tables, and queues survive a restart.
- `apiKeyRequired` is not enforced, so API Gateway API keys are not a real gate.
- The HTTP API's `invoke_url` output is not reachable from curl locally (see
  [Example workflow](#example-workflow)); invoke the Lambda functions directly
  instead.

## Status

Parcel is a work in progress, built in order rather than all at once.

**Working.** The full workflow in the architecture diagram runs end to end
against a live Floci: `POST /files/upload-url` creates a `PENDING` metadata
record, returns a presigned S3 URL, publishes the processing job to SQS, and
moves status to `QUEUED`. The worker Lambda's SQS event source mapping picks
the job up automatically, checks current status for idempotency, streams the
S3 object, computes SHA-256 and size in one pass, sniffs content type, and
writes `COMPLETED` (or `FAILED` on error) back to DynamoDB. `GET`/`DELETE
/files/{id}` and `GET /files` round out the API. Infrastructure — S3, DynamoDB,
the job queue and its dead-letter queue, both Lambdas with least-privilege
per-role IAM, and the API Gateway HTTP API — is fully provisioned by Terraform
and asserted against live Floci by `scripts/verify-infra.sh`. The stack
survives a container restart and a full destroy/rebuild.

**Not built yet.** Broader reliability hardening (systematic failure-mode
testing, tuned retry behavior beyond the default SQS redrive) and additional
Go processors beyond `inspect`.
