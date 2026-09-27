import pytest

import validation


def test_valid_request_passes():
    validation.validate_upload_request("photo.jpg", "image/jpeg")


def test_missing_filename_raises():
    with pytest.raises(ValueError, match="filename and content_type are required"):
        validation.validate_upload_request(None, "image/jpeg")


def test_missing_content_type_raises():
    with pytest.raises(ValueError, match="filename and content_type are required"):
        validation.validate_upload_request("photo.jpg", None)


def test_empty_filename_raises():
    with pytest.raises(ValueError, match="filename and content_type are required"):
        validation.validate_upload_request("", "image/jpeg")


def test_filename_too_long_raises():
    name = "a" * 256
    with pytest.raises(ValueError, match="at most 255"):
        validation.validate_upload_request(name, "image/jpeg")


def test_filename_with_slash_raises():
    with pytest.raises(ValueError, match="path separators"):
        validation.validate_upload_request("dir/photo.jpg", "image/jpeg")


def test_filename_with_backslash_raises():
    with pytest.raises(ValueError, match="path separators"):
        validation.validate_upload_request("dir\\photo.jpg", "image/jpeg")


def test_filename_with_null_byte_raises():
    with pytest.raises(ValueError, match="path separators"):
        validation.validate_upload_request("photo\x00.jpg", "image/jpeg")


def test_disallowed_content_type_raises():
    with pytest.raises(ValueError, match="not allowed"):
        validation.validate_upload_request("file.exe", "application/x-executable")


def test_allowed_content_types():
    for ct in validation.ALLOWED_CONTENT_TYPES:
        validation.validate_upload_request("file.bin", ct)
