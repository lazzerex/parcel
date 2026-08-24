import json


def publish_job(
    client, queue_url: str, job_id: str, file_id: str, bucket: str, key: str, operation: str
) -> None:
    client.send_message(
        QueueUrl=queue_url,
        MessageBody=json.dumps(
            {
                "job_id": job_id,
                "file_id": file_id,
                "bucket": bucket,
                "key": key,
                "operation": operation,
            }
        ),
    )
