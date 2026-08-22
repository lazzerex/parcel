from dataclasses import dataclass


@dataclass(frozen=True)
class FileMetadata:
    id: str
    filename: str
    s3_key: str
    content_type: str
    size: int | None
    sha256: str | None
    status: str
    created_at: str
    updated_at: str
    processed_at: str | None

    def to_item(self) -> dict:
        item = {
            "PK": {"S": f"FILE#{self.id}"},
            "id": {"S": self.id},
            "filename": {"S": self.filename},
            "s3_key": {"S": self.s3_key},
            "content_type": {"S": self.content_type},
            "status": {"S": self.status},
            "created_at": {"S": self.created_at},
            "updated_at": {"S": self.updated_at},
        }
        if self.size is not None:
            item["size"] = {"N": str(self.size)}
        if self.sha256 is not None:
            item["sha256"] = {"S": self.sha256}
        if self.processed_at is not None:
            item["processed_at"] = {"S": self.processed_at}
        return item

    @staticmethod
    def from_item(item: dict) -> "FileMetadata":
        return FileMetadata(
            id=item["id"]["S"],
            filename=item["filename"]["S"],
            s3_key=item["s3_key"]["S"],
            content_type=item["content_type"]["S"],
            size=int(item["size"]["N"]) if "size" in item else None,
            sha256=item.get("sha256", {}).get("S"),
            status=item["status"]["S"],
            created_at=item["created_at"]["S"],
            updated_at=item["updated_at"]["S"],
            processed_at=item.get("processed_at", {}).get("S"),
        )
