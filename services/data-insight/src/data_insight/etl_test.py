from __future__ import annotations

import unittest

from data_insight.etl import ETLService, aggregate_leads, map_ad_row


class ETLServiceTest(unittest.TestCase):
    def test_map_ad_row_normalizes_metrics(self) -> None:
        fact = map_ad_row(
            {
                "tenant_id": "tenant-1",
                "platform": "xiaohongshu",
                "account_id": "account-1",
                "report_type": "unit_daily",
                "stat_date": "2026-05-20",
                "entity_type": "unit",
                "external_entity_id": "unit-1",
                "raw": {"fee": 12.3, "impression": 1000, "click": 80, "leads": 6},
            },
            "batch-1",
        )
        self.assertEqual(fact.spend, 12.3)
        self.assertEqual(fact.impressions, 1000)
        self.assertEqual(fact.clicks, 80)
        self.assertEqual(fact.conversions, 6)

    def test_aggregate_leads_by_day_team_and_channel(self) -> None:
        facts = aggregate_leads(
            [
                {"tenant_id": "tenant-1", "organization_id": "org-1", "team_id": "team-1", "source_channel": "douyin", "stage": "new", "reported_at": "2026-05-20T09:00:00Z"},
                {"tenant_id": "tenant-1", "organization_id": "org-1", "team_id": "team-1", "source_channel": "douyin", "stage": "visited", "reported_at": "2026-05-20T10:00:00Z"},
                {"tenant_id": "tenant-1", "organization_id": "org-1", "team_id": "team-1", "source_channel": "douyin", "stage": "invalid", "reported_at": "2026-05-20T11:00:00Z"},
            ],
            "batch-1",
        )
        self.assertEqual(len(facts), 1)
        self.assertEqual(facts[0].leads_count, 3)
        self.assertEqual(facts[0].valid_leads_count, 2)
        self.assertEqual(facts[0].visit_count, 1)

    def test_run_generates_daily_and_hourly_facts(self) -> None:
        service = ETLService()
        batch = service.run()
        latest = service.latest()
        self.assertEqual(batch.status, "success")
        self.assertEqual(len(latest["fact_ad_daily"]), 1)
        self.assertGreaterEqual(len(latest["fact_lead_daily"]), 2)
        self.assertGreaterEqual(len(latest["fact_hourly_aggregate"]), 3)


if __name__ == "__main__":
    unittest.main()

