from __future__ import annotations

import json
import uuid
import zipfile
from dataclasses import asdict, dataclass, field
from datetime import datetime, timezone
from html import escape
from io import BytesIO
from typing import Any


def utc_now() -> str:
    return datetime.now(timezone.utc).isoformat().replace("+00:00", "Z")


@dataclass
class ReportTask:
    id: str
    tenant_id: str
    report_type: str
    status: str
    format: str
    parameters: dict[str, Any]
    file_name: str
    download_token: str
    download_url: str
    row_count: int
    summary: dict[str, Any]
    consistency: dict[str, Any]
    advice: list[dict[str, Any]]
    created_by: str = "system"
    created_at: str = field(default_factory=utc_now)
    completed_at: str = field(default_factory=utc_now)

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


class ReportService:
    def __init__(self, bi: Any) -> None:
        self.bi = bi
        self.tasks: dict[str, ReportTask] = {}
        self.files: dict[str, bytes] = {}

    def list_tasks(self, tenant_id: str = "") -> list[dict[str, Any]]:
        rows = list(self.tasks.values())
        if tenant_id:
            rows = [row for row in rows if row.tenant_id == tenant_id]
        return [row.to_dict() for row in sorted(rows, key=lambda item: item.created_at, reverse=True)]

    def get_task(self, report_id: str) -> dict[str, Any]:
        task = self.tasks.get(report_id)
        if not task:
            raise KeyError("report not found")
        return task.to_dict()

    def create_task(self, payload: dict[str, Any]) -> dict[str, Any]:
        report_type = str(payload.get("report_type", "standard_summary"))
        tenant_id = str(payload.get("tenant_id", "demo-tenant"))
        report_format = str(payload.get("format", "xlsx")).lower()
        if report_format != "xlsx":
            raise ValueError("only xlsx export is supported")

        sheets = self.build_sheets(report_type)
        summary = self.summary()
        consistency = self.consistency(summary)
        advice = self.action_advice()
        if report_type == "standard_summary":
            sheets["行动建议"] = advice
            sheets["口径校验"] = consistency["checks"]
        else:
            sheets["行动建议"] = advice[:5]
            sheets["口径校验"] = consistency["checks"]

        report_id = str(uuid.uuid4())
        token = uuid.uuid4().hex
        file_name = f"{report_type}-{datetime.now(timezone.utc).strftime('%Y%m%d%H%M%S')}.xlsx"
        content = build_xlsx(sheets)
        task = ReportTask(
            id=report_id,
            tenant_id=tenant_id,
            report_type=report_type,
            status="success",
            format=report_format,
            parameters=payload.get("parameters") or {},
            file_name=file_name,
            download_token=token,
            download_url=f"/api/reports/{report_id}/download?token={token}",
            row_count=sum(len(rows) for rows in sheets.values()),
            summary=summary,
            consistency=consistency,
            advice=advice,
            created_by=str(payload.get("created_by", "system")),
        )
        self.tasks[report_id] = task
        self.files[report_id] = content
        return task.to_dict()

    def download_file(self, report_id: str, token: str) -> tuple[str, str, bytes]:
        task = self.tasks.get(report_id)
        if not task:
            raise KeyError("report not found")
        if not token or token != task.download_token:
            raise PermissionError("invalid download token")
        return task.file_name, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", self.files[report_id]

    def build_sheets(self, report_type: str) -> dict[str, list[dict[str, Any]]]:
        if report_type == "group_overview":
            return {"集团总览": [self.bi.overview()]}
        if report_type == "team_efficiency":
            return {"团队效率": self.bi.team_efficiency()}
        if report_type == "channel_roi":
            return {"渠道 ROI": self.bi.channel_roi()}
        if report_type == "funnel":
            return {"招生漏斗": [self.bi.funnel()]}
        if report_type == "conflicts":
            return {"冲突治理": [self.bi.conflicts()]}
        if report_type == "hourly_trend":
            return {"小时趋势": self.bi.hourly_trend()}
        return {
            "集团总览": [self.bi.overview()],
            "团队效率": self.bi.team_efficiency(),
            "渠道 ROI": self.bi.channel_roi(),
            "招生漏斗": [self.bi.funnel()],
            "冲突治理": [self.bi.conflicts()],
            "小时趋势": self.bi.hourly_trend(),
        }

    def summary(self) -> dict[str, Any]:
        overview = self.bi.overview()
        funnel = self.bi.funnel()
        conflicts = self.bi.conflicts()
        return {
            "spend": overview["spend"],
            "leads_count": overview["leads_count"],
            "enroll_count": overview["enroll_count"],
            "cost_per_lead": overview["cost_per_lead"],
            "conversion_rate": overview["conversion_rate"],
            "conflict_count": conflicts["conflict_count"],
            "conflict_rate": conflicts["conflict_rate"],
            "funnel_valid": funnel["valid"],
            "metric_version": overview["metric_version"],
        }

    def consistency(self, summary: dict[str, Any] | None = None) -> dict[str, Any]:
        summary = summary or self.summary()
        overview = self.bi.overview()
        conflicts = self.bi.conflicts()
        checks = [
            metric_check("spend", summary["spend"], overview["spend"]),
            metric_check("leads_count", summary["leads_count"], overview["leads_count"]),
            metric_check("enroll_count", summary["enroll_count"], overview["enroll_count"]),
            metric_check("cost_per_lead", summary["cost_per_lead"], overview["cost_per_lead"]),
            metric_check("conflict_rate", summary["conflict_rate"], conflicts["conflict_rate"]),
        ]
        return {"status": "passed" if all(item["matched"] for item in checks) else "failed", "checks": checks}

    def action_advice(self) -> list[dict[str, Any]]:
        overview = self.bi.overview()
        channel_rows = self.bi.channel_roi()
        team_rows = self.bi.team_efficiency()
        conflicts = self.bi.conflicts()
        advice: list[dict[str, Any]] = []

        high_cost_channels = [row for row in channel_rows if float(row.get("cost_per_lead", 0)) >= 80]
        if high_cost_channels:
            advice.append(
                {
                    "severity": "high",
                    "metric_code": "cost_per_lead",
                    "title": "线索成本偏高",
                    "detail": f"{len(high_cost_channels)} 个渠道线索成本高于 80。",
                    "action": "下调高成本渠道预算，并复盘计划定向、素材和表单质量。",
                }
            )
        if float(overview.get("conversion_rate", 0)) < 0.25:
            advice.append(
                {
                    "severity": "medium",
                    "metric_code": "conversion_rate",
                    "title": "报名转化率低于阈值",
                    "detail": f"当前报名转化率为 {float(overview.get('conversion_rate', 0)) * 100:.2f}%。",
                    "action": "检查咨询跟进时效和到校邀约策略，优先处理高意向线索。",
                }
            )
        weak_teams = [row for row in team_rows if float(row.get("valid_rate", 0)) < 0.7]
        if weak_teams:
            advice.append(
                {
                    "severity": "medium",
                    "metric_code": "valid_rate",
                    "title": "部分团队有效率不足",
                    "detail": f"{len(weak_teams)} 个团队有效率低于 70%。",
                    "action": "复核无效原因，优化团队分配规则和渠道准入口径。",
                }
            )
        if float(conflicts.get("conflict_rate", 0)) >= 0.1:
            advice.append(
                {
                    "severity": "high",
                    "metric_code": "conflict_rate",
                    "title": "线索冲突率偏高",
                    "detail": f"当前冲突率为 {float(conflicts.get('conflict_rate', 0)) * 100:.2f}%。",
                    "action": "启用跨团队冲突治理看板，按主咨询裁定结果回溯渠道归属。",
                }
            )
        if not advice:
            advice.append(
                {
                    "severity": "low",
                    "metric_code": "overall",
                    "title": "核心指标稳定",
                    "detail": "当前成本、转化和冲突指标未触发风险阈值。",
                    "action": "保持当前投放节奏，并继续观察小时趋势中的异常波动。",
                }
            )
        return advice


