import json

import files
import validation


def handler(event, context):
    request_id = (
        event.get("requestContext", {})
        .get("http", {})
        .get("requestId", "")
    )
    method = event["requestContext"]["http"]["method"]
    path = event.get("rawPath", "")
    path_params = event.get("pathParameters") or {}

    try:
        if method == "POST" and path == "/files/upload-url":
            return _create_upload(event, request_id)
        if method == "GET" and path == "/files":
            return _response(200, files.list_files(), request_id)
        if method == "GET" and "id" in path_params:
            return _get_file(path_params["id"], request_id)
        if method == "DELETE" and "id" in path_params:
            return _delete_file(path_params["id"], request_id)
        if method == "POST" and path == "/files/cleanup":
            import expiry

            return _response(200, {"cleaned": expiry.cleanup_expired()}, request_id)
    except ValueError as error:
        return _error_response(400, "VALIDATION_FAILED", str(error), request_id)

    return _error_response(404, "NOT_FOUND", "not found", request_id)


def _create_upload(event, request_id):
    body = _parse_body(event)
    validation.validate_upload_request(body.get("filename"), body.get("content_type"))
    return _response(201, files.create_upload(body["filename"], body["content_type"]), request_id)


def _get_file(file_id: str, request_id):
    result = files.get_file(file_id)
    if result is None:
        return _error_response(404, "NOT_FOUND", "file not found", request_id)
    return _response(200, result, request_id)


def _delete_file(file_id: str, request_id):
    if not files.delete_file(file_id):
        return _error_response(404, "NOT_FOUND", "file not found", request_id)
    return _response(204, None, request_id)


def _parse_body(event) -> dict:
    body = event.get("body") or "{}"
    try:
        return json.loads(body)
    except json.JSONDecodeError as error:
        raise ValueError("invalid JSON body") from error


def _response(status_code: int, payload, request_id=""):
    body = payload if payload is not None else ""
    return {
        "statusCode": status_code,
        "headers": {"Content-Type": "application/json"},
        "body": json.dumps({"data": body, "request_id": request_id}) if body else json.dumps({"request_id": request_id}),
    }


def _error_response(status_code: int, code: str, message: str, request_id=""):
    return {
        "statusCode": status_code,
        "headers": {"Content-Type": "application/json"},
        "body": json.dumps({"error": {"code": code, "message": message}, "request_id": request_id}),
    }
