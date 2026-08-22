import dataclasses

from models import FileMetadata


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


def test_to_item_uses_composite_pk():
    metadata = _sample()

    assert metadata.to_item()["PK"] == {"S": "FILE#abc123"}


def test_round_trip_without_optional_fields():
    metadata = _sample()

    assert FileMetadata.from_item(metadata.to_item()) == metadata


def test_round_trip_with_optional_fields():
    metadata = dataclasses.replace(
        _sample(),
        sha256="deadbeef",
        processed_at="2026-08-23T00:05:00+00:00",
    )

    assert FileMetadata.from_item(metadata.to_item()) == metadata


def test_to_item_omits_unset_optional_fields():
    metadata = _sample()

    item = metadata.to_item()

    assert "sha256" not in item
    assert "processed_at" not in item


def test_round_trip_without_known_size():
    metadata = dataclasses.replace(_sample(), size=None)

    assert FileMetadata.from_item(metadata.to_item()) == metadata
    assert "size" not in metadata.to_item()
