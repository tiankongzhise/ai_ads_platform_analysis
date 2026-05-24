from __future__ import annotations

import unittest
from datetime import datetime, timedelta, timezone

from virtual_account.service import VirtualAccountService


class VirtualAccountServiceTest(unittest.TestCase):
    def test_create_account_defaults_to_seven_days(self) -> None:
        now = datetime(2026, 5, 24, tzinfo=timezone.utc)
        service = VirtualAccountService()

        account = service.create_account({"tenant_id": "demo-tenant", "valid_days": 30}, now)

        self.assertEqual(account["valid_days"], 7)
        self.assertEqual(account["expires_at"], "2026-05-31T00:00:00Z")
        self.assertEqual(account["status"], "active")
        self.assertEqual(len(account["resources"]["token_refs"]), 2)

    def test_authorize_migration_marks_account_without_destroying(self) -> None:
        service = VirtualAccountService()
        account = service.create_account({"tenant_id": "demo-tenant"})

        result = service.authorize_migration(account["id"], {"target_user_id": "real-user", "authorized_by": "manager"})

        self.assertEqual(result["account"]["status"], "migration_authorized")
        self.assertEqual(result["authorization"]["target_user_id"], "real-user")
        self.assertFalse(result["account"]["cleanup_verified"])

    def test_manual_destroy_deletes_ephemeral_resources_and_audits(self) -> None:
        service = VirtualAccountService()
        account = service.create_account({"tenant_id": "demo-tenant"})

        result = service.destroy_account(account["id"], {"actor": "manager"})

        self.assertEqual(result["account"]["status"], "destroyed")
        self.assertTrue(result["account"]["cleanup_verified"])
        self.assertEqual(result["cleanup_log"]["token_refs_deleted"], 2)
        self.assertEqual(result["cleanup_log"]["files_deleted"], 1)
        self.assertEqual(result["cleanup_log"]["cache_keys_deleted"], 1)
        self.assertEqual(result["cleanup_log"]["queue_messages_deleted"], 1)
        self.assertTrue(result["cleanup_log"]["no_business_data"])

    def test_cleanup_blocks_when_business_data_exists(self) -> None:
        service = VirtualAccountService()
        account = service.create_account({"tenant_id": "demo-tenant", "business_records": ["lead-1"]})

        result = service.destroy_account(account["id"], {"actor": "manager"})

        self.assertEqual(result["account"]["status"], "cleanup_blocked")
        self.assertFalse(result["cleanup_log"]["no_business_data"])
        self.assertEqual(result["cleanup_log"]["business_records_found"], 1)
        self.assertEqual(result["cleanup_log"]["token_refs_deleted"], 0)

    def test_cleanup_expired_only_cleans_due_accounts(self) -> None:
        now = datetime(2026, 5, 24, tzinfo=timezone.utc)
        service = VirtualAccountService()
        expired = service.create_account({"tenant_id": "demo-tenant"}, now - timedelta(days=8))
        service.create_account({"tenant_id": "demo-tenant"}, now)

        result = service.cleanup_expired({"actor": "scheduler"}, now)

        self.assertEqual(result["cleaned"], 1)
        self.assertEqual(service.get_account(expired["id"])["status"], "destroyed")


if __name__ == "__main__":
    unittest.main()
