from __future__ import annotations

import json
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import parse_qs, urlparse

try:
    from data_insight.adapters import baidu, xiaohongshu
except ModuleNotFoundError:
    from adapters import baidu, xiaohongshu


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
        self.send_error(404)

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
