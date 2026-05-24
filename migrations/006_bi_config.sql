CREATE TABLE IF NOT EXISTS data_insight.bi_datasets (
  id uuid PRIMARY KEY,
  dataset_code text NOT NULL,
  name text NOT NULL,
  description text,
  source_type text NOT NULL CHECK (source_type IN ('table', 'view', 'materialized_view', 'service')),
  source_ref text NOT NULL,
  schema_def jsonb NOT NULL DEFAULT '{}'::jsonb,
  version text NOT NULL,
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (dataset_code, version)
);

CREATE TABLE IF NOT EXISTS data_insight.bi_metrics (
  id uuid PRIMARY KEY,
  metric_code text NOT NULL,
  name text NOT NULL,
  definition text NOT NULL,
  dataset_code text NOT NULL,
  formula jsonb NOT NULL DEFAULT '{}'::jsonb,
  unit text,
  version text NOT NULL,
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (metric_code, version)
);

CREATE TABLE IF NOT EXISTS data_insight.bi_dashboards (
  id uuid PRIMARY KEY,
  tenant_id text NOT NULL,
  dashboard_code text NOT NULL,
  name text NOT NULL,
  owner_user_id text,
  layout jsonb NOT NULL DEFAULT '{}'::jsonb,
  filters jsonb NOT NULL DEFAULT '{}'::jsonb,
  permission_rules jsonb NOT NULL DEFAULT '{}'::jsonb,
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS data_insight.tenant_bi_extensions (
  id uuid PRIMARY KEY,
  tenant_id text NOT NULL,
  extension_code text NOT NULL,
  version text NOT NULL,
  status text NOT NULL CHECK (status IN ('enabled', 'disabled', 'rollback_pending')),
  manifest jsonb NOT NULL DEFAULT '{}'::jsonb,
  enabled_at timestamptz,
  disabled_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, extension_code, version)
);

CREATE INDEX IF NOT EXISTS idx_bi_dashboards_tenant ON data_insight.bi_dashboards(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_tenant_bi_extensions_status ON data_insight.tenant_bi_extensions(tenant_id, status);

