from __future__ import annotations

import uuid
from collections import defaultdict
from dataclasses import asdict, dataclass, field
from datetime import datetime, timezone
from typing import Any


def utc_now() -> str:
    return datetime.now(timezone.utc).isoformat().replace("+00:00", "Z")


@dataclass
class ETLBatch:
    id: str
    status: str
    source: str
    input_rows: int
    fact_ad_daily_rows: int
    fact_lead_daily_rows: int
    hourly_rows: int
    rule_version: str = "etl-standard-v1"
    metric_version: str = "metric-v1"
    created_at: str = field(default_factory=utc_now)

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


@dataclass
class FactAdDaily:
    tenant_id: str
    report_date: str
    platform: str
    account_id: str
    entity_type: str
    external_entity_id: str
    spend: float
    impressions: int
    clicks: int
    conversions: int
    source_report_type: str
    etl_batch_id: str

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


@dataclass
class FactLeadDaily:
    tenant_id: str
    report_date: str
    organization_id: str
    team_id: str
    channel_text: str
    leads_count: int
    valid_leads_count: int
    visit_count: int
    enroll_count: int
    deal_amount: float
    conflict_count: int
    etl_batch_id: str

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


@dataclass
class FactHourlyAggregate:
    tenant_id: str
    hour_start: str
    organization_id: str
    team_id: str
    platform: str
    channel_id: str
    metrics: dict[str, Any]
    etl_batch_id: str

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


class ETLService:
    def __init__(self) -> None:
        self.batches: list[ETLBatch] = []
        self.fact_ad_daily: list[FactAdDaily] = []
        self.fact_lead_daily: list[FactLeadDaily] = []
        self.fact_hourly: list[FactHourlyAggregate] = []

    def run(self, ad_rows: list[dict[str, Any]] | None = None, lead_rows: list[dict[str, Any]] | None = None) -> ETLBatch:
        ad_rows = ad_rows or demo_ad_rows()
        lead_rows = lead_rows or demo_lead_rows()
        batch_id = str(uuid.uuid4())
        ad_facts = [map_ad_row(row, batch_id) for row in ad_rows]
        lead_facts = aggregate_leads(lead_rows, batch_id)
        hourly_facts = aggregate_hourly(ad_facts, lead_facts, batch_id)
        self.fact_ad_daily = ad_facts
        self.fact_lead_daily = lead_facts
        self.fact_hourly = hourly_facts
        batch = ETLBatch(
            id=batch_id,
            status="success",
            source="inline_payload" if ad_rows or lead_rows else "demo",
            input_rows=len(ad_rows) + len(lead_rows),
            fact_ad_daily_rows=len(ad_facts),
            fact_lead_daily_rows=len(lead_facts),
            hourly_rows=len(hourly_facts),
        )
        self.batches.insert(0, batch)
        return batch

    def latest(self) -> dict[str, Any]:
        return {
            "batches": [item.to_dict() for item in self.batches[:10]],
            "fact_ad_daily": [item.to_dict() for item in self.fact_ad_daily],
            "fact_lead_daily": [item.to_dict() for item in self.fact_lead_daily],
            "fact_hourly_aggregate": [item.to_dict() for item in self.fact_hourly],
        }


def map_ad_row(row: dict[str, Any], batch_id: str) -> FactAdDaily:
    metrics = row.get("metrics") or {}
    raw = row.get("raw") or {}
    spend = number(metrics.get("cost", raw.get("stat_cost", raw.get("cost", raw.get("fee", 0)))))
    impressions = integer(metrics.get("impressions", raw.get("show_cnt", raw.get("impression", raw.get("view_count", 0)))))
    clicks = integer(metrics.get("clicks", raw.get("click_cnt", raw.get("click", raw.get("valid_click_count", 0)))))
    conversions = integer(metrics.get("conversions", raw.get("convert_cnt", raw.get("conversion", raw.get("leads", 0)))))
    return FactAdDaily(
        tenant_id=str(row.get("tenant_id", "demo-tenant")),
        report_date=str(row.get("stat_date", row.get("date", "2026-05-20")))[:10],
        platform=str(row.get("platform", "douyin")),
        account_id=str(row.get("account_id", "")),
        entity_type=str(row.get("entity_type", "account")),
        external_entity_id=str(row.get("external_entity_id", row.get("external_id", ""))),
        spend=round(spend, 2),
        impressions=impressions,
        clicks=clicks,
        conversions=conversions,
        source_report_type=str(row.get("report_type", "account_daily")),
        etl_batch_id=batch_id,
    )


