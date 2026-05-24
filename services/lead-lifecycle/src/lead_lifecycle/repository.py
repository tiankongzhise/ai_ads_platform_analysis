from __future__ import annotations

from collections import OrderedDict

from lead_lifecycle.domain import ImportBatch, Lead


class MemoryRepository:
    def __init__(self) -> None:
        self._batches: OrderedDict[str, ImportBatch] = OrderedDict()
        self._leads: OrderedDict[str, Lead] = OrderedDict()
        self._lead_key_index: dict[tuple[str, str, str], str] = {}
        self._staged_rows: dict[str, list[dict[str, str]]] = {}

    def save_batch(self, batch: ImportBatch) -> ImportBatch:
        self._batches[batch.id] = batch
        return batch

    def batch(self, batch_id: str) -> ImportBatch | None:
        return self._batches.get(batch_id)

    def batches(self) -> list[ImportBatch]:
        return list(reversed(self._batches.values()))

    def save_staged_rows(self, batch_id: str, rows: list[dict[str, str]]) -> None:
        self._staged_rows[batch_id] = rows

    def staged_rows(self, batch_id: str) -> list[dict[str, str]]:
        return self._staged_rows.get(batch_id, [])

    def save_lead(self, lead: Lead) -> Lead:
        self._leads[lead.id] = lead
        self._lead_key_index[(lead.tenant_id, lead.team_id, lead.phone_hash)] = lead.id
        return lead

    def lead_by_team_phone(self, tenant_id: str, team_id: str, phone_hash: str) -> Lead | None:
        lead_id = self._lead_key_index.get((tenant_id, team_id, phone_hash))
        if not lead_id:
            return None
        return self._leads.get(lead_id)

    def leads(self, tenant_id: str = "", team_id: str = "", limit: int = 100) -> list[Lead]:
        rows = list(reversed(self._leads.values()))
        if tenant_id:
            rows = [item for item in rows if item.tenant_id == tenant_id]
        if team_id:
            rows = [item for item in rows if item.team_id == team_id]
        return rows[:limit]

