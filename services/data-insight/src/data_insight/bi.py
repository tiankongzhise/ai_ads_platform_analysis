from __future__ import annotations

from typing import Any

from data_insight.etl import ETLService


class BIService:
    def __init__(self, etl: ETLService) -> None:
        self.etl = etl

    def overview(self) -> dict[str, Any]:
        latest = self.ensure_data()
        ad_facts = latest["fact_ad_daily"]
        lead_facts = latest["fact_lead_daily"]
        spend = sum(float(row["spend"]) for row in ad_facts)
        leads = sum(int(row["leads_count"]) for row in lead_facts)
        enrolls = sum(int(row["enroll_count"]) for row in lead_facts)
        return {
            "spend": round(spend, 2),
            "leads_count": leads,
            "enroll_count": enrolls,
            "cost_per_lead": round(spend / leads, 2) if leads else 0,
            "conversion_rate": round(enrolls / leads, 4) if leads else 0,
            "metric_version": "metric-v1",
        }

    def team_efficiency(self) -> list[dict[str, Any]]:
        latest = self.ensure_data()
        grouped: dict[str, dict[str, Any]] = {}
        for row in latest["fact_lead_daily"]:
            team_id = row["team_id"]
            item = grouped.setdefault(
                team_id,
                {"team_id": team_id, "leads_count": 0, "valid_leads_count": 0, "visit_count": 0, "enroll_count": 0, "conflict_count": 0},
            )
            for key in ["leads_count", "valid_leads_count", "visit_count", "enroll_count", "conflict_count"]:
                item[key] += int(row[key])
        for item in grouped.values():
            item["valid_rate"] = round(item["valid_leads_count"] / item["leads_count"], 4) if item["leads_count"] else 0
            item["enroll_rate"] = round(item["enroll_count"] / item["leads_count"], 4) if item["leads_count"] else 0
        return sorted(grouped.values(), key=lambda item: item["leads_count"], reverse=True)

    def channel_roi(self) -> list[dict[str, Any]]:
        latest = self.ensure_data()
        spend_by_platform: dict[str, float] = {}
        for row in latest["fact_ad_daily"]:
            spend_by_platform[row["platform"]] = spend_by_platform.get(row["platform"], 0) + float(row["spend"])
        leads_by_channel: dict[str, int] = {}
        for row in latest["fact_lead_daily"]:
            leads_by_channel[row["channel_text"]] = leads_by_channel.get(row["channel_text"], 0) + int(row["leads_count"])
        channels = sorted(set(spend_by_platform) | set(leads_by_channel))
        return [
            {
                "channel": channel,
                "spend": round(spend_by_platform.get(channel, 0), 2),
                "leads_count": leads_by_channel.get(channel, 0),
                "cost_per_lead": round(spend_by_platform.get(channel, 0) / leads_by_channel[channel], 2) if leads_by_channel.get(channel) else 0,
            }
            for channel in channels
        ]

    def funnel(self) -> dict[str, int]:
        latest = self.ensure_data()
        return {
            "leads": sum(int(row["leads_count"]) for row in latest["fact_lead_daily"]),
            "valid": sum(int(row["valid_leads_count"]) for row in latest["fact_lead_daily"]),
            "visited": sum(int(row["visit_count"]) for row in latest["fact_lead_daily"]),
            "enrolled": sum(int(row["enroll_count"]) for row in latest["fact_lead_daily"]),
        }

    def conflicts(self) -> dict[str, Any]:
        latest = self.ensure_data()
        leads = sum(int(row["leads_count"]) for row in latest["fact_lead_daily"])
        conflicts = sum(int(row["conflict_count"]) for row in latest["fact_lead_daily"])
        return {"conflict_count": conflicts, "conflict_rate": round(conflicts / leads, 4) if leads else 0}

    def hourly_trend(self) -> list[dict[str, Any]]:
        latest = self.ensure_data()
        return latest["fact_hourly_aggregate"]

    def ensure_data(self) -> dict[str, Any]:
        if not self.etl.batches:
            self.etl.run()
        return self.etl.latest()

