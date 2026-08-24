import json

from botocore.stub import Stubber

import config
import jobs

QUEUE_URL = "http://localhost:4566/000000000000/parcel-jobs"


def _client(monkeypatch):
    monkeypatch.setenv("AWS_DEFAULT_REGION", "us-east-1")
    monkeypatch.setenv("AWS_ACCESS_KEY_ID", "test")
    monkeypatch.setenv("AWS_SECRET_ACCESS_KEY", "test")
    return config.client("sqs")


def test_publish_job_sends_expected_message_body(monkeypatch):
    client = _client(monkeypatch)
    stubber = Stubber(client)

    expected_body = json.dumps(
        {
            "job_id": "job1",
            "file_id": "abc123",
            "bucket": "parcel-files",
            "key": "uploads/abc123/photo.jpg",
            "operation": "inspect",
        }
    )
    stubber.add_response(
        "send_message",
        {},
        {"QueueUrl": QUEUE_URL, "MessageBody": expected_body},
    )

    with stubber:
        jobs.publish_job(
            client, QUEUE_URL, "job1", "abc123", "parcel-files", "uploads/abc123/photo.jpg", "inspect"
        )

    stubber.assert_no_pending_responses()
