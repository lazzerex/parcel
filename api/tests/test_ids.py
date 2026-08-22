import ids


def test_generate_file_id_is_unique_hex():
    a = ids.generate_file_id()
    b = ids.generate_file_id()

    assert a != b
    assert len(a) == 32
    int(a, 16)


def test_build_s3_key():
    assert ids.build_s3_key("abc123", "photo.jpg") == "uploads/abc123/photo.jpg"
