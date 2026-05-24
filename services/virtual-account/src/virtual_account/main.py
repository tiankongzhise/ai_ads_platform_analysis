from __future__ import annotations

import json
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import parse_qs, urlparse

try:
    from virtual_account.service import VirtualAccountService
except ModuleNotFoundError:
    from service import VirtualAccountService


service = VirtualAccountService()


class Handler(BaseHTTPRequestHandler):
    def do_GET(self) -> None:
        if self.path == "/healthz":
            self._json({"success": True, "data": {"status": "ok", "service": "virtual-account"}})
            return
        parsed = urlparse(self.path)
        if parsed.path == "/api/virtual-accounts":
            params = parse_qs(parsed.query)
            self._json({"success": True, "data": service.list_accounts(params.get("tenant_id", [""])[0])})
            return
        if parsed.path == "/api/virtual-accounts/cleanup-logs":
            params = parse_qs(parsed.query)
            self._json({"success": True, "data": service.list_cleanup_logs(params.get("tenant_id", [""])[0])})
            return
        if parsed.path.startswith("/api/virtual-accounts/"):
            account_id = parsed.path.removeprefix("/api/virtual-accounts/").strip("/")
            try:
                self._json({"success": True, "data": service.get_account(account_id)})
            except KeyError:
                self._json({"success": False, "error": {"message": "virtual account not found"}})
            return
        self.send_error(404)

    def do_POST(self) -> None:
        parsed = urlparse(self.path)
        if parsed.path == "/api/virtual-accounts":
            account = service.create_account(self._read_json())
            self._json({"success": True, "data": account})
            return
        if parsed.path == "/api/virtual-accounts/cleanup-expired":
            result = service.cleanup_expired(self._read_json())
            self._json({"success": True, "data": result})
            return
        if parsed.path.startswith("/api/virtual-accounts/"):
            parts = parsed.path.strip("/").split("/")
            if len(parts) == 4 and parts[0:2] == ["api", "virtual-accounts"]:
                account_id = parts[2]
                action = parts[3]
                try:
                    if action == "authorize-migration":
                        result = service.authorize_migration(account_id, self._read_json())
                    elif action == "destroy":
                        result = service.destroy_account(account_id, self._read_json())
                    else:
                        self.send_error(404)
                        return
                except KeyError:
                    self._json({"success": False, "error": {"message": "virtual account not found"}})
                    return
                except ValueError as exc:
                    self._json({"success": False, "error": {"message": str(exc)}})
                    return
                self._json({"success": True, "data": result})
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
    server = ThreadingHTTPServer(("127.0.0.1", 8092), Handler)
    print("virtual-account listening on http://127.0.0.1:8092", flush=True)
    server.serve_forever()


if __name__ == "__main__":
    main()
