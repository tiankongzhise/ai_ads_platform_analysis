CREATE SCHEMA IF NOT EXISTS virtual_account;

CREATE TABLE IF NOT EXISTS virtual_account.accounts (
  id uuid PRIMARY KEY,
  tenant_id text NOT NULL,
  owner_user_id text NOT NULL,
  display_name text NOT NULL,
  purpose text NOT NULL,
  status text NOT NULL CHECK (status IN ('active', 'migration_authorized', 'destroyed', 'cleanup_blocked')),
  valid_days int NOT NULL DEFAULT 7 CHECK (valid_days BETWEEN 1 AND 7),
  resources jsonb NOT NULL DEFAULT '{}'::jsonb,
  cleanup_verified boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now(),
  expires_at timestamptz NOT NULL,
  destroyed_at timestamptz,
  migration_authorized_at timestamptz,
  migration_authorized_by text,
  migration_target_user_id text
);

CREATE TABLE IF NOT EXISTS virtual_account.migration_authorizations (
  id uuid PRIMARY KEY,
  virtual_account_id uuid NOT NULL REFERENCES virtual_account.accounts(id),
  tenant_id text NOT NULL,
  target_user_id text NOT NULL,
  authorized_by text NOT NULL,
  reason text NOT NULL,
  status text NOT NULL CHECK (status IN ('authorized', 'used', 'expired', 'revoked')),
  created_at timestamptz NOT NULL DEFAULT now(),
  expires_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS virtual_account.cleanup_logs (
  id uuid PRIMARY KEY,
  virtual_account_id uuid NOT NULL REFERENCES virtual_account.accounts(id),
  tenant_id text NOT NULL,
  reason text NOT NULL,
  status text NOT NULL CHECK (status IN ('success', 'blocked', 'failed')),
  actor text NOT NULL,
  token_refs_deleted int NOT NULL DEFAULT 0,
  files_deleted int NOT NULL DEFAULT 0,
  cache_keys_deleted int NOT NULL DEFAULT 0,
  queue_messages_deleted int NOT NULL DEFAULT 0,
  business_records_found int NOT NULL DEFAULT 0,
  no_business_data boolean NOT NULL DEFAULT false,
  details jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_virtual_accounts_tenant_status ON virtual_account.accounts(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_virtual_accounts_expires ON virtual_account.accounts(expires_at, status);
CREATE INDEX IF NOT EXISTS idx_virtual_cleanup_logs_account ON virtual_account.cleanup_logs(virtual_account_id, created_at DESC);
