from __future__ import annotations

from collections import OrderedDict

from lead_lifecycle.domain import Attribution, ConflictGroup, ImportBatch, Lead


class MemoryRepository:
    def __init__(self) -> None:
        self._batches: OrderedDict[str, ImportBatch] = OrderedDict()
        self._leads: OrderedDict[str, Lead] = OrderedDict()
        self._lead_key_index: dict[tuple[str, str, str], str] = {}
        self._phone_index: dict[tuple[str, str], list[str]] = {}
        self._staged_rows: dict[str, list[dict[str, str]]] = {}
        self._conflict_groups: OrderedDict[str, ConflictGroup] = OrderedDict()
        self._conflict_key_index: dict[tuple[str, str], str] = {}
        self._attributions: OrderedDict[str, Attribution] = OrderedDict()

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
        phone_key = (lead.tenant_id, lead.phone_hash)
        lead_ids = self._phone_index.setdefault(phone_key, [])
        if lead.id not in lead_ids:
            lead_ids.append(lead.id)
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

    def leads_by_phone_hash(self, tenant_id: str, phone_hash: str) -> list[Lead]:
        lead_ids = self._phone_index.get((tenant_id, phone_hash), [])
        return [self._leads[lead_id] for lead_id in lead_ids if lead_id in self._leads]

    def save_conflict_group(self, group: ConflictGroup) -> ConflictGroup:
        self._conflict_groups[group.id] = group
        self._conflict_key_index[(group.tenant_id, group.phone_hash)] = group.id
        return group

    def conflict_group_by_phone(self, tenant_id: str, phone_hash: str) -> ConflictGroup | None:
        group_id = self._conflict_key_index.get((tenant_id, phone_hash))
        if not group_id:
            return None
        return self._conflict_groups.get(group_id)

    def conflict_group(self, group_id: str) -> ConflictGroup | None:
        return self._conflict_groups.get(group_id)

    def conflict_groups(self, tenant_id: str = "", status: str = "") -> list[ConflictGroup]:
        rows = list(reversed(self._conflict_groups.values()))
        if tenant_id:
            rows = [item for item in rows if item.tenant_id == tenant_id]
        if status:
            rows = [item for item in rows if item.status == status]
        return rows

    def save_attribution(self, attribution: Attribution) -> Attribution:
        self._attributions[attribution.id] = attribution
        return attribution

    def attributions(self, tenant_id: str = "", rule: str = "") -> list[Attribution]:
        rows = list(reversed(self._attributions.values()))
        if tenant_id:
            rows = [item for item in rows if item.tenant_id == tenant_id]
        if rule:
            rows = [item for item in rows if item.attribution_rule == rule]
        return rows
