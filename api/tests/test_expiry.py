import time

from botocore.stub import ANY, Stubber

import config
import expiry
from models import FileMetadata


def _configure(monkeypatch):
    monkeypatch.setenv("AWS_ENDPOINT_URL", "http://localhost:4566")
    monkeypatch.setenv("AWS_DEFAULT_REGION", "us-east-1")
    monkeypatch.setenv("AWS_ACCESS_KEY_ID", "test")
    monkeypatch.setenv("AWS_SECRET_ACCESS_KEY", "test")
    monkeypatch.setenv("DYNAMODB_TABLE", "parcel-metadata")
    monkeypatch.setenv("S3_BUCKET", "parcel-files")
    dynamodb_client = config.client("dynamodb")
    s3_client = config.client("s3")
    clients = {"dynamodb": dynamodb_client, "s3": s3_client}
    monkeypatch.setattr(config, "client", lambda service: clients[service])
    return clients


def _make_item(file_id, s3_key, expires_at):
    return FileMetadata(
        id=file_id,
        filename="photo.jpg",
        s3_key=s3_key,
        content_type="image/jpeg",
        size=1024,
        sha256="abc",
        status="COMPLETED",
        created_at="2026-01-01T00:00:00+00:00",
        updated_at="2026-01-01T00:00:00+00:00",
        processed_at="2026-01-01T00:00:01+00:00",
        expires_at=expires_at,
    )


def test_cleanup_removes_expired_files(monkeypatch):
    clients = _configure(monkeypatch)
    now = int(time.time())
    expired = _make_item("f1", "uploads/f1/photo.jpg", expires_at=now - 100)
    active = _make_item("f2", "uploads/f2/photo.jpg", expires_at=now + 86400)

    dynamodb_stub = Stubber(clients["dynamodb"])
    dynamodb_stub.add_response(
        "scan",
        {"Items": [expired.to_item(), active.to_item()]},
        {"TableName": "parcel-metadata", "FilterExpression": "attribute_exists(expires_at)"},
    )
    dynamodb_stub.add_response("delete_item", {}, {"TableName": "parcel-metadata", "Key": {"PK": {"S": "FILE#f1"}}})

    s3_stub = Stubber(clients["s3"])
    s3_stub.add_response("delete_object", {}, {"Bucket": "parcel-files", "Key": "uploads/f1/photo.jpg"})

    with dynamodb_stub, s3_stub:
        cleaned = expiry.cleanup_expired()

    assert cleaned == 1
    dynamodb_stub.assert_no_pending_responses()
    s3_stub.assert_no_pending_responses()


def test_cleanup_removes_nothing_when_no_expired_files(monkeypatch):
    clients = _configure(monkeypatch)
    now = int(time.time())
    active = _make_item("f2", "uploads/f2/photo.jpg", expires_at=now + 86400)

    dynamodb_stub = Stubber(clients["dynamodb"])
    dynamodb_stub.add_response(
        "scan",
        {"Items": [active.to_item()]},
        {"TableName": "parcel-metadata", "FilterExpression": "attribute_exists(expires_at)"},
    )

    with dynamodb_stub:
        cleaned = expiry.cleanup_expired()

    assert cleaned == 0
    dynamodb_stub.assert_no_pending_responses()


def test_cleanup_handles_empty_scan(monkeypatch):
    clients = _configure(monkeypatch)

    dynamodb_stub = Stubber(clients["dynamodb"])
    dynamodb_stub.add_response(
        "scan",
        {},
        {"TableName": "parcel-metadata", "FilterExpression": "attribute_exists(expires_at)"},
    )

    with dynamodb_stub:
        cleaned = expiry.cleanup_expired()

    assert cleaned == 0
    dynamodb_stub.assert_no_pending_responses()


def test_cleanup_removes_multiple_expired_files(monkeypatch):
    clients = _configure(monkeypatch)
    now = int(time.time())
    expired1 = _make_item("f1", "uploads/f1/photo.jpg", expires_at=now - 200)
    expired2 = _make_item("f2", "uploads/f2/photo.jpg", expires_at=now - 100)

    dynamodb_stub = Stubber(clients["dynamodb"])
    dynamodb_stub.add_response(
        "scan",
        {"Items": [expired1.to_item(), expired2.to_item()]},
        {"TableName": "parcel-metadata", "FilterExpression": "attribute_exists(expires_at)"},
    )
    dynamodb_stub.add_response("delete_item", {}, {"TableName": "parcel-metadata", "Key": {"PK": {"S": "FILE#f1"}}})
    dynamodb_stub.add_response("delete_item", {}, {"TableName": "parcel-metadata", "Key": {"PK": {"S": "FILE#f2"}}})

    s3_stub = Stubber(clients["s3"])
    s3_stub.add_response("delete_object", {}, {"Bucket": "parcel-files", "Key": "uploads/f1/photo.jpg"})
    s3_stub.add_response("delete_object", {}, {"Bucket": "parcel-files", "Key": "uploads/f2/photo.jpg"})

    with dynamodb_stub, s3_stub:
        cleaned = expiry.cleanup_expired()

    assert cleaned == 2
    dynamodb_stub.assert_no_pending_responses()
    s3_stub.assert_no_pending_responses()
