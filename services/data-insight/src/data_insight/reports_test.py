from __future__ import annotations

import unittest
import zipfile
from io import BytesIO

from data_insight.bi import BIService
from data_insight.etl import ETLService
from data_insight.reports import ReportService, build_xlsx


class ReportServiceTest(unittest.TestCase):
    def test_create_report_generates_downloadable_xlsx(self) -> None:
        service = ReportService(BIService(ETLService()))

        task = service.create_task({"tenant_id": "demo-tenant", "report_type": "standard_summary", "created_by": "tester"})
        filename, content_type, body = service.download_file(task["id"], task["download_token"])

        self.assertTrue(filename.endswith(".xlsx"))
        self.assertEqual(content_type, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
        self.assertGreater(len(body), 1000)
        self.assertEqual(task["status"], "success")
        self.assertEqual(task["consistency"]["status"], "passed")
        self.assertGreaterEqual(len(task["advice"]), 1)

    def test_download_rejects_wrong_token(self) -> None:
        service = ReportService(BIService(ETLService()))
        task = service.create_task({"tenant_id": "demo-tenant"})

        with self.assertRaises(PermissionError):
            service.download_file(task["id"], "wrong-token")

    def test_report_task_can_filter_by_tenant(self) -> None:
        service = ReportService(BIService(ETLService()))
        service.create_task({"tenant_id": "tenant-a"})
        service.create_task({"tenant_id": "tenant-b"})

        self.assertEqual(len(service.list_tasks("tenant-a")), 1)
        self.assertEqual(len(service.list_tasks("tenant-b")), 1)
        self.assertEqual(len(service.list_tasks()), 2)

    def test_build_xlsx_contains_workbook_and_sheet(self) -> None:
        body = build_xlsx({"测试": [{"name": "线索", "value": 3}]})

        with zipfile.ZipFile(BytesIO(body), "r") as archive:
            names = set(archive.namelist())
            self.assertIn("xl/workbook.xml", names)
            self.assertIn("xl/worksheets/sheet1.xml", names)
            sheet = archive.read("xl/worksheets/sheet1.xml").decode("utf-8")
            self.assertIn("线索", sheet)


if __name__ == "__main__":
    unittest.main()
