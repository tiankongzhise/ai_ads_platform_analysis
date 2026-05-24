from __future__ import annotations

from lead_lifecycle.domain import Attribution, ConflictGroup, ConflictItem, Lead, utc_now
from lead_lifecycle.repository import MemoryRepository
from lead_lifecycle.service import new_id


class ConflictService:
    def __init__(self, repository: MemoryRepository) -> None:
        self.repository = repository

    def inspect_lead(self, lead: Lead) -> ConflictGroup | None:
        related = self.repository.leads_by_phone_hash(lead.tenant_id, lead.phone_hash)
        teams = {item.team_id for item in related}
        if len(teams) < 2:
            self.ensure_attribution(lead, "first_report", lead)
            return None
        primary = first_report_lead(related)
        group = self.repository.conflict_group_by_phone(lead.tenant_id, lead.phone_hash)
        if not group:
            group = ConflictGroup(
                id=new_id(),
                tenant_id=lead.tenant_id,
                phone_hash=lead.phone_hash,
                phone_masked=lead.phone_masked,
                primary_lead_id=primary.id,
                resolution_meta={
                    "suggested_rule": "first_report",
                    "suggested_primary_lead_id": primary.id,
                    "suggested_team_id": primary.team_id,
                },
            )
        group.items = [
            ConflictItem(
                lead_id=item.id,
                team_id=item.team_id,
                is_primary=item.id == primary.id,
                conflict_role="suggested_primary" if item.id == primary.id else "candidate",
                created_at=item.created_at,
            )
            for item in related
        ]
        group.primary_lead_id = group.primary_lead_id or primary.id
        group.conflict_count = len(group.items)
        group.updated_at = utc_now()
        self.repository.save_conflict_group(group)
        for item in related:
            self.ensure_attribution(item, "first_report", primary)
        return group

    def resolve(self, group_id: str, primary_lead_id: str, rule: str = "manual", resolved_by: str = "system") -> ConflictGroup:
        group = self.repository.conflict_group(group_id)
        if not group:
            raise KeyError("conflict group not found")
        lead_by_id = {lead.id: lead for lead in self.repository.leads_by_phone_hash(group.tenant_id, group.phone_hash)}
        primary = lead_by_id.get(primary_lead_id)
        if not primary:
            raise ValueError("primary lead is not in conflict group")
        group.status = "resolved"
        group.primary_lead_id = primary_lead_id
        group.resolved_at = utc_now()
        group.updated_at = group.resolved_at
        group.resolution_meta = {
            "rule": rule,
            "resolved_by": resolved_by,
            "primary_team_id": primary.team_id,
        }
        group.items = [
            ConflictItem(
                lead_id=item.lead_id,
                team_id=item.team_id,
                is_primary=item.lead_id == primary_lead_id,
                conflict_role="primary" if item.lead_id == primary_lead_id else "conflict",
                created_at=item.created_at,
            )
            for item in group.items
        ]
        self.repository.save_conflict_group(group)
        for lead in lead_by_id.values():
            self.ensure_attribution(lead, rule, primary)
        return group

    def ensure_attribution(self, lead: Lead, rule: str, owner: Lead) -> Attribution:
        attribution = Attribution(
            id=new_id(),
            tenant_id=lead.tenant_id,
            lead_id=lead.id,
            attribution_rule=rule,
            owner_team_id=owner.team_id,
            owner_organization_id=owner.organization_id,
            evidence={
                "owner_lead_id": owner.id,
                "owner_reported_at": owner.reported_at,
                "phone_masked": owner.phone_masked,
            },
        )
        return self.repository.save_attribution(attribution)


def first_report_lead(leads: list[Lead]) -> Lead:
    return sorted(leads, key=lambda item: (item.reported_at, item.created_at, item.id))[0]

