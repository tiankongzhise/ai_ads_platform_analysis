from __future__ import annotations

import json
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import parse_qs, urlparse

try:
    from data_insight.adapters import baidu, xiaohongshu
    from data_insight.bi_config import BIConfigService
    from data_insight.bi import BIService
    from data_insight.etl import ETLService
except ModuleNotFoundError:
    from adapters import baidu, xiaohongshu
    from bi_config import BIConfigService
    from bi import BIService
    from etl import ETLService


etl_service = ETLService()
bi_service = BIService(etl_service)
bi_config_service = BIConfigService()


class Handler(BaseHTTPRequestHandler):
    def do_GET(self) -> None:
        if self.path == "/healthz":
            self._json({"success": True, "data": {"status": "ok", "service": "data-insight"}})
            return
        parsed = urlparse(self.path)
        if parsed.path == "/api/platform-adapters":
            self._json(
                {
                    "success": True,
                    "data": {
                        "baidu": baidu.request_shape(),
                        "xiaohongshu": xiaohongshu.request_shape(),
                    },
                }
            )
            return
        if parsed.path == "/api/platform-adapters/preview":
            params = parse_qs(parsed.query)
            platform = params.get("platform", ["baidu"])[0]
            account_id = params.get("account_id", [""])[0]
            if platform == "baidu":
                self._json({"success": True, "data": baidu.preview_rows(account_id or "demo_baidu_account")})
                return
            if platform == "xiaohongshu":
                self._json({"success": True, "data": xiaohongshu.preview_rows(account_id or "demo_xiaohongshu_account")})
                return
            self._json({"success": False, "error": {"message": "unsupported platform"}})
            return
        if parsed.path == "/api/etl/batches":
            self._json({"success": True, "data": etl_service.latest()["batches"]})
            return
        if parsed.path == "/api/etl/facts/ad-daily":
            self._json({"success": True, "data": etl_service.latest()["fact_ad_daily"]})
            return
        if parsed.path == "/api/etl/facts/lead-daily":
            self._json({"success": True, "data": etl_service.latest()["fact_lead_daily"]})
            return
        if parsed.path == "/api/etl/facts/hourly":
            self._json({"success": True, "data": etl_service.latest()["fact_hourly_aggregate"]})
            return
        if parsed.path == "/api/analytics/group-overview":
            self._json({"success": True, "data": bi_service.overview()})
            return
        if parsed.path == "/api/analytics/team-efficiency":
            self._json({"success": True, "data": bi_service.team_efficiency()})
            return
        if parsed.path == "/api/analytics/channel-roi":
            self._json({"success": True, "data": bi_service.channel_roi()})
            return
        if parsed.path == "/api/analytics/funnel":
            self._json({"success": True, "data": bi_service.funnel()})
            return
        if parsed.path == "/api/analytics/conflicts":
            self._json({"success": True, "data": bi_service.conflicts()})
            return
        if parsed.path == "/api/analytics/hourly-trend":
            self._json({"success": True, "data": bi_service.hourly_trend()})
            return
        if parsed.path == "/api/bi/catalog/metrics":
            self._json({"success": True, "data": bi_config_service.metrics()})
            return
        if parsed.path == "/api/bi/catalog/dimensions":
            self._json({"success": True, "data": bi_config_service.dimensions()})
            return
        if parsed.path == "/api/bi/catalog/datasets":
            self._json({"success": True, "data": bi_config_service.datasets()})
            return
        if parsed.path == "/api/bi/catalog/charts":
            self._json({"success": True, "data": bi_config_service.charts()})
            return
        if parsed.path == "/api/bi/dashboards":
            params = parse_qs(parsed.query)
            self._json({"success": True, "data": bi_config_service.list_dashboards(params.get("tenant_id", [""])[0])})
            return
        if parsed.path == "/api/bi/extensions":
            params = parse_qs(parsed.query)
            self._json({"success": True, "data": bi_config_service.list_extensions(params.get("tenant_id", [""])[0])})
            return
        self.send_error(404)

    def do_POST(self) -> None:
        parsed = urlparse(self.path)
        if parsed.path == "/api/etl/run":
            payload = self._read_json()
            batch = etl_service.run(payload.get("ad_rows"), payload.get("lead_rows"))
            self._json({"success": True, "data": {"batch": batch.to_dict(), "facts": etl_service.latest()}})
            return
        if parsed.path == "/api/bi/dashboards":
            dashboard = bi_config_service.create_dashboard(self._read_json())
            self._json({"success": True, "data": dashboard})
            return
        if parsed.path.startswith("/api/bi/dashboards/"):
            dashboard_id = parsed.path.removeprefix("/api/bi/dashboards/").strip("/")
            try:
                dashboard = bi_config_service.update_dashboard(dashboard_id, self._read_json())
            except KeyError:
                self._json({"success": False, "error": {"message": "dashboard not found"}})
                return
            self._json({"success": True, "data": dashboard})
            return
        if parsed.path.startswith("/api/bi/extensions/"):
            parts = parsed.path.strip("/").split("/")
            if len(parts) == 5 and parts[0:2] == ["api", "bi"] and parts[2] == "extensions":
                extension_code = parts[3]
                action = parts[4]
                payload = self._read_json()
                tenant_id = payload.get("tenant_id", "demo-tenant")
                try:
                    if action == "enable":
                        state = bi_config_service.enable_extension(tenant_id, extension_code, payload.get("manifest"))
                    elif action == "disable":
                        state = bi_config_service.disable_extension(tenant_id, extension_code)
                    elif action == "rollback":
                        state = bi_config_service.rollback_extension(tenant_id, extension_code)
                    else:
                        self.send_error(404)
                        return
                except KeyError:
                    self._json({"success": False, "error": {"message": "extension not found"}})
                    return
                self._json({"success": True, "data": state})
                return
        self.send_error(404)

    def _read_json(self) -> dict:
        length = int(self.headers.get("Content-Length", "0"))
        if length <= 0:
            return {}
        return json.loads(self.rfile.read(length).decode("utf-8"))

    def _json(self, payload: dict) -> None:
        body = json.dumps(payload, ensure_ascii=False).encode("utf-8")
        self.send_response(200)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)


def main() -> None:
    server = ThreadingHTTPServer(("127.0.0.1", 8091), Handler)
    print("data-insight listening on http://127.0.0.1:8091", flush=True)
    server.serve_forever()


if __name__ == "__main__":
    main()
