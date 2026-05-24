CREATE SCHEMA IF NOT EXISTS data_insight;

CREATE TABLE IF NOT EXISTS data_insight.etl_batches (
  id uuid PRIMARY KEY,
  status text NOT NULL CHECK (status IN ('running', 'success', 'failed')),
  source text NOT NULL,
  input_rows int NOT NULL DEFAULT 0,
  fact_ad_daily_rows int NOT NULL DEFAULT 0,
  fact_lead_daily_rows int NOT NULL DEFAULT 0,
  hourly_rows int NOT NULL DEFAULT 0,
  rule_version text NOT NULL,
  metric_version text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS data_insight.fact_ad_daily (
  id uuid PRIMARY KEY,
  tenant_id text NOT NULL,
  report_date date NOT NULL,
  platform text NOT NULL,
  account_id text,
  entity_type text,
  external_entity_id text,
  spend numeric(14,2) NOT NULL DEFAULT 0,
  impressions bigint NOT NULL DEFAULT 0,
  clicks bigint NOT NULL DEFAULT 0,
  conversions bigint NOT NULL DEFAULT 0,
  source_report_type text NOT NULL,
  etl_batch_id uuid REFERENCES data_insight.etl_batches(id),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS data_insight.fact_lead_daily (
  id uuid PRIMARY KEY,
  tenant_id text NOT NULL,
  report_date date NOT NULL,
  organization_id text,
  team_id text,
  channel_text text,
  leads_count int NOT NULL DEFAULT 0,
  valid_leads_count int NOT NULL DEFAULT 0,
  visit_count int NOT NULL DEFAULT 0,
  enroll_count int NOT NULL DEFAULT 0,
  deal_amount numeric(14,2) NOT NULL DEFAULT 0,
  conflict_count int NOT NULL DEFAULT 0,
  etl_batch_id uuid REFERENCES data_insight.etl_batches(id),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS data_insight.fact_hourly_aggregate (
  id uuid PRIMARY KEY,
  tenant_id text NOT NULL,
  hour_start timestamptz NOT NULL,
  organization_id text,
  team_id text,
  platform text,
  channel_id text,
  metrics jsonb NOT NULL DEFAULT '{}'::jsonb,
  etl_batch_id uuid REFERENCES data_insight.etl_batches(id),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_fact_ad_daily_tenant_date ON data_insight.fact_ad_daily(tenant_id, report_date DESC);
CREATE INDEX IF NOT EXISTS idx_fact_lead_daily_tenant_date ON data_insight.fact_lead_daily(tenant_id, report_date DESC);
CREATE INDEX IF NOT EXISTS idx_fact_hourly_tenant_hour ON data_insight.fact_hourly_aggregate(tenant_id, hour_start DESC);