def metric_check(metric_code: str, report_value: Any, bi_value: Any) -> dict[str, Any]:
    report_number = float(report_value)
    bi_number = float(bi_value)
    return {
        "metric_code": metric_code,
        "report_value": report_value,
        "bi_value": bi_value,
        "matched": abs(report_number - bi_number) <= 0.01,
    }


def build_xlsx(sheets: dict[str, list[dict[str, Any]]]) -> bytes:
    buffer = BytesIO()
    sheet_items = list(sheets.items())
    with zipfile.ZipFile(buffer, "w", zipfile.ZIP_DEFLATED) as archive:
        archive.writestr("[Content_Types].xml", content_types_xml(len(sheet_items)))
        archive.writestr("_rels/.rels", package_rels_xml())
        archive.writestr("xl/workbook.xml", workbook_xml([name for name, _ in sheet_items]))
        archive.writestr("xl/_rels/workbook.xml.rels", workbook_rels_xml(len(sheet_items)))
        archive.writestr("xl/styles.xml", styles_xml())
        for index, (_, rows) in enumerate(sheet_items, start=1):
            archive.writestr(f"xl/worksheets/sheet{index}.xml", worksheet_xml(rows))
    return buffer.getvalue()


def worksheet_xml(rows: list[dict[str, Any]]) -> str:
    headers = headers_for(rows)
    sheet_rows: list[str] = []
    if headers:
        sheet_rows.append(row_xml(1, headers))
        for index, row in enumerate(rows, start=2):
            sheet_rows.append(row_xml(index, [format_cell(row.get(header, "")) for header in headers]))
    return '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>' + "".join(sheet_rows) + "</sheetData></worksheet>"


