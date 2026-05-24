from __future__ import annotations

import hashlib
import re
import uuid
from dataclasses import dataclass
from typing import Any

from lead_lifecycle.domain import ImportBatch, ImportErrorItem, Lead, utc_now
from lead_lifecycle.parser import parse_tabular
from lead_lifecycle.repository import MemoryRepository


FIELD_ALIASES = {
    "student_name": ["student_name", "student", "name", "姓名", "学生姓名", "学员姓名", "客户姓名"],
    "phone": ["phone", "mobile", "tel", "手机号", "手机", "联系电话", "电话"],
    "source_channel": ["source_channel", "source", "channel", "渠道", "来源", "线索来源"],
    "stage": ["stage", "status", "阶段", "线索阶段", "跟进阶段"],
    "reported_at": ["reported_at", "created_at", "上报时间", "创建时间", "咨询时间"],
}

DEFAULT_MAPPING = {
    "student_name": "student_name",
    "phone": "phone",
    "source_channel": "source_channel",
    "stage": "stage",
    "reported_at": "reported_at",
}


@dataclass(frozen=True)
class CreateBatchRequest:
    tenant_id: str
    organization_id: str
    team_id: str
    channel_id: str


class LeadImportService:
    def __init__(self, repository: MemoryRepository) -> None:
        self.repository = repository

    def create_batch(self, req: CreateBatchRequest) -> ImportBatch:
        batch = ImportBatch(
            id=new_id(),
            tenant_id=req.tenant_id or "demo-tenant",
            organization_id=req.organization_id or "demo-org",
            team_id=req.team_id or "demo-team",
            channel_id=req.channel_id or "demo-channel",
        )
        return self.repository.save_batch(batch)

    def upload(self, batch_id: str, filename: str, content: bytes) -> ImportBatch:
        batch = self.must_batch(batch_id)
        headers, rows = parse_tabular(filename, content)
        batch.file_name = filename
        batch.headers = headers
        batch.total_rows = len(rows)
        batch.preview_rows = rows[:5]
        batch.mapping_suggestion = suggest_mapping(headers)
        batch.status = "uploaded"
        batch.updated_at = utc_now()
        self.repository.save_staged_rows(batch_id, rows)
        return self.repository.save_batch(batch)

    def confirm(self, batch_id: str, mapping: dict[str, str] | None = None) -> ImportBatch:
        batch = self.must_batch(batch_id)
        rows = self.repository.staged_rows(batch_id)
        if not rows:
            batch.status = "failed"
            batch.errors = [ImportErrorItem(row_number=0, field="file", message="no uploaded rows")]
            batch.updated_at = utc_now()
            return self.repository.save_batch(batch)
        selected_mapping = {**batch.mapping_suggestion, **(mapping or {})}
        errors: list[ImportErrorItem] = []
        success_rows = 0
        for index, raw in enumerate(rows, start=2):
            lead, error = self.build_lead(batch, raw, selected_mapping)
            if error:
                errors.append(ImportErrorItem(row_number=index, field=error[0], message=error[1], raw=raw))
                continue
            assert lead is not None
            self.repository.save_lead(lead)
            success_rows += 1
        batch.success_rows = success_rows
        batch.failed_rows = len(errors)
        batch.errors = errors
        batch.status = "success" if not errors else "partial_success" if success_rows else "failed"
        batch.updated_at = utc_now()
        return self.repository.save_batch(batch)

    def create_lead(self, payload: dict[str, Any]) -> tuple[Lead | None, ImportErrorItem | None]:
        batch = ImportBatch(
            id=payload.get("import_batch_id") or new_id(),
            tenant_id=payload.get("tenant_id") or "demo-tenant",
            organization_id=payload.get("organization_id") or "demo-org",
            team_id=payload.get("team_id") or "demo-team",
            channel_id=payload.get("channel_id") or "demo-channel",
            status="api_submitted",
            total_rows=1,
        )
        raw = {key: stringify(value) for key, value in payload.items()}
        lead, error = self.build_lead(batch, raw, DEFAULT_MAPPING)
        if error:
            return None, ImportErrorItem(row_number=1, field=error[0], message=error[1], raw=raw)
        assert lead is not None
        self.repository.save_batch(batch)
        self.repository.save_lead(lead)
        batch.success_rows = 1
        batch.status = "success"
        batch.updated_at = utc_now()
        self.repository.save_batch(batch)
        return lead, None

    def build_lead(self, batch: ImportBatch, raw: dict[str, str], mapping: dict[str, str]) -> tuple[Lead | None, tuple[str, str] | None]:
        phone = raw_value(raw, mapping.get("phone", ""))
        normalized_phone = normalize_phone(phone)
        if not normalized_phone:
            return None, ("phone", "手机号缺失或格式无效")
        phone_hash = hash_phone(normalized_phone)
        if self.repository.lead_by_team_phone(batch.tenant_id, batch.team_id, phone_hash):
            return None, ("phone", "当前团队已存在相同手机号线索")
        lead = Lead(
            id=new_id(),
            tenant_id=batch.tenant_id,
            organization_id=batch.organization_id,
            team_id=batch.team_id,
            channel_id=batch.channel_id,
            import_batch_id=batch.id,
            system_lead_no=system_lead_no(batch.tenant_id, batch.team_id, phone_hash),
            student_name=raw_value(raw, mapping.get("student_name", "")),
            phone_hash=phone_hash,
            phone_masked=mask_phone(normalized_phone),
            source_channel=raw_value(raw, mapping.get("source_channel", "")) or batch.channel_id,
            stage=raw_value(raw, mapping.get("stage", "")) or "new",
            raw=raw,
            reported_at=raw_value(raw, mapping.get("reported_at", "")) or utc_now(),
        )
        return lead, None

    def must_batch(self, batch_id: str) -> ImportBatch:
        batch = self.repository.batch(batch_id)
        if not batch:
            raise KeyError("batch not found")
        return batch


def suggest_mapping(headers: list[str]) -> dict[str, str]:
    lower_headers = {header.lower(): header for header in headers}
    mapping: dict[str, str] = {}
    for target, aliases in FIELD_ALIASES.items():
        for alias in aliases:
            normalized_alias = alias.lower()
            if normalized_alias in lower_headers:
                mapping[target] = lower_headers[normalized_alias]
                break
        if target not in mapping:
            for header in headers:
                if any(alias.lower() in header.lower() for alias in aliases):
                    mapping[target] = header
                    break
    return mapping


def normalize_phone(phone: str) -> str:
    digits = re.sub(r"\D", "", phone)
    if len(digits) == 13 and digits.startswith("86"):
        digits = digits[2:]
    if len(digits) == 11 and digits.startswith("1"):
        return digits
    return ""


def hash_phone(phone: str) -> str:
    return hashlib.sha256(phone.encode("utf-8")).hexdigest()


def mask_phone(phone: str) -> str:
    return f"{phone[:3]}****{phone[-4:]}" if len(phone) == 11 else ""


def raw_value(raw: dict[str, str], key: str) -> str:
    return stringify(raw.get(key, ""))


def stringify(value: Any) -> str:
    return "" if value is None else str(value).strip()


def system_lead_no(tenant_id: str, team_id: str, phone_hash: str) -> str:
    seed = f"{tenant_id}:{team_id}:{phone_hash}"
    return "LD" + hashlib.sha1(seed.encode("utf-8")).hexdigest()[:16].upper()


def new_id() -> str:
    return str(uuid.uuid4())

