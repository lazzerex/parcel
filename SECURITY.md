# Security Policy

## Supported Versions

Security fixes are considered for the latest version of Parcel.

Older versions may not receive security updates.

## Reporting a Vulnerability

If you discover a potential security vulnerability in Parcel, please report it privately rather than opening a public issue.

Use GitHub's private vulnerability reporting feature if it is available for this repository.

If private reporting is unavailable, contact the repository maintainer through a private communication channel before publicly disclosing the vulnerability.

## What to Include

Please provide as much of the following information as possible:

- A description of the vulnerability.
- The affected component or feature.
- Steps to reproduce the issue.
- A minimal proof of concept, if available.
- The potential impact.
- Any suggested mitigation or fix.

Please avoid including sensitive personal information or confidential system data.

## Scope

Security reports may involve:

- Hardcoded credentials or secrets in the codebase.
- Unsafe input handling in the Python API.
- Unsafe file processing in the Go worker.
- IAM role misconfigurations in Terraform.
- Presigned URL misuse or exposure.
- Dependency-related security issues.
- Vulnerabilities in build or release workflows.

## Disclosure

Please allow reasonable time for the issue to be investigated and addressed before publicly disclosing technical details.

The maintainer may request additional information to reproduce and verify the issue.

## Security Practices

- No credentials, API keys, or secrets are committed to the repository.
- `.env` files are gitignored; only `.env.example` is tracked.
- IAM roles follow least-privilege scoping.
- AWS endpoints are configurable, not hardcoded.

## Limitations

Parcel runs entirely on a local AWS emulator (Floci) and is not deployed to production. It is a learning project and should not be used to handle real user data.

This policy does not guarantee that every report will result in a security fix or that every affected environment can be supported.
