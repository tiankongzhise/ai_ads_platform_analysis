from __future__ import annotations

from dataclasses import asdict, dataclass, field
from datetime import datetime, timezone
from typing import Any


def utc_now() -> str:
    return datetime.now(timezone.utc).isoformat().replace("+00:00", "Z")


@dataclass
class ImportErrorItem:
    row_number: int
    field: str
    message: str
    raw: dict[str, Any] = field(default_factory=dict)

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


@dataclass
class ImportBatch:
    id: str
    tenant_id: str
    organization_id: str
    team_id: str
    channel_id: str
    status: str = "created"
    file_name: str = ""
    total_rows: int = 0
    success_rows: int = 0
    failed_rows: int = 0
    headers: list[str] = field(default_factory=list)
    mapping_suggestion: dict[str, str] = field(default_factory=dict)
    preview_rows: list[dict[str, Any]] = field(default_factory=list)
    errors: list[ImportErrorItem] = field(default_factory=list)
    created_at: str = field(default_factory=utc_now)
    updated_at: str = field(default_factory=utc_now)

    def to_dict(self) -> dict[str, Any]:
        payload = asdict(self)
        payload["errors"] = [item.to_dict() for item in self.errors]
        return payload


@dataclass
class Lead:
    id: str
    tenant_id: str
    organization_id: str
    team_id: str
    channel_id: str
    import_batch_id: str
    system_lead_no: str
    student_name: str
    phone_hash: str
    phone_masked: str
    source_channel: str
    stage: str
    raw: dict[str, Any]
    reported_at: str
    created_at: str = field(default_factory=utc_now)

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)

