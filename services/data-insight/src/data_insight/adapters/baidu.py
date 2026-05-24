from __future__ import annotations

from dataclasses import asdict, dataclass
from importlib.util import find_spec


@dataclass(frozen=True)
class BaiduReportRequestShape:
    platform: str
    sdk_available: bool
    header_model: str
    auth_service: str
    report_service: str
    report_types: list[str]
    dimensions: list[str]
    metrics: list[str]
    raw_policy: str


def sdk_available() -> bool:
    return find_spec("baiduads") is not None


def request_shape() -> dict:
    shape = BaiduReportRequestShape(
        platform="baidu",
        sdk_available=sdk_available(),
        header_model="ApiRequestHeader",
        auth_service="OAuthAuthorizedToolAPI",
        report_service="ReportService",
        report_types=["account_daily", "campaign_daily", "adgroup_daily"],
        dimensions=["userName", "campaignId", "adgroupId", "date"],
        metrics=["cost", "impression", "click", "conversion"],
        raw_policy="preserve original body.data rows before ETL",
    )
    return asdict(shape)


def preview_rows(account_id: str = "demo_baidu_account") -> list[dict]:
    return [
        {
            "userName": account_id,
            "date": "2026-05-20",
            "reportType": "account_daily",
            "cost": 128.5,
            "impression": 6400,
            "click": 320,
            "conversion": 24,
            "header_model": "ApiRequestHeader",
            "sdk_service": "ReportService",
        },
        {
            "userName": account_id,
            "campaignId": f"{account_id}_campaign_1",
            "date": "2026-05-20",
            "reportType": "campaign_daily",
            "cost": 72.3,
            "impression": 3100,
            "click": 151,
            "conversion": 11,
            "header_model": "ApiRequestHeader",
            "sdk_service": "ReportService",
        },
    ]

