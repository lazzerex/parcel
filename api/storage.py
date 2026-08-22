def presigned_upload_url(
    client, bucket: str, key: str, content_type: str, expires_in: int = 900
) -> str:
    return client.generate_presigned_url(
        "put_object",
        Params={"Bucket": bucket, "Key": key, "ContentType": content_type},
        ExpiresIn=expires_in,
    )


def presigned_download_url(client, bucket: str, key: str, expires_in: int = 900) -> str:
    return client.generate_presigned_url(
        "get_object",
        Params={"Bucket": bucket, "Key": key},
        ExpiresIn=expires_in,
    )