def aggregate_leads(rows: list[dict[str, Any]], batch_id: str) -> list[FactLeadDaily]:
    buckets: dict[tuple[str, str, str, str, str], dict[str, Any]] = defaultdict(lambda: defaultdict(int))
    for row in rows:
        report_date = str(row.get("reported_at", row.get("created_at", "2026-05-20")))[:10]
        key = (
            str(row.get("tenant_id", "demo-tenant")),
            report_date,
            str(row.get("organization_id", "demo-org")),
            str(row.get("team_id", "demo-team")),
            str(row.get("source_channel", row.get("channel_id", "unknown"))),
        )
        bucket = buckets[key]
        bucket["leads_count"] += 1
        stage = str(row.get("stage", "new"))
        if stage not in {"invalid", "duplicate"}:
            bucket["valid_leads_count"] += 1
        if stage in {"visited", "enrolled", "deal"}:
            bucket["visit_count"] += 1
        if stage in {"enrolled", "deal"}:
            bucket["enroll_count"] += 1
        bucket["deal_amount"] += number(row.get("deal_amount", 0))
        if row.get("conflict_role") == "conflict":
            bucket["conflict_count"] += 1
    facts = []
    for (tenant_id, report_date, organization_id, team_id, channel_text), values in buckets.items():
        facts.append(
            FactLeadDaily(
                tenant_id=tenant_id,
                report_date=report_date,
                organization_id=organization_id,
                team_id=team_id,
                channel_text=channel_text,
                leads_count=int(values["leads_count"]),
                valid_leads_count=int(values["valid_leads_count"]),
                visit_count=int(values["visit_count"]),
                enroll_count=int(values["enroll_count"]),
                deal_amount=round(float(values["deal_amount"]), 2),
                conflict_count=int(values["conflict_count"]),
                etl_batch_id=batch_id,
            )
        )
    return facts


def aggregate_hourly(ad_facts: list[FactAdDaily], lead_facts: list[FactLeadDaily], batch_id: str) -> list[FactHourlyAggregate]:
    rows: list[FactHourlyAggregate] = []
    for fact in ad_facts:
        rows.append(
            FactHourlyAggregate(
                tenant_id=fact.tenant_id,
                hour_start=f"{fact.report_date}T00:00:00Z",
                organization_id="",
                team_id="",
                platform=fact.platform,
                channel_id=fact.account_id,
                metrics={
                    "spend": fact.spend,
                    "impressions": fact.impressions,
                    "clicks": fact.clicks,
                    "conversions": fact.conversions,
                },
                etl_batch_id=batch_id,
            )
        )
    for fact in lead_facts:
        rows.append(
            FactHourlyAggregate(
                tenant_id=fact.tenant_id,
                hour_start=f"{fact.report_date}T00:00:00Z",
                organization_id=fact.organization_id,
                team_id=fact.team_id,
                platform="lead",
                channel_id=fact.channel_text,
                metrics={
                    "leads_count": fact.leads_count,
                    "valid_leads_count": fact.valid_leads_count,
                    "visit_count": fact.visit_count,
                    "enroll_count": fact.enroll_count,
                    "conflict_count": fact.conflict_count,
                },
                etl_batch_id=batch_id,
            )
        )
    return rows


def demo_ad_rows() -> list[dict[str, Any]]:
    return [
        {
            "tenant_id": "demo-tenant",
            "platform": "douyin",
            "account_id": "demo_douyin_account",
            "report_type": "account_daily",
            "stat_date": "2026-05-20",
            "entity_type": "account",
            "external_entity_id": "demo_douyin_account",
            "metrics": {"cost": 128.5, "impressions": 6400, "clicks": 320, "conversions": 24},
        }
    ]


def demo_lead_rows() -> list[dict[str, Any]]:
    return [
        {"tenant_id": "demo-tenant", "organization_id": "demo-org", "team_id": "team-a", "source_channel": "douyin", "stage": "new", "reported_at": "2026-05-20T09:00:00Z"},
        {"tenant_id": "demo-tenant", "organization_id": "demo-org", "team_id": "team-a", "source_channel": "douyin", "stage": "visited", "reported_at": "2026-05-20T10:00:00Z"},
        {"tenant_id": "demo-tenant", "organization_id": "demo-org", "team_id": "team-b", "source_channel": "tencent", "stage": "enrolled", "reported_at": "2026-05-20T11:00:00Z", "conflict_role": "conflict"},
    ]


def number(value: Any) -> float:
    try:
        return float(value)
    except (TypeError, ValueError):
        return 0.0


def integer(value: Any) -> int:
    try:
        return int(float(value))
    except (TypeError, ValueError):
        return 0

