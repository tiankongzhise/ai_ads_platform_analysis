CREATE SCHEMA IF NOT EXISTS lead_lifecycle;

CREATE TABLE IF NOT EXISTS lead_lifecycle.lead_import_batches (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL,
  organization_id uuid,
  team_id uuid NOT NULL,
  channel_id uuid,
  status text NOT NULL CHECK (status IN ('created', 'uploaded', 'processing', 'success', 'partial_success', 'failed', 'api_submitted')),
  file_name text,
  total_rows int NOT NULL DEFAULT 0,
  success_rows int NOT NULL DEFAULT 0,
  failed_rows int NOT NULL DEFAULT 0,
  headers jsonb NOT NULL DEFAULT '[]'::jsonb,
  mapping_suggestion jsonb NOT NULL DEFAULT '{}'::jsonb,
  error_report jsonb NOT NULL DEFAULT '[]'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS lead_lifecycle.leads (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL,
  organization_id uuid,
  team_id uuid NOT NULL,
  channel_id uuid,
  import_batch_id uuid REFERENCES lead_lifecycle.lead_import_batches(id),
  system_lead_no text NOT NULL,
  student_name text,
  phone_hash text NOT NULL,
  phone_masked text,
  source_channel text,
  stage text NOT NULL DEFAULT 'new',
  raw jsonb NOT NULL DEFAULT '{}'::jsonb,
  reported_at timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, team_id, phone_hash)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_leads_system_no ON lead_lifecycle.leads(tenant_id, system_lead_no);
CREATE INDEX IF NOT EXISTS idx_leads_phone_hash ON lead_lifecycle.leads(tenant_id, phone_hash);
CREATE INDEX IF NOT EXISTS idx_leads_org_reported ON lead_lifecycle.leads(tenant_id, organization_id, reported_at DESC);

