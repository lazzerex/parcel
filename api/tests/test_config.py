import config


def test_no_hardcoded_defaults(monkeypatch):
    monkeypatch.delenv("AWS_DEFAULT_REGION", raising=False)
    monkeypatch.delenv("AWS_ENDPOINT_URL", raising=False)
    monkeypatch.delenv("DYNAMODB_TABLE", raising=False)
    monkeypatch.delenv("S3_BUCKET", raising=False)

    current = config.settings()

    assert current.region is None
    assert current.endpoint_url is None
    assert current.dynamodb_table is None
    assert current.s3_bucket is None


def test_reads_resource_names_from_environment(monkeypatch):
    monkeypatch.setenv("DYNAMODB_TABLE", "parcel-metadata")
    monkeypatch.setenv("S3_BUCKET", "parcel-files")

    current = config.settings()

    assert current.dynamodb_table == "parcel-metadata"
    assert current.s3_bucket == "parcel-files"


def test_reads_endpoint_from_environment(monkeypatch):
    monkeypatch.setenv("AWS_ENDPOINT_URL", "http://localhost:4566")
    monkeypatch.setenv("AWS_DEFAULT_REGION", "eu-west-1")

    current = config.settings()

    assert current.region == "eu-west-1"
    assert current.endpoint_url == "http://localhost:4566"


def test_empty_endpoint_is_treated_as_unset(monkeypatch):
    monkeypatch.setenv("AWS_ENDPOINT_URL", "")

    assert config.settings().endpoint_url is None


def test_empty_region_is_treated_as_unset(monkeypatch):
    monkeypatch.setenv("AWS_DEFAULT_REGION", "")

    assert config.settings().region is None


def test_client_targets_configured_endpoint(monkeypatch):
    monkeypatch.setenv("AWS_ENDPOINT_URL", "http://localhost:4566")
    monkeypatch.setenv("AWS_DEFAULT_REGION", "us-east-1")
    monkeypatch.setenv("AWS_ACCESS_KEY_ID", "test")
    monkeypatch.setenv("AWS_SECRET_ACCESS_KEY", "test")

    s3 = config.client("s3")

    assert s3.meta.endpoint_url == "http://localhost:4566"
    assert s3.meta.region_name == "us-east-1"
