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


if __name__ == "__main__":
    unittest.main()

