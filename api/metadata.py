import datetime as dt

from models import FileMetadata


def create(client, table_name: str, file_metadata: FileMetadata) -> None:
    client.put_item(
        TableName=table_name,
        Item=file_metadata.to_item(),
        ConditionExpression="attribute_not_exists(PK)",
    )


def get(client, table_name: str, file_id: str) -> FileMetadata | None:
    response = client.get_item(
        TableName=table_name,
        Key={"PK": {"S": f"FILE#{file_id}"}},
    )
    item = response.get("Item")
    return FileMetadata.from_item(item) if item else None


def update_status(client, table_name: str, file_id: str, status: str) -> None:
    now = dt.datetime.now(dt.timezone.utc).isoformat()
    client.update_item(
        TableName=table_name,
        Key={"PK": {"S": f"FILE#{file_id}"}},
        UpdateExpression="SET #status = :status, updated_at = :updated_at",
        ExpressionAttributeNames={"#status": "status"},
        ExpressionAttributeValues={
            ":status": {"S": status},
            ":updated_at": {"S": now},
        },
        ConditionExpression="attribute_exists(PK)",
    )


def delete(client, table_name: str, file_id: str) -> None:
    client.delete_item(
        TableName=table_name,
        Key={"PK": {"S": f"FILE#{file_id}"}},
    )
