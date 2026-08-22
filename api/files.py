import datetime as dt

import config
import ids
import metadata
import storage
from models import FileMetadata


def create_upload(filename: str, content_type: str) -> dict:
    settings = config.settings()
    file_id = ids.generate_file_id()
    s3_key = ids.build_s3_key(file_id, filename)
    now = dt.datetime.now(dt.timezone.utc).isoformat()

    # size and sha256 are unknown until the Go worker processes the object.
    file_metadata = FileMetadata(
        id=file_id,
        filename=filename,
        s3_key=s3_key,
        content_type=content_type,
        size=None,
        sha256=None,
        status="PENDING",
        created_at=now,
        updated_at=now,
        processed_at=None,
    )

    metadata.create(config.client("dynamodb"), settings.dynamodb_table, file_metadata)
    upload_url = storage.presigned_upload_url(
        config.client("s3"), settings.s3_bucket, s3_key, content_type
    )

    return {"file_id": file_id, "s3_key": s3_key, "upload_url": upload_url}


def list_files() -> list[dict]:
    settings = config.settings()
    response = config.client("dynamodb").scan(TableName=settings.dynamodb_table)
    return [FileMetadata.from_item(item).__dict__ for item in response.get("Items", [])]


def get_file(file_id: str) -> dict | None:
    settings = config.settings()
    file_metadata = metadata.get(config.client("dynamodb"), settings.dynamodb_table, file_id)
    return file_metadata.__dict__ if file_metadata else None


def delete_file(file_id: str) -> bool:
    settings = config.settings()
    dynamodb = config.client("dynamodb")
    file_metadata = metadata.get(dynamodb, settings.dynamodb_table, file_id)
    if file_metadata is None:
        return False

    config.client("s3").delete_object(Bucket=settings.s3_bucket, Key=file_metadata.s3_key)
    metadata.delete(dynamodb, settings.dynamodb_table, file_id)
    return True
