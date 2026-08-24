import os
from dataclasses import dataclass

import boto3


@dataclass(frozen=True)
class Settings:
    region: str | None
    endpoint_url: str | None
    dynamodb_table: str | None
    s3_bucket: str | None
    sqs_queue_url: str | None


def settings() -> Settings:
    return Settings(
        region=os.getenv("AWS_DEFAULT_REGION") or None,
        endpoint_url=os.getenv("AWS_ENDPOINT_URL") or None,
        dynamodb_table=os.getenv("DYNAMODB_TABLE") or None,
        s3_bucket=os.getenv("S3_BUCKET") or None,
        sqs_queue_url=os.getenv("SQS_QUEUE_URL") or None,
    )


def client(service: str):
    # Passing None defers to boto3's own resolution, so unsetting the
    # environment is the whole switch from the emulator to real AWS.
    current = settings()
    return boto3.client(
        service,
        region_name=current.region,
        endpoint_url=current.endpoint_url,
    )
