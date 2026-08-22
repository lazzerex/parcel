import config
import storage


def _client(monkeypatch):
    monkeypatch.setenv("AWS_ENDPOINT_URL", "http://localhost:4566")
    monkeypatch.setenv("AWS_DEFAULT_REGION", "us-east-1")
    monkeypatch.setenv("AWS_ACCESS_KEY_ID", "test")
    monkeypatch.setenv("AWS_SECRET_ACCESS_KEY", "test")
    return config.client("s3")


def test_presigned_upload_url_targets_bucket_and_key(monkeypatch):
    client = _client(monkeypatch)

    url = storage.presigned_upload_url(
        client, "parcel-files", "uploads/abc123/photo.jpg", "image/jpeg"
    )

    assert url.startswith("http://localhost:4566/parcel-files/uploads/abc123/photo.jpg")
    assert "Expires=" in url


def test_presigned_download_url_targets_bucket_and_key(monkeypatch):
    client = _client(monkeypatch)

    url = storage.presigned_download_url(client, "parcel-files", "processed/abc123/photo.jpg")

    assert url.startswith("http://localhost:4566/parcel-files/processed/abc123/photo.jpg")
