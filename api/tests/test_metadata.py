from botocore.stub import ANY, Stubber

import config
import metadata
from models import FileMetadata

TABLE = "parcel-metadata"


def _client(monkeypatch):
    monkeypatch.setenv("AWS_DEFAULT_REGION", "us-east-1")
    monkeypatch.setenv("AWS_ACCESS_KEY_ID", "test")
    monkeypatch.setenv("AWS_SECRET_ACCESS_KEY", "test")
    return config.client("dynamodb")


def _sample() -> FileMetadata:
    return FileMetadata(
        id="abc123",
        filename="photo.jpg",
        s3_key="uploads/abc123/photo.jpg",
        content_type="image/jpeg",
        size=1024,
        sha256=None,
        status="PENDING",
        created_at="2026-08-23T00:00:00+00:00",
        updated_at="2026-08-23T00:00:00+00:00",
        processed_at=None,
    )


def test_create_puts_item_with_condition(monkeypatch):
    client = _client(monkeypatch)
    stubber = Stubber(client)
    file_metadata = _sample()
    stubber.add_response(
        "put_item",
        {},
        {
            "TableName": TABLE,
            "Item": file_metadata.to_item(),
            "ConditionExpression": "attribute_not_exists(PK)",
        },
    )

    with stubber:
        metadata.create(client, TABLE, file_metadata)

    stubber.assert_no_pending_responses()


def test_get_returns_none_when_missing(monkeypatch):
    client = _client(monkeypatch)
    stubber = Stubber(client)
    stubber.add_response(
        "get_item",
        {},
        {"TableName": TABLE, "Key": {"PK": {"S": "FILE#abc123"}}},
    )

    with stubber:
        result = metadata.get(client, TABLE, "abc123")

    assert result is None
    stubber.assert_no_pending_responses()


def test_get_returns_metadata_when_present(monkeypatch):
    client = _client(monkeypatch)
    stubber = Stubber(client)
    file_metadata = _sample()
    stubber.add_response(
        "get_item",
        {"Item": file_metadata.to_item()},
        {"TableName": TABLE, "Key": {"PK": {"S": "FILE#abc123"}}},
    )

    with stubber:
        result = metadata.get(client, TABLE, "abc123")

    assert result == file_metadata
    stubber.assert_no_pending_responses()


def test_update_status_sets_status_and_updated_at(monkeypatch):
    client = _client(monkeypatch)
    stubber = Stubber(client)
    stubber.add_response(
        "update_item",
        {},
        {
            "TableName": TABLE,
            "Key": {"PK": {"S": "FILE#abc123"}},
            "UpdateExpression": "SET #status = :status, updated_at = :updated_at",
            "ExpressionAttributeNames": {"#status": "status"},
            "ExpressionAttributeValues": {
                ":status": {"S": "QUEUED"},
                ":updated_at": {"S": ANY},
            },
            "ConditionExpression": "attribute_exists(PK)",
        },
    )

    with stubber:
        metadata.update_status(client, TABLE, "abc123", "QUEUED")

    stubber.assert_no_pending_responses()


def test_delete_calls_delete_item(monkeypatch):
    client = _client(monkeypatch)
    stubber = Stubber(client)
    stubber.add_response(
        "delete_item",
        {},
        {"TableName": TABLE, "Key": {"PK": {"S": "FILE#abc123"}}},
    )

    with stubber:
        metadata.delete(client, TABLE, "abc123")

    stubber.assert_no_pending_responses()
