from __future__ import annotations

from dataclasses import asdict, dataclass


@dataclass(frozen=True)
class XiaohongshuReportRequestShape:
    platform: str
    sdk_source: str
    account_model: str
    campaign_model: str
    offline_model: str
    report_model: str
    report_types: list[str]
    dimensions: list[str]
    metrics: list[str]
    raw_policy: str


def request_shape() -> dict:
    shape = XiaohongshuReportRequestShape(
        platform="xiaohongshu",
        sdk_source="community-go-sdk-reference",
        account_model="realtime.AdvertiserRequest",
        campaign_model="realtime.CampaignRequest",
        offline_model="offline.Request",
        report_model="DataReportDTO",
        report_types=["account_daily", "campaign_daily", "unit_daily", "account_realtime"],
        dimensions=["advertiser_id", "campaign_id", "unit_id", "date"],
        metrics=["fee", "impression", "click", "leads"],
        raw_policy="wrap REST response rows without using the community SDK as production dependency",
    )
    return asdict(shape)


def preview_rows(advertiser_id: str = "demo_xiaohongshu_account") -> list[dict]:
    return [
        {
            "advertiser_id": advertiser_id,
            "date": "2026-05-20",
            "report_type": "account_realtime",
            "fee": 96.8,
            "impression": 4100,
            "click": 205,
            "leads": 19,
            "sdk_model": "realtime.AdvertiserRequest",
            "sdk_response_body": "DataReportDTO",
        },
        {
            "advertiser_id": advertiser_id,
            "campaign_id": f"{advertiser_id}_campaign_1",
            "unit_id": f"{advertiser_id}_campaign_1_unit_1",
            "date": "2026-05-20",
            "report_type": "unit_daily",
            "fee": 38.6,
            "impression": 1700,
            "click": 81,
            "leads": 7,
            "sdk_model": "offline.Request",
            "sdk_response_body": "DataReportDTO",
        },
    ]

