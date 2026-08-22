from botocore.stub import ANY, Stubber

import config
import files


def _configure_env(monkeypatch):
    monkeypatch.setenv("AWS_ENDPOINT_URL", "http://localhost:4566")
    monkeypatch.setenv("AWS_DEFAULT_REGION", "us-east-1")
    monkeypatch.setenv("AWS_ACCESS_KEY_ID", "test")
    monkeypatch.setenv("AWS_SECRET_ACCESS_KEY", "test")
    monkeypatch.setenv("DYNAMODB_TABLE", "parcel-metadata")
    monkeypatch.setenv("S3_BUCKET", "parcel-files")


def test_create_upload_stores_pending_record_and_returns_url(monkeypatch):
    _configure_env(monkeypatch)
    dynamodb_client = config.client("dynamodb")
    s3_client = config.client("s3")
    monkeypatch.setattr(config, "client", lambda service: dynamodb_client if service == "dynamodb" else s3_client)

    dynamodb_stub = Stubber(dynamodb_client)
    dynamodb_stub.add_response(
        "put_item",
        {},
        {
            "TableName": "parcel-metadata",
            "Item": ANY,
            "ConditionExpression": "attribute_not_exists(PK)",
        },
    )

    with dynamodb_stub:
        result = files.create_upload("photo.jpg", "image/jpeg")

    dynamodb_stub.assert_no_pending_responses()
    assert result["s3_key"].startswith("uploads/")
    assert result["s3_key"].endswith("/photo.jpg")
    assert result["upload_url"].startswith("http://localhost:4566/parcel-files/")


def test_delete_file_removes_object_and_record(monkeypatch):
    _configure_env(monkeypatch)
    dynamodb_client = config.client("dynamodb")
    s3_client = config.client("s3")
    monkeypatch.setattr(config, "client", lambda service: dynamodb_client if service == "dynamodb" else s3_client)

    dynamodb_stub = Stubber(dynamodb_client)
    s3_stub = Stubber(s3_client)

    item = {
        "id": {"S": "abc123"},
        "filename": {"S": "photo.jpg"},
        "s3_key": {"S": "uploads/abc123/photo.jpg"},
        "content_type": {"S": "image/jpeg"},
        "status": {"S": "PENDING"},
        "created_at": {"S": "2026-08-23T00:00:00+00:00"},
        "updated_at": {"S": "2026-08-23T00:00:00+00:00"},
    }
    dynamodb_stub.add_response(
        "get_item", {"Item": item}, {"TableName": "parcel-metadata", "Key": {"PK": {"S": "FILE#abc123"}}}
    )
    s3_stub.add_response(
        "delete_object", {}, {"Bucket": "parcel-files", "Key": "uploads/abc123/photo.jpg"}
    )
    dynamodb_stub.add_response(
        "delete_item", {}, {"TableName": "parcel-metadata", "Key": {"PK": {"S": "FILE#abc123"}}}
    )

    with dynamodb_stub, s3_stub:
        assert files.delete_file("abc123") is True

    dynamodb_stub.assert_no_pending_responses()
    s3_stub.assert_no_pending_responses()


def test_delete_file_returns_false_when_missing(monkeypatch):
    _configure_env(monkeypatch)
    dynamodb_client = config.client("dynamodb")
    monkeypatch.setattr(config, "client", lambda service: dynamodb_client)

    dynamodb_stub = Stubber(dynamodb_client)
    dynamodb_stub.add_response(
        "get_item", {}, {"TableName": "parcel-metadata", "Key": {"PK": {"S": "FILE#abc123"}}}
    )

    with dynamodb_stub:
        assert files.delete_file("abc123") is False

    dynamodb_stub.assert_no_pending_responses()
