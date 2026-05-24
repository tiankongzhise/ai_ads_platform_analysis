from __future__ import annotations

import uuid
from dataclasses import asdict, dataclass, field
from datetime import datetime, timedelta, timezone
from typing import Any


def utc_now() -> datetime:
    return datetime.now(timezone.utc)


def iso(value: datetime) -> str:
    return value.isoformat().replace("+00:00", "Z")


@dataclass
class VirtualAccount:
    id: str
    tenant_id: str
    owner_user_id: str
    display_name: str
    purpose: str
    status: str
    valid_days: int
    created_at: str
    expires_at: str
    destroyed_at: str = ""
    migration_authorized_at: str = ""
    migration_authorized_by: str = ""
    migration_target_user_id: str = ""
    cleanup_verified: bool = False
    resources: dict[str, list[str]] = field(default_factory=dict)

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


@dataclass
class MigrationAuthorization:
    id: str
    virtual_account_id: str
    tenant_id: str
    target_user_id: str
    authorized_by: str
    reason: str
    status: str
    created_at: str
    expires_at: str

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


@dataclass
class CleanupLog:
    id: str
    virtual_account_id: str
    tenant_id: str
    reason: str
    status: str
    actor: str
    token_refs_deleted: int
    files_deleted: int
    cache_keys_deleted: int
    queue_messages_deleted: int
    business_records_found: int
    no_business_data: bool
    created_at: str
    details: dict[str, Any] = field(default_factory=dict)

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