def row_xml(row_index: int, values: list[Any]) -> str:
    cells = []
    for column_index, value in enumerate(values, start=1):
        ref = f"{column_name(column_index)}{row_index}"
        cells.append(f'<c r="{ref}" t="inlineStr"><is><t>{escape(str(value))}</t></is></c>')
    return f'<row r="{row_index}">{"".join(cells)}</row>'


def headers_for(rows: list[dict[str, Any]]) -> list[str]:
    headers: list[str] = []
    for row in rows:
        for key in row:
            if key not in headers:
                headers.append(key)
    return headers


def format_cell(value: Any) -> str:
    if isinstance(value, (dict, list)):
        return json.dumps(value, ensure_ascii=False, separators=(",", ":"))
    return str(value)


def column_name(index: int) -> str:
    result = ""
    while index:
        index, remainder = divmod(index - 1, 26)
        result = chr(65 + remainder) + result
    return result


def content_types_xml(sheet_count: int) -> str:
    overrides = "".join(
        f'<Override PartName="/xl/worksheets/sheet{index}.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>'
        for index in range(1, sheet_count + 1)
    )
    return '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>' + overrides + "</Types>"


def package_rels_xml() -> str:
    return '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>'


def workbook_xml(sheet_names: list[str]) -> str:
    sheets = "".join(
        f'<sheet name="{escape(name[:31])}" sheetId="{index}" r:id="rId{index}"/>'
        for index, name in enumerate(sheet_names, start=1)
    )
    return '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets>' + sheets + "</sheets></workbook>"


def workbook_rels_xml(sheet_count: int) -> str:
    sheet_rels = "".join(
        f'<Relationship Id="rId{index}" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet{index}.xml"/>'
        for index in range(1, sheet_count + 1)
    )
    style_rel = f'<Relationship Id="rId{sheet_count + 1}" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>'
    return '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">' + sheet_rels + style_rel + "</Relationships>"


def styles_xml() -> str:
    return '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><fonts count="1"><font><sz val="11"/><name val="Calibri"/></font></fonts><fills count="1"><fill><patternFill patternType="none"/></fill></fills><borders count="1"><border/></borders><cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs><cellXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellXfs></styleSheet>'
