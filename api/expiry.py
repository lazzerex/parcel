import time

import config
import metadata
from models import FileMetadata


def cleanup_expired() -> int:
    settings = config.settings()
    dynamodb = config.client("dynamodb")
    s3 = config.client("s3")
    now = int(time.time())
    cleaned = 0

    items = dynamodb.scan(
        TableName=settings.dynamodb_table,
        FilterExpression="attribute_exists(expires_at)",
    ).get("Items", [])

    for item in items:
        file_metadata = FileMetadata.from_item(item)
        if file_metadata.expires_at is not None and file_metadata.expires_at < now:
            s3.delete_object(Bucket=settings.s3_bucket, Key=file_metadata.s3_key)
            metadata.delete(dynamodb, settings.dynamodb_table, file_metadata.id)
            cleaned += 1

    return cleaned