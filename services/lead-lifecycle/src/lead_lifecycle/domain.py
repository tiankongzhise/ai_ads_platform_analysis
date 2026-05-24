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


@dataclass
class ConflictItem:
    lead_id: str
    team_id: str
    is_primary: bool = False
    conflict_role: str = "candidate"
    created_at: str = field(default_factory=utc_now)

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


@dataclass
class ConflictGroup:
    id: str
    tenant_id: str
    phone_hash: str
    phone_masked: str
    status: str = "open"
    primary_lead_id: str = ""
    conflict_count: int = 0
    items: list[ConflictItem] = field(default_factory=list)
    resolution_meta: dict[str, Any] = field(default_factory=dict)
    created_at: str = field(default_factory=utc_now)
    updated_at: str = field(default_factory=utc_now)
    resolved_at: str = ""

    def to_dict(self) -> dict[str, Any]:
        payload = asdict(self)
        payload["items"] = [item.to_dict() for item in self.items]
        return payload


@dataclass
class Attribution:
    id: str
    tenant_id: str
    lead_id: str
    attribution_rule: str
    owner_team_id: str
    owner_organization_id: str
    evidence: dict[str, Any] = field(default_factory=dict)
    decided_by: str = "system"
    decided_at: str = field(default_factory=utc_now)

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)
