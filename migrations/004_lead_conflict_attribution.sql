CREATE TABLE IF NOT EXISTS lead_lifecycle.lead_conflict_groups (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL,
  phone_hash text NOT NULL,
  phone_masked text,
  status text NOT NULL CHECK (status IN ('open', 'resolved', 'ignored')),
  primary_lead_id uuid REFERENCES lead_lifecycle.leads(id),
  conflict_count int NOT NULL DEFAULT 0,
  resolution_meta jsonb NOT NULL DEFAULT '{}'::jsonb,
  resolved_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, phone_hash)
);

CREATE TABLE IF NOT EXISTS lead_lifecycle.lead_conflict_items (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL,
  conflict_group_id uuid NOT NULL REFERENCES lead_lifecycle.lead_conflict_groups(id),
  lead_id uuid NOT NULL REFERENCES lead_lifecycle.leads(id),
  team_id text NOT NULL,
  is_primary boolean NOT NULL DEFAULT false,
  conflict_role text NOT NULL DEFAULT 'candidate',
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, conflict_group_id, lead_id)
);

CREATE TABLE IF NOT EXISTS lead_lifecycle.lead_attributions (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL,
  lead_id uuid NOT NULL REFERENCES lead_lifecycle.leads(id),
  attribution_rule text NOT NULL CHECK (attribution_rule IN ('first_report', 'first_visit', 'first_enroll', 'manual')),
  owner_team_id text,
  owner_organization_id text,
  evidence jsonb NOT NULL DEFAULT '{}'::jsonb,
  decided_by text,
  decided_at timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_lead_conflicts_status ON lead_lifecycle.lead_conflict_groups(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_lead_attributions_rule ON lead_lifecycle.lead_attributions(tenant_id, attribution_rule);

