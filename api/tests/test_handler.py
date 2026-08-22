import json

import files
import handler


def _event(method: str, path: str, body=None, path_params=None) -> dict:
    event = {
        "requestContext": {"http": {"method": method}},
        "rawPath": path,
    }
    if body is not None:
        event["body"] = body
    if path_params is not None:
        event["pathParameters"] = path_params
    return event


def test_create_upload_returns_201(monkeypatch):
    monkeypatch.setattr(
        files,
        "create_upload",
        lambda filename, content_type: {
            "file_id": "abc123",
            "s3_key": "uploads/abc123/photo.jpg",
            "upload_url": "http://x",
        },
    )
    event = _event(
        "POST",
        "/files/upload-url",
        body=json.dumps({"filename": "photo.jpg", "content_type": "image/jpeg"}),
    )

    response = handler.handler(event, None)

    assert response["statusCode"] == 201
    assert json.loads(response["body"])["file_id"] == "abc123"


def test_create_upload_missing_fields_returns_400():
    event = _event("POST", "/files/upload-url", body=json.dumps({}))

    response = handler.handler(event, None)

    assert response["statusCode"] == 400


def test_create_upload_invalid_json_returns_400():
    event = _event("POST", "/files/upload-url", body="not json")

    response = handler.handler(event, None)

    assert response["statusCode"] == 400


def test_list_files_returns_200(monkeypatch):
    monkeypatch.setattr(files, "list_files", lambda: [{"id": "abc123"}])
    event = _event("GET", "/files")

    response = handler.handler(event, None)

    assert response["statusCode"] == 200
    assert json.loads(response["body"]) == [{"id": "abc123"}]


def test_get_file_returns_200(monkeypatch):
    monkeypatch.setattr(files, "get_file", lambda file_id: {"id": file_id})
    event = _event("GET", "/files/abc123", path_params={"id": "abc123"})

    response = handler.handler(event, None)

    assert response["statusCode"] == 200
    assert json.loads(response["body"]) == {"id": "abc123"}


def test_get_file_returns_404_when_missing(monkeypatch):
    monkeypatch.setattr(files, "get_file", lambda file_id: None)
    event = _event("GET", "/files/abc123", path_params={"id": "abc123"})

    response = handler.handler(event, None)

    assert response["statusCode"] == 404


def test_delete_file_returns_204(monkeypatch):
    monkeypatch.setattr(files, "delete_file", lambda file_id: True)
    event = _event("DELETE", "/files/abc123", path_params={"id": "abc123"})

    response = handler.handler(event, None)

    assert response["statusCode"] == 204


def test_delete_file_returns_404_when_missing(monkeypatch):
    monkeypatch.setattr(files, "delete_file", lambda file_id: False)
    event = _event("DELETE", "/files/abc123", path_params={"id": "abc123"})

    response = handler.handler(event, None)

    assert response["statusCode"] == 404


def test_unmatched_route_returns_404():
    event = _event("PATCH", "/files/abc123")

    response = handler.handler(event, None)

    assert response["statusCode"] == 404