class VirtualAccountService:
    def __init__(self) -> None:
        self.accounts: dict[str, VirtualAccount] = {}
        self.authorizations: list[MigrationAuthorization] = []
        self.cleanup_logs: list[CleanupLog] = []

    def list_accounts(self, tenant_id: str = "") -> list[dict[str, Any]]:
        rows = list(self.accounts.values())
        if tenant_id:
            rows = [row for row in rows if row.tenant_id == tenant_id]
        return [row.to_dict() for row in sorted(rows, key=lambda item: item.created_at, reverse=True)]

    def get_account(self, account_id: str) -> dict[str, Any]:
        return self.must_account(account_id).to_dict()

    def create_account(self, payload: dict[str, Any], now: datetime | None = None) -> dict[str, Any]:
        now = now or utc_now()
        valid_days = clamp_valid_days(payload.get("valid_days", 7))
        account_id = str(uuid.uuid4())
        account = VirtualAccount(
            id=account_id,
            tenant_id=str(payload.get("tenant_id", "demo-tenant")),
            owner_user_id=str(payload.get("owner_user_id", "demo-user")),
            display_name=str(payload.get("display_name", "7 天虚拟账户")),
            purpose=str(payload.get("purpose", "短期投放联调")),
            status="active",
            valid_days=valid_days,
            created_at=iso(now),
            expires_at=iso(now + timedelta(days=valid_days)),
            resources=resource_manifest(account_id, payload),
        )
        self.accounts[account_id] = account
        return account.to_dict()

    def authorize_migration(self, account_id: str, payload: dict[str, Any], now: datetime | None = None) -> dict[str, Any]:
        now = now or utc_now()
        account = self.must_account(account_id)
        if account.status in {"destroyed", "cleanup_blocked"}:
            raise ValueError("account is not migratable")
        authorization = MigrationAuthorization(
            id=str(uuid.uuid4()),
            virtual_account_id=account.id,
            tenant_id=account.tenant_id,
            target_user_id=str(payload.get("target_user_id", "real-user")),
            authorized_by=str(payload.get("authorized_by", "system")),
            reason=str(payload.get("reason", "迁移虚拟账户配置")),
            status="authorized",
            created_at=iso(now),
            expires_at=iso(now + timedelta(hours=24)),
        )
        account.status = "migration_authorized"
        account.migration_authorized_at = authorization.created_at
        account.migration_authorized_by = authorization.authorized_by
        account.migration_target_user_id = authorization.target_user_id
        self.authorizations.insert(0, authorization)
        return {"account": account.to_dict(), "authorization": authorization.to_dict()}

    def destroy_account(self, account_id: str, payload: dict[str, Any] | None = None, now: datetime | None = None) -> dict[str, Any]:
        payload = payload or {}
        account = self.must_account(account_id)
        return self.cleanup_account(account, str(payload.get("reason", "manual_destroy")), str(payload.get("actor", "system")), now)

    def cleanup_expired(self, payload: dict[str, Any] | None = None, now: datetime | None = None) -> dict[str, Any]:
        payload = payload or {}
        now = now or utc_now()
        actor = str(payload.get("actor", "system-cleaner"))
        logs = []
        for account in list(self.accounts.values()):
            if account.status in {"destroyed", "cleanup_blocked"}:
                continue
            if parse_iso(account.expires_at) <= now:
                logs.append(self.cleanup_account(account, "expired_cleanup", actor, now)["cleanup_log"])
        return {"cleaned": len(logs), "logs": logs}

    def list_cleanup_logs(self, tenant_id: str = "") -> list[dict[str, Any]]:
        rows = self.cleanup_logs
        if tenant_id:
            rows = [row for row in rows if row.tenant_id == tenant_id]
        return [row.to_dict() for row in rows]

    def cleanup_account(self, account: VirtualAccount, reason: str, actor: str, now: datetime | None = None) -> dict[str, Any]:
        now = now or utc_now()
        resources = account.resources
        business_records = resources.get("business_records", [])
        no_business_data = len(business_records) == 0
        log = CleanupLog(
            id=str(uuid.uuid4()),
            virtual_account_id=account.id,
            tenant_id=account.tenant_id,
            reason=reason,
            status="success" if no_business_data else "blocked",
            actor=actor,
            token_refs_deleted=len(resources.get("token_refs", [])) if no_business_data else 0,
            files_deleted=len(resources.get("file_refs", [])) if no_business_data else 0,
            cache_keys_deleted=len(resources.get("cache_keys", [])) if no_business_data else 0,
            queue_messages_deleted=len(resources.get("queue_messages", [])) if no_business_data else 0,
            business_records_found=len(business_records),
            no_business_data=no_business_data,
            created_at=iso(now),
            details={
                "token_refs_checked": len(resources.get("token_refs", [])),
                "file_refs_checked": len(resources.get("file_refs", [])),
                "cache_keys_checked": len(resources.get("cache_keys", [])),
                "queue_messages_checked": len(resources.get("queue_messages", [])),
            },
        )
        if no_business_data:
            account.resources = {
                "token_refs": [],
                "file_refs": [],
                "cache_keys": [],
                "queue_messages": [],
                "business_records": [],
            }
            account.status = "destroyed"
            account.destroyed_at = log.created_at
            account.cleanup_verified = True
        else:
            account.status = "cleanup_blocked"
            account.cleanup_verified = False
        self.cleanup_logs.insert(0, log)
        return {"account": account.to_dict(), "cleanup_log": log.to_dict()}

    def must_account(self, account_id: str) -> VirtualAccount:
        account = self.accounts.get(account_id)
        if not account:
            raise KeyError("virtual account not found")
        return account


def clamp_valid_days(value: Any) -> int:
    try:
        days = int(value)
    except (TypeError, ValueError):
        days = 7
    return max(1, min(days, 7))


def resource_manifest(account_id: str, payload: dict[str, Any]) -> dict[str, list[str]]:
    return {
        "token_refs": list(payload.get("token_refs") or [f"token:{account_id}:access", f"token:{account_id}:refresh"]),
        "file_refs": list(payload.get("file_refs") or [f"minio://virtual-account/{account_id}/import-preview.csv"]),
        "cache_keys": list(payload.get("cache_keys") or [f"virtual-account:{account_id}:profile"]),
        "queue_messages": list(payload.get("queue_messages") or [f"queue:ad-sync:{account_id}"]),
        "business_records": list(payload.get("business_records") or []),
    }


def parse_iso(value: str) -> datetime:
    return datetime.fromisoformat(value.replace("Z", "+00:00"))
