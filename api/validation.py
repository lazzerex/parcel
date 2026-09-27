ALLOWED_CONTENT_TYPES = {
    "image/jpeg",
    "image/png",
    "image/gif",
    "image/webp",
    "application/pdf",
    "text/plain",
    "application/json",
    "text/csv",
}

MAX_FILENAME_LENGTH = 255


def validate_upload_request(filename: str | None, content_type: str | None) -> None:
    if not filename or not content_type:
        raise ValueError("filename and content_type are required")

    if len(filename) > MAX_FILENAME_LENGTH:
        raise ValueError(f"filename must be at most {MAX_FILENAME_LENGTH} characters")

    if "/" in filename or "\\" in filename or "\x00" in filename:
        raise ValueError("filename must not contain path separators or null bytes")

    if content_type not in ALLOWED_CONTENT_TYPES:
        raise ValueError(f"content_type {content_type!r} is not allowed")