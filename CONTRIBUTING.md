# Contributing to Parcel

Thank you for your interest in contributing to Parcel.

Parcel is an event-driven file processing platform built as a hands-on learning project for AWS architecture. Python orchestrates, Go processes, and Terraform provisions, all running entirely on a local AWS emulator.

Contributions are welcome, whether they involve bug fixes, documentation, performance improvements, testing, or new processors.

## Before You Start

Before making changes:

1. Read the [README](README.md) and [ARCHITECTURE.md](ARCHITECTURE.md) to understand the project structure and language responsibilities.
2. Search existing issues and pull requests to avoid duplicate work.
3. For significant feature changes, open an issue first to discuss the proposed approach.
4. Keep changes focused and avoid unrelated modifications.

## Prerequisites

- Python 3.12
- Go 1.27
- Terraform 1.15+
- Docker with Docker Compose
- Floci 1.7.0+

## Getting Started

Clone and set up the project:

```bash
git clone https://github.com/lazzerex/parcel.git
cd parcel
```

### Python API

```bash
cd api
python3 -m venv .venv
.venv/bin/pip install -r requirements-dev.txt
```

### Go Worker

```bash
cd worker
go mod download
```

### Infrastructure

```bash
docker compose up -d
eval $(floci env)
cd terraform
terraform init
terraform apply -auto-approve
```

## Development Workflow

1. Fork the repository if you are working from an external fork.
2. Clone your fork locally.
3. Create a dedicated branch for your changes.
4. Implement and test your changes.
5. Run the relevant formatting, linting, and test checks.
6. Commit your changes using a clear commit message.
7. Push your branch and open a pull request.

Example:

```bash
git checkout -b feat/add-retry-logic
```

## Code Quality

Before submitting a pull request, run the following checks where applicable.

### Python Formatting and Linting

```bash
cd api
.venv/bin/python -m pytest -q
```

### Go Formatting and Linting

```bash
cd worker
go vet ./...
go test ./... -count=1 -short
```

### Terraform Formatting

```bash
terraform -chdir=terraform fmt -check
```

### Infrastructure Verification

Requires running Floci and Terraform:

```bash
./scripts/verify-infra.sh
```

Parcel's CI workflow runs pytest, go vet, go test, go build, and terraform fmt. Please ensure your changes pass the relevant checks before opening a pull request.

## Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` New functionality
- `fix:` Bug fixes
- `test:` Test changes
- `docs:` Documentation changes
- `chore:` Tooling, dependencies, CI
- `refactor:` Code restructuring without behavior change

Examples:

```
feat: add retry logic for SQS job processing
fix: handle empty file upload gracefully
docs: update architecture diagram
test: cover DynamoDB conditional writes
```

## Pull Requests

Before opening a pull request:

- Ensure the change has a clear purpose.
- Keep the pull request focused on a single concern.
- Update documentation when behavior or usage changes.
- Add or update tests where applicable.
- Explain important implementation decisions.
- Mention any limitations or environment-specific requirements.

## Bug Reports

When reporting a bug, include:

- Operating system and version.
- Python version, Go version, Terraform version.
- Floci version or Docker image tag.
- Steps to reproduce the issue.
- Expected behavior.
- Actual behavior, including error messages or logs.

Remove any sensitive information before sharing logs or system details.

## Feature Requests

Feature requests are welcome. Please explain:

- The problem the feature would solve.
- The proposed behavior.
- Why the feature fits Parcel's scope.

For substantial changes, discuss the design in an issue before implementation.

## Versioning

Parcel follows [SemVer](https://semver.org/). The `VERSION` file at the repo root is the source of truth. Releases are created through the `release.yml` GitHub Actions workflow.

## License

By contributing to Parcel, you agree that your contributions will be licensed under the same license as the project.
