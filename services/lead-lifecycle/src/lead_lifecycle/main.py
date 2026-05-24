from __future__ import annotations

import json
from http import HTTPStatus
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import parse_qs, urlparse

from lead_lifecycle.repository import MemoryRepository
from lead_lifecycle.service import CreateBatchRequest, LeadImportService


repository = MemoryRepository()
service = LeadImportService(repository)


class Handler(BaseHTTPRequestHandler):
    def do_GET(self) -> None:
        parsed = urlparse(self.path)
        if parsed.path == "/healthz":
            self._json({"success": True, "data": {"status": "ok", "service": "lead-lifecycle"}})
            return
        if parsed.path == "/api/leads/imports":
            self._json({"success": True, "data": [batch.to_dict() for batch in repository.batches()]})
            return
        if parsed.path.startswith("/api/leads/imports/"):
            batch_id = parsed.path.removeprefix("/api/leads/imports/").strip("/")
            batch = repository.batch(batch_id)
            if not batch:
                self._error(HTTPStatus.NOT_FOUND, "BATCH_NOT_FOUND", "导入批次不存在")
                return
            self._json({"success": True, "data": batch.to_dict()})
            return
        if parsed.path == "/api/leads":
            params = parse_qs(parsed.query)
            leads = repository.leads(
                tenant_id=params.get("tenant_id", [""])[0],
                team_id=params.get("team_id", [""])[0],
                limit=int(params.get("limit", ["100"])[0]),
            )
            self._json({"success": True, "data": [lead.to_dict() for lead in leads]})
            return
        self.send_error(404)

    def do_POST(self) -> None:
        parsed = urlparse(self.path)
        if parsed.path == "/api/leads/imports":
            payload = self._read_json()
            batch = service.create_batch(
                CreateBatchRequest(
                    tenant_id=payload.get("tenant_id", ""),
                    organization_id=payload.get("organization_id", ""),
                    team_id=payload.get("team_id", ""),
                    channel_id=payload.get("channel_id", ""),
                )
            )
            self._json({"success": True, "data": batch.to_dict()})
            return
        if parsed.path.endswith("/upload") and parsed.path.startswith("/api/leads/imports/"):
            batch_id = parsed.path.removeprefix("/api/leads/imports/").removesuffix("/upload").strip("/")
            filename, content = self._read_upload()
            try:
                batch = service.upload(batch_id, filename, content)
            except KeyError:
                self._error(HTTPStatus.NOT_FOUND, "BATCH_NOT_FOUND", "导入批次不存在")
                return
            self._json({"success": True, "data": batch.to_dict()})
            return
        if parsed.path.endswith("/confirm") and parsed.path.startswith("/api/leads/imports/"):
            batch_id = parsed.path.removeprefix("/api/leads/imports/").removesuffix("/confirm").strip("/")
            payload = self._read_json()
            try:
                batch = service.confirm(batch_id, payload.get("mapping", {}))
            except KeyError:
                self._error(HTTPStatus.NOT_FOUND, "BATCH_NOT_FOUND", "导入批次不存在")
                return
            self._json({"success": True, "data": batch.to_dict()})
            return
        if parsed.path == "/api/leads":
            lead, error = service.create_lead(self._read_json())
            if error:
                self._error(HTTPStatus.BAD_REQUEST, "LEAD_INVALID", error.message, {"field": error.field})
                return
            self._json({"success": True, "data": lead.to_dict() if lead else None})
            return
        if parsed.path == "/api/leads/batch":
            payload = self._read_json()
            results = []
            errors = []
            for row in payload.get("items", []):
                lead, error = service.create_lead(row)
                if error:
                    errors.append(error.to_dict())
                elif lead:
                    results.append(lead.to_dict())
            self._json({"success": True, "data": {"created": results, "errors": errors}})
            return
        self.send_error(404)

    def _read_json(self) -> dict:
        length = int(self.headers.get("Content-Length", "0"))
        if length <= 0:
            return {}
        return json.loads(self.rfile.read(length).decode("utf-8"))

    def _read_upload(self) -> tuple[str, bytes]:
        content_type = self.headers.get("Content-Type", "")
        length = int(self.headers.get("Content-Length", "0"))
        body = self.rfile.read(length)
        if content_type.startswith("application/json"):
            payload = json.loads(body.decode("utf-8"))
            return payload.get("filename", "leads.csv"), payload.get("content", "").encode("utf-8")
        if "multipart/form-data" not in content_type:
            return "leads.csv", body
        boundary = content_type.split("boundary=", 1)[-1].encode("utf-8")
        for part in body.split(b"--" + boundary):
            if b"filename=" not in part:
                continue
            header_blob, _, content = part.partition(b"\r\n\r\n")
            disposition = header_blob.decode("utf-8", errors="ignore")
            filename = "leads.csv"
            for item in disposition.split(";"):
                item = item.strip()
                if item.startswith("filename="):
                    filename = item.split("=", 1)[1].strip('"')
            return filename, content.rstrip(b"\r\n-")
        return "leads.csv", body

    def _json(self, payload: dict, status: int = 200) -> None:
        body = json.dumps({**payload, "request_id": "req_local"}, ensure_ascii=False).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def _error(self, status: HTTPStatus, code: str, message: str, details: dict | None = None) -> None:
        self._json(
            {"success": False, "error": {"code": code, "message": message, "details": details or {}}},
            status=status.value,
        )


def main() -> None:
    server = ThreadingHTTPServer(("127.0.0.1", 8090), Handler)
    print("lead-lifecycle listening on http://127.0.0.1:8090", flush=True)
    server.serve_forever()


if __name__ == "__main__":
    main()
