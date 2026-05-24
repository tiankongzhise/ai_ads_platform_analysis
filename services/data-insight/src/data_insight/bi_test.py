from __future__ import annotations

import unittest

from data_insight.bi import BIService
from data_insight.etl import ETLService


class BIServiceTest(unittest.TestCase):
    def test_overview_and_funnel_use_etl_facts(self) -> None:
        etl = ETLService()
        etl.run()
        bi = BIService(etl)
        overview = bi.overview()
        funnel = bi.funnel()
        self.assertEqual(overview["spend"], 128.5)
        self.assertEqual(overview["leads_count"], 3)
        self.assertEqual(funnel["visited"], 2)
        self.assertEqual(funnel["enrolled"], 1)

    def test_team_efficiency_and_channel_roi(self) -> None:
        etl = ETLService()
        etl.run()
        bi = BIService(etl)
        teams = bi.team_efficiency()
        channels = {row["channel"]: row for row in bi.channel_roi()}
        self.assertEqual(teams[0]["team_id"], "team-a")
        self.assertEqual(channels["douyin"]["spend"], 128.5)
        self.assertEqual(channels["douyin"]["leads_count"], 2)


if __name__ == "__main__":
    unittest.main()

