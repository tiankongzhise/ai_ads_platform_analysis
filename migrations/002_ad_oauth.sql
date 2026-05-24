CREATE SCHEMA IF NOT EXISTS ad_sync;

CREATE TABLE IF NOT EXISTS ad_sync.ad_oauth_tokens (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL,
  platform text NOT NULL,
  external_account_id text NOT NULL,
  token_ref text NOT NULL UNIQUE,
  access_token_ciphertext text NOT NULL,
  refresh_token_ciphertext text,
  access_token_expires_at timestamptz,
  refresh_token_expires_at timestamptz,
  scope_list jsonb NOT NULL DEFAULT '[]'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, platform, external_account_id)
);

CREATE TABLE IF NOT EXISTS ad_sync.ad_oauth_callback_logs (
  id uuid PRIMARY KEY,
  tenant_id uuid,
  platform text NOT NULL,
  state_hash text NOT NULL,
  result text NOT NULL CHECK (result IN ('success', 'failed')),
  error_summary jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS ad_sync.ad_platform_accounts (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL,
  organization_id uuid,
  channel_id uuid,
  platform text NOT NULL,
  external_account_id text NOT NULL,
  account_name text,
  status text NOT NULL DEFAULT 'active',
  token_ref text,
  last_sync_at timestamptz,
  last_sync_status text,
  last_error jsonb NOT NULL DEFAULT '{}'::jsonb,
  meta jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, platform, external_account_id)
);

CREATE TABLE IF NOT EXISTS ad_sync.ad_sync_jobs (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL,
  account_id uuid NOT NULL,
  platform text NOT NULL,
  report_type text NOT NULL,
  date_from date NOT NULL,
  date_to date NOT NULL,
  status text NOT NULL CHECK (status IN ('pending', 'running', 'success', 'failed', 'cancelled')),
  retry_count int NOT NULL DEFAULT 0,
  request_meta jsonb NOT NULL DEFAULT '{}'::jsonb,
  result_meta jsonb NOT NULL DEFAULT '{}'::jsonb,
  error_meta jsonb NOT NULL DEFAULT '{}'::jsonb,
  started_at timestamptz,
  finished_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS ad_sync.ad_entities (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL,
  account_id uuid NOT NULL,
  platform text NOT NULL,
  entity_type text NOT NULL CHECK (entity_type IN ('account', 'campaign', 'adgroup', 'ad')),
  external_id text NOT NULL,
  parent_external_id text,
  name text,
  status text,
  raw jsonb NOT NULL DEFAULT '{}'::jsonb,
  synced_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (account_id, platform, entity_type, external_id)
);

CREATE INDEX IF NOT EXISTS idx_ad_entities_account_type ON ad_sync.ad_entities(account_id, entity_type);

CREATE TABLE IF NOT EXISTS ad_sync.ad_raw_report_rows (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL,
  account_id uuid NOT NULL,
  platform text NOT NULL,
  report_type text NOT NULL,
  granularity text NOT NULL CHECK (granularity IN ('day', 'hour')),
  stat_date date NOT NULL,
  stat_hour text,
  entity_type text NOT NULL,
  external_entity_id text NOT NULL,
  metrics jsonb NOT NULL DEFAULT '{}'::jsonb,
  raw jsonb NOT NULL DEFAULT '{}'::jsonb,
  synced_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (account_id, report_type, stat_date, stat_hour, external_entity_id)
);

CREATE INDEX IF NOT EXISTS idx_ad_raw_reports_account_type_date ON ad_sync.ad_raw_report_rows(account_id, report_type, stat_date DESC);

CREATE TABLE IF NOT EXISTS ad_sync.pgmq_messages (
  id bigserial PRIMARY KEY,
  queue_name text NOT NULL,
  message_id uuid NOT NULL UNIQUE,
  payload jsonb NOT NULL,
  status text NOT NULL CHECK (status IN ('ready', 'processing', 'done', 'failed')),
  retry_count int NOT NULL DEFAULT 0,
  next_visible_at timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_ad_sync_pgmq_ready ON ad_sync.pgmq_messages(queue_name, status, next_visible_at);
