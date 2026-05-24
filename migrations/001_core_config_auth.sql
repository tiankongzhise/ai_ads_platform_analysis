CREATE SCHEMA IF NOT EXISTS config;
CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS audit;

CREATE TABLE IF NOT EXISTS config.config_definitions (
  id uuid PRIMARY KEY,
  config_key text NOT NULL UNIQUE,
  value_type text NOT NULL CHECK (value_type IN ('string', 'duration', 'bool', 'int', 'secret_ref')),
  default_value text NOT NULL,
  editable boolean NOT NULL DEFAULT true,
  sensitive boolean NOT NULL DEFAULT false,
  description text NOT NULL,
  validation jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS config.config_values (
  id uuid PRIMARY KEY,
  config_key text NOT NULL REFERENCES config.config_definitions(config_key),
  scope_type text NOT NULL DEFAULT 'platform',
  scope_id uuid,
  value text NOT NULL,
  version integer NOT NULL,
  status text NOT NULL CHECK (status IN ('draft', 'published', 'rolled_back')),
  published_by uuid,
  published_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (config_key, scope_type, scope_id, version)
);

CREATE TABLE IF NOT EXISTS config.config_change_logs (
  id uuid PRIMARY KEY,
  config_key text NOT NULL,
  scope_type text NOT NULL,
  scope_id uuid,
  old_value text,
  new_value text,
  changed_by uuid,
  action text NOT NULL,
  version integer NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS auth.tenants (
  id uuid PRIMARY KEY,
  name text NOT NULL,
  tenant_type text NOT NULL CHECK (tenant_type IN ('formal', 'virtual_demo')),
  status text NOT NULL CHECK (status IN ('active', 'suspended', 'expired', 'destroying', 'destroyed')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS auth.users (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES auth.tenants(id),
  email text NOT NULL,
  phone text,
  display_name text NOT NULL,
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, email)
);

CREATE TABLE IF NOT EXISTS auth.password_credentials (
  user_id uuid PRIMARY KEY REFERENCES auth.users(id),
  password_hash text NOT NULL,
  password_algo text NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS auth.refresh_tokens (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES auth.tenants(id),
  user_id uuid NOT NULL REFERENCES auth.users(id),
  token_hash text NOT NULL UNIQUE,
  expires_at timestamptz NOT NULL,
  revoked_at timestamptz,
  rotation_parent_id uuid,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS auth.organizations (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES auth.tenants(id),
  parent_id uuid REFERENCES auth.organizations(id),
  name text NOT NULL,
  org_type text NOT NULL CHECK (org_type IN ('group', 'region', 'school', 'campus', 'department', 'other')),
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_auth_organizations_tenant_parent ON auth.organizations(tenant_id, parent_id);

CREATE TABLE IF NOT EXISTS auth.admission_teams (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES auth.tenants(id),
  organization_id uuid NOT NULL REFERENCES auth.organizations(id),
  name text NOT NULL,
  leader_user_id uuid,
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_auth_admission_teams_tenant_org ON auth.admission_teams(tenant_id, organization_id);

CREATE TABLE IF NOT EXISTS auth.ad_channels (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES auth.tenants(id),
  organization_id uuid NOT NULL REFERENCES auth.organizations(id),
  platform text NOT NULL CHECK (platform IN ('douyin', 'tencent', 'baidu', 'xiaohongshu', 'other')),
  display_name text NOT NULL,
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_auth_ad_channels_tenant_org ON auth.ad_channels(tenant_id, organization_id);

CREATE TABLE IF NOT EXISTS audit.audit_logs (
  id uuid PRIMARY KEY,
  tenant_id uuid,
  user_id uuid,
  action text NOT NULL,
  resource_type text NOT NULL,
  resource_id uuid,
  request_id text,
  detail jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);
