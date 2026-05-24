from __future__ import annotations

import unittest

from data_insight.bi_config import BIConfigService


class BIConfigServiceTest(unittest.TestCase):
    def test_catalog_contains_controlled_metrics_and_datasets(self) -> None:
        service = BIConfigService()
        metric_codes = {item["metric_code"] for item in service.metrics()}
        dataset_codes = {item["dataset_code"] for item in service.datasets()}
        self.assertIn("spend", metric_codes)
        self.assertIn("cost_per_lead", metric_codes)
        self.assertIn("fact_ad_daily", dataset_codes)

    def test_dashboard_create_and_update(self) -> None:
        service = BIConfigService()
        dashboard = service.create_dashboard({"tenant_id": "tenant-1", "dashboard_code": "ops", "name": "运营看板"})
        updated = service.update_dashboard(dashboard["id"], {"name": "运营看板 V2", "filters": {"team_id": "team-a"}})
        self.assertEqual(updated["name"], "运营看板 V2")
        self.assertEqual(updated["filters"]["team_id"], "team-a")

    def test_extension_enable_disable_and_rollback(self) -> None:
        service = BIConfigService()
        enabled = service.enable_extension("tenant-1", "school_custom")
        self.assertEqual(enabled["status"], "enabled")
        disabled = service.disable_extension("tenant-1", "school_custom")
        self.assertEqual(disabled["status"], "disabled")
        rollback = service.rollback_extension("tenant-1", "school_custom")
        self.assertEqual(rollback["status"], "rollback_pending")
        self.assertIn("rollback_requested_at", rollback["manifest"])


if __name__ == "__main__":
    unittest.main()

