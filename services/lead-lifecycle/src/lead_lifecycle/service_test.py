from __future__ import annotations

import unittest

from lead_lifecycle.repository import MemoryRepository
from lead_lifecycle.service import CreateBatchRequest, LeadImportService, hash_phone, normalize_phone, suggest_mapping


class LeadImportServiceTest(unittest.TestCase):
    def test_normalize_phone_and_hash(self) -> None:
        self.assertEqual(normalize_phone("+86 138-0013-8000"), "13800138000")
        self.assertEqual(normalize_phone("12345"), "")
        self.assertEqual(len(hash_phone("13800138000")), 64)

    def test_suggest_mapping_from_chinese_headers(self) -> None:
        mapping = suggest_mapping(["学生姓名", "手机号", "渠道", "线索阶段"])
        self.assertEqual(mapping["student_name"], "学生姓名")
        self.assertEqual(mapping["phone"], "手机号")
        self.assertEqual(mapping["source_channel"], "渠道")
        self.assertEqual(mapping["stage"], "线索阶段")

    def test_confirm_import_deduplicates_within_team(self) -> None:
        repository = MemoryRepository()
        service = LeadImportService(repository)
        batch = service.create_batch(
            CreateBatchRequest(
                tenant_id="tenant-1",
                organization_id="org-1",
                team_id="team-1",
                channel_id="douyin",
            )
        )
        csv_content = "学生姓名,手机号,渠道\n张三,13800138000,抖音\n李四,13800138000,腾讯\n".encode("utf-8")
        uploaded = service.upload(batch.id, "leads.csv", csv_content)
        self.assertEqual(uploaded.total_rows, 2)
        self.assertEqual(uploaded.mapping_suggestion["phone"], "手机号")
        confirmed = service.confirm(batch.id)
        self.assertEqual(confirmed.status, "partial_success")
        self.assertEqual(confirmed.success_rows, 1)
        self.assertEqual(confirmed.failed_rows, 1)
        self.assertEqual(confirmed.errors[0].message, "当前团队已存在相同手机号线索")
        leads = repository.leads(tenant_id="tenant-1", team_id="team-1")
        self.assertEqual(len(leads), 1)
        self.assertEqual(leads[0].phone_masked, "138****8000")

    def test_cross_team_same_phone_creates_conflict_group(self) -> None:
        repository = MemoryRepository()
        service = LeadImportService(repository)
        first, first_error = service.create_lead(
            {
                "tenant_id": "tenant-1",
                "organization_id": "org-a",
                "team_id": "team-a",
                "channel_id": "douyin",
                "student_name": "Alice",
                "phone": "13800138000",
                "reported_at": "2026-05-20T09:00:00Z",
            }
        )
        second, second_error = service.create_lead(
            {
                "tenant_id": "tenant-1",
                "organization_id": "org-b",
                "team_id": "team-b",
                "channel_id": "tencent",
                "student_name": "Bob",
                "phone": "13800138000",
                "reported_at": "2026-05-20T10:00:00Z",
            }
        )
        self.assertIsNone(first_error)
        self.assertIsNone(second_error)
        self.assertIsNotNone(first)
        self.assertIsNotNone(second)
        groups = repository.conflict_groups(tenant_id="tenant-1", status="open")
        self.assertEqual(len(groups), 1)
        self.assertEqual(groups[0].primary_lead_id, first.id)
        self.assertEqual(groups[0].conflict_count, 2)
        self.assertEqual(repository.attributions(tenant_id="tenant-1", rule="first_report")[0].owner_team_id, "team-a")

    def test_manual_resolve_marks_primary_and_conflict_roles(self) -> None:
        repository = MemoryRepository()
        service = LeadImportService(repository)
        first, _ = service.create_lead({"tenant_id": "tenant-1", "team_id": "team-a", "organization_id": "org-a", "phone": "13800138000"})
        second, _ = service.create_lead({"tenant_id": "tenant-1", "team_id": "team-b", "organization_id": "org-b", "phone": "13800138000"})
        assert first is not None
        assert second is not None
        group = repository.conflict_groups()[0]
        resolved = service.conflicts.resolve(group.id, second.id, rule="manual", resolved_by="manager-1")
        self.assertEqual(resolved.status, "resolved")
        self.assertEqual(resolved.primary_lead_id, second.id)
        roles = {item.lead_id: item.conflict_role for item in resolved.items}
        self.assertEqual(roles[second.id], "primary")
        self.assertEqual(roles[first.id], "conflict")


if __name__ == "__main__":
    unittest.main()
