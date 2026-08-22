import json

import files


def handler(event, context):
    method = event["requestContext"]["http"]["method"]
    path = event.get("rawPath", "")
    path_params = event.get("pathParameters") or {}

    try:
        if method == "POST" and path == "/files/upload-url":
            return _create_upload(event)
        if method == "GET" and path == "/files":
            return _response(200, files.list_files())
        if method == "GET" and "id" in path_params:
            return _get_file(path_params["id"])
        if method == "DELETE" and "id" in path_params:
            return _delete_file(path_params["id"])
    except ValueError as error:
        return _response(400, {"error": str(error)})

    return _response(404, {"error": "not found"})


def _create_upload(event):
    body = _parse_body(event)
    filename = body.get("filename")
    content_type = body.get("content_type")
    if not filename or not content_type:
        raise ValueError("filename and content_type are required")

    return _response(201, files.create_upload(filename, content_type))


def _get_file(file_id: str):
    result = files.get_file(file_id)
    if result is None:
        return _response(404, {"error": "file not found"})
    return _response(200, result)


def _delete_file(file_id: str):
    if not files.delete_file(file_id):
        return _response(404, {"error": "file not found"})
    return _response(204, None)


def _parse_body(event) -> dict:
    body = event.get("body") or "{}"
    try:
        return json.loads(body)
    except json.JSONDecodeError as error:
        raise ValueError("invalid JSON body") from error


def _response(status_code: int, payload):
    return {
        "statusCode": status_code,
        "headers": {"Content-Type": "application/json"},
        "body": json.dumps(payload) if payload is not None else "",
    }
