import json

from botocore.stub import ANY, Stubber

import config
import handler
from models import FileMetadata


def _configure_env(monkeypatch):
    monkeypatch.setenv("AWS_ENDPOINT_URL", "http://localhost:4566")
    monkeypatch.setenv("AWS_DEFAULT_REGION", "us-east-1")
    monkeypatch.setenv("AWS_ACCESS_KEY_ID", "test")
    monkeypatch.setenv("AWS_SECRET_ACCESS_KEY", "test")
    monkeypatch.setenv("DYNAMODB_TABLE", "parcel-metadata")
    monkeypatch.setenv("S3_BUCKET", "parcel-files")
    monkeypatch.setenv("SQS_QUEUE_URL", "http://localhost:4566/000000000000/parcel-jobs")


def _configure_clients(monkeypatch):
    dynamodb_client = config.client("dynamodb")
    s3_client = config.client("s3")
    sqs_client = config.client("sqs")
    clients = {"dynamodb": dynamodb_client, "s3": s3_client, "sqs": sqs_client}
    monkeypatch.setattr(config, "client", lambda service: clients[service])
    return clients


def _api_event(method, path, body=None, path_params=None):
    event = {
        "requestContext": {"http": {"method": method}},
        "rawPath": path,
    }
    if body is not None:
        event["body"] = body
    if path_params is not None:
        event["pathParameters"] = path_params
    return event


def _make_item(file_id, filename="photo.jpg", status="PENDING", sha256=None, expires_at=None):
    return FileMetadata(
        id=file_id, filename=filename, s3_key=f"uploads/{file_id}/{filename}",
        content_type="image/jpeg", size=1024, sha256=sha256, status=status,
        created_at="2026-01-01T00:00:00+00:00", updated_at="2026-01-01T00:00:00+00:00",
        processed_at=None, expires_at=expires_at,
    )


def test_full_upload_lifecycle(monkeypatch):
    _configure_env(monkeypatch)
    clients = _configure_clients(monkeypatch)

    dynamodb_stub = Stubber(clients["dynamodb"])
    dynamodb_stub.add_response("put_item", {}, {"TableName": "parcel-metadata", "Item": ANY, "ConditionExpression": "attribute_not_exists(PK)"})
    sqs_stub = Stubber(clients["sqs"])
    sqs_stub.add_response("send_message", {}, {"QueueUrl": "http://localhost:4566/000000000000/parcel-jobs", "MessageBody": ANY})
    dynamodb_stub.add_response("update_item", {}, {"TableName": "parcel-metadata", "Key": ANY, "UpdateExpression": "SET #status = :status, updated_at = :updated_at", "ExpressionAttributeNames": {"#status": "status"}, "ExpressionAttributeValues": {":status": {"S": "QUEUED"}, ":updated_at": {"S": ANY}}, "ConditionExpression": "attribute_exists(PK)"})

    with dynamodb_stub, sqs_stub:
        event = _api_event("POST", "/files/upload-url", body=json.dumps({"filename": "photo.jpg", "content_type": "image/jpeg"}))
        response = handler.handler(event, None)

    assert response["statusCode"] == 201
    body = json.loads(response["body"])
    file_id = body["data"]["file_id"]
    assert body["data"]["s3_key"].startswith("uploads/")
    assert body["data"]["s3_key"].endswith("/photo.jpg")
    assert "upload_url" in body["data"]


def test_get_then_delete_flow(monkeypatch):
    _configure_env(monkeypatch)
    clients = _configure_clients(monkeypatch)
    item = _make_item("abc123", status="COMPLETED", sha256="deadbeef")

    dynamodb_stub = Stubber(clients["dynamodb"])
    dynamodb_stub.add_response("get_item", {"Item": item.to_item()}, {"TableName": "parcel-metadata", "Key": {"PK": {"S": "FILE#abc123"}}})

    with dynamodb_stub:
        response = handler.handler(_api_event("GET", "/files/abc123", path_params={"id": "abc123"}), None)

    assert response["statusCode"] == 200
    data = json.loads(response["body"])["data"]
    assert data["id"] == "abc123"
    assert data["status"] == "COMPLETED"
    assert data["sha256"] == "deadbeef"

    dynamodb_stub2 = Stubber(clients["dynamodb"])
    dynamodb_stub2.add_response("get_item", {"Item": item.to_item()}, {"TableName": "parcel-metadata", "Key": {"PK": {"S": "FILE#abc123"}}})
    s3_stub = Stubber(clients["s3"])
    s3_stub.add_response("delete_object", {}, {"Bucket": "parcel-files", "Key": "uploads/abc123/photo.jpg"})
    dynamodb_stub2.add_response("delete_item", {}, {"TableName": "parcel-metadata", "Key": {"PK": {"S": "FILE#abc123"}}})

    with dynamodb_stub2, s3_stub:
        response = handler.handler(_api_event("DELETE", "/files/abc123", path_params={"id": "abc123"}), None)

    assert response["statusCode"] == 204


def test_list_files_flow(monkeypatch):
    _configure_env(monkeypatch)
    clients = _configure_clients(monkeypatch)

    item1 = _make_item("f1", filename="a.jpg", status="PENDING")
    item2 = _make_item("f2", filename="b.png", status="QUEUED")

    dynamodb_stub = Stubber(clients["dynamodb"])
    dynamodb_stub.add_response("scan", {"Items": [item1.to_item(), item2.to_item()]}, {"TableName": "parcel-metadata"})

    with dynamodb_stub:
        response = handler.handler(_api_event("GET", "/files"), None)

    assert response["statusCode"] == 200
    data = json.loads(response["body"])["data"]
    assert len(data) == 2
    ids = {f["id"] for f in data}
    assert ids == {"f1", "f2"}


def test_cleanup_endpoint_flow(monkeypatch):
    _configure_env(monkeypatch)
    clients = _configure_clients(monkeypatch)
    expired = _make_item("e1", status="COMPLETED", sha256="aaa", expires_at=999999999)

    dynamodb_stub = Stubber(clients["dynamodb"])
    dynamodb_stub.add_response(
        "scan", {"Items": [expired.to_item()]},
        {"TableName": "parcel-metadata", "FilterExpression": "attribute_exists(expires_at)"},
    )
    dynamodb_stub.add_response("delete_item", {}, {"TableName": "parcel-metadata", "Key": {"PK": {"S": "FILE#e1"}}})
    s3_stub = Stubber(clients["s3"])
    s3_stub.add_response("delete_object", {}, {"Bucket": "parcel-files", "Key": "uploads/e1/photo.jpg"})

    with dynamodb_stub, s3_stub:
        response = handler.handler(_api_event("POST", "/files/cleanup"), None)

    assert response["statusCode"] == 200
    data = json.loads(response["body"])["data"]
    assert data["cleaned"] == 1
