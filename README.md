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
  config.py           AWS client construction
  log.py              structured JSON logging
  tests/
worker/               Go processing worker
  cmd/worker/         entrypoint
  internal/awsconfig/ AWS config loading
terraform/            infrastructure as code
  s3.tf               file bucket
  dynamodb.tf         metadata table
  sqs.tf              job queue and dead-letter queue
  iam.tf              per-role scoped policies
scripts/
  verify-infra.sh     asserts live infrastructure matches the config
docker-compose.yml    Floci
```

Directories for handlers, services, processors, and queues appear as the work
that needs them lands, rather than being scaffolded empty up front.

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

## Status

Parcel is a work in progress, built in order rather than all at once.

**Working.** The local environment is reproducible from a clean clone. Both
projects build, test, and reach Floci through a single centralized point of AWS
configuration. Terraform provisions the data plane end to end: the S3 bucket
with public access blocked, the DynamoDB metadata table, the job queue and its
dead-letter queue, and a separate least-privilege IAM role for the API and for
the worker. `scripts/verify-infra.sh` asserts all of it against a live Floci,
and the stack survives both a container restart and a full destroy and rebuild.

**Not built yet.** The Lambda functions and the API Gateway in the diagram
above. No handler code exists to deploy into them, so provisioning them now
would only mean stub functions rewritten twice. They land alongside the code
they run: the Python API first, then the Go processing worker, then the
reliability work around retries, duplicate deliveries, and the dead-letter
queue.
