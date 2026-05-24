from __future__ import annotations

import uuid
from dataclasses import asdict, dataclass, field
from datetime import datetime, timezone
from typing import Any


def utc_now() -> str:
    return datetime.now(timezone.utc).isoformat().replace("+00:00", "Z")


@dataclass
class Dashboard:
    id: str
    tenant_id: str
    dashboard_code: str
    name: str
    layout: dict[str, Any]
    filters: dict[str, Any] = field(default_factory=dict)
    permission_rules: dict[str, Any] = field(default_factory=dict)
    status: str = "active"
    created_at: str = field(default_factory=utc_now)
    updated_at: str = field(default_factory=utc_now)

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


@dataclass
class ExtensionState:
    id: str
    tenant_id: str
    extension_code: str
    version: str
    status: str
    manifest: dict[str, Any]
    enabled_at: str = ""
    disabled_at: str = ""
    created_at: str = field(default_factory=utc_now)

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


class BIConfigService:
    def __init__(self) -> None:
        self.dashboards: dict[str, Dashboard] = {}
        self.extensions: dict[tuple[str, str], ExtensionState] = {}

    def metrics(self) -> list[dict[str, Any]]:
        return [
            {"metric_code": "spend", "name": "广告消耗", "dataset_code": "fact_ad_daily", "unit": "currency", "formula": {"field": "spend", "agg": "sum"}, "version": "metric-v1"},
            {"metric_code": "leads_count", "name": "线索数", "dataset_code": "fact_lead_daily", "unit": "count", "formula": {"field": "leads_count", "agg": "sum"}, "version": "metric-v1"},
            {"metric_code": "cost_per_lead", "name": "线索成本", "dataset_code": "analytics_channel_roi", "unit": "currency", "formula": {"divide": ["spend", "leads_count"]}, "version": "metric-v1"},
            {"metric_code": "conflict_rate", "name": "冲突率", "dataset_code": "analytics_conflicts", "unit": "percent", "formula": {"divide": ["conflict_count", "leads_count"]}, "version": "metric-v1"},
        ]

    def dimensions(self) -> list[dict[str, Any]]:
        return [
            {"dimension_code": "report_date", "name": "日期", "dataset_code": "fact_ad_daily", "type": "date"},
            {"dimension_code": "platform", "name": "广告平台", "dataset_code": "fact_ad_daily", "type": "string"},
            {"dimension_code": "team_id", "name": "招生团队", "dataset_code": "fact_lead_daily", "type": "string"},
            {"dimension_code": "channel_text", "name": "线索渠道", "dataset_code": "fact_lead_daily", "type": "string"},
        ]

    def datasets(self) -> list[dict[str, Any]]:
        return [
            {"dataset_code": "fact_ad_daily", "name": "广告日事实", "source_type": "table", "source_ref": "data_insight.fact_ad_daily", "version": "v1"},
            {"dataset_code": "fact_lead_daily", "name": "线索日事实", "source_type": "table", "source_ref": "data_insight.fact_lead_daily", "version": "v1"},
            {"dataset_code": "fact_hourly_aggregate", "name": "小时聚合事实", "source_type": "table", "source_ref": "data_insight.fact_hourly_aggregate", "version": "v1"},
            {"dataset_code": "analytics_channel_roi", "name": "渠道 ROI 服务数据集", "source_type": "service", "source_ref": "/api/analytics/channel-roi", "version": "v1"},
        ]

    def charts(self) -> list[dict[str, Any]]:
        return [
            {"chart_code": "kpi_card", "name": "KPI 卡片", "supported_metrics": ["spend", "leads_count", "cost_per_lead"]},
            {"chart_code": "table", "name": "明细表", "supported_metrics": ["spend", "leads_count", "conflict_rate"]},
            {"chart_code": "line", "name": "趋势折线", "supported_metrics": ["spend", "leads_count"]},
            {"chart_code": "funnel", "name": "招生漏斗", "supported_metrics": ["leads_count"]},
        ]

    def list_dashboards(self, tenant_id: str = "") -> list[dict[str, Any]]:
        rows = list(self.dashboards.values())
        if tenant_id:
            rows = [row for row in rows if row.tenant_id == tenant_id]
        return [row.to_dict() for row in rows]

    def create_dashboard(self, payload: dict[str, Any]) -> dict[str, Any]:
        dashboard = Dashboard(
            id=str(uuid.uuid4()),
            tenant_id=str(payload.get("tenant_id", "demo-tenant")),
            dashboard_code=str(payload.get("dashboard_code", "custom_dashboard")),
            name=str(payload.get("name", "自定义看板")),
            layout=payload.get("layout") or default_layout(),
            filters=payload.get("filters") or {},
            permission_rules=payload.get("permission_rules") or {},
        )
        self.dashboards[dashboard.id] = dashboard
        return dashboard.to_dict()

    def update_dashboard(self, dashboard_id: str, payload: dict[str, Any]) -> dict[str, Any]:
        dashboard = self.dashboards.get(dashboard_id)
        if not dashboard:
            raise KeyError("dashboard not found")
        dashboard.name = str(payload.get("name", dashboard.name))
        dashboard.layout = payload.get("layout", dashboard.layout)
        dashboard.filters = payload.get("filters", dashboard.filters)
        dashboard.permission_rules = payload.get("permission_rules", dashboard.permission_rules)
        dashboard.status = str(payload.get("status", dashboard.status))
        dashboard.updated_at = utc_now()
        return dashboard.to_dict()

    def enable_extension(self, tenant_id: str, extension_code: str, manifest: dict[str, Any] | None = None) -> dict[str, Any]:
        manifest = manifest or default_extension_manifest(extension_code)
        state = ExtensionState(
            id=str(uuid.uuid4()),
            tenant_id=tenant_id or "demo-tenant",
            extension_code=extension_code,
            version=str(manifest.get("version", "v1")),
            status="enabled",
            manifest=manifest,
            enabled_at=utc_now(),
        )
        self.extensions[(state.tenant_id, extension_code)] = state
        return state.to_dict()

    def disable_extension(self, tenant_id: str, extension_code: str) -> dict[str, Any]:
        state = self.must_extension(tenant_id, extension_code)
        state.status = "disabled"
        state.disabled_at = utc_now()
        return state.to_dict()

    def rollback_extension(self, tenant_id: str, extension_code: str) -> dict[str, Any]:
        state = self.must_extension(tenant_id, extension_code)
        state.status = "rollback_pending"
        state.manifest = {**state.manifest, "rollback_requested_at": utc_now()}
        return state.to_dict()

    def list_extensions(self, tenant_id: str = "") -> list[dict[str, Any]]:
        rows = list(self.extensions.values())
        if tenant_id:
            rows = [row for row in rows if row.tenant_id == tenant_id]
        return [row.to_dict() for row in rows]

    def must_extension(self, tenant_id: str, extension_code: str) -> ExtensionState:
        state = self.extensions.get((tenant_id or "demo-tenant", extension_code))
        if not state:
            raise KeyError("extension not found")
        return state


def default_layout() -> dict[str, Any]:
    return {"widgets": [{"chart_code": "kpi_card", "metric_code": "spend"}, {"chart_code": "table", "metric_code": "leads_count"}]}


def default_extension_manifest(extension_code: str) -> dict[str, Any]:
    return {
        "extension_code": extension_code,
        "version": "v1",
        "metrics": [{"metric_code": f"{extension_code}_custom_metric", "name": "定制指标"}],
        "datasets": [{"dataset_code": f"{extension_code}_dataset", "source_type": "service"}],
        "charts": [{"chart_code": f"{extension_code}_chart", "name": "定制图表"}],
        "rollback": {"strategy": "disable_extension_and_restore_layout"},
    }

