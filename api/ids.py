import uuid


def generate_file_id() -> str:
    return uuid.uuid4().hex


def build_s3_key(file_id: str, filename: str) -> str:
    return f"uploads/{file_id}/{filename}"
