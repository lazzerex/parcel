import os
from dataclasses import dataclass

import boto3

DEFAULT_REGION = "us-east-1"


@dataclass(frozen=True)
class Settings:
    region: str
    endpoint_url: str | None


def settings() -> Settings:
    return Settings(
        region=os.getenv("AWS_DEFAULT_REGION", DEFAULT_REGION),
        endpoint_url=os.getenv("AWS_ENDPOINT_URL") or None,
    )


def client(service: str):
    # endpoint_url=None restores boto3's normal resolution, so unsetting
    # AWS_ENDPOINT_URL is the whole switch from Floci to real AWS.
    current = settings()
    return boto3.client(
        service,
        region_name=current.region,
        endpoint_url=current.endpoint_url,
    )
