# 教育广告 CRM 数据分析平台数据库设计

## 1. 文档目的

本文档定义 PostgreSQL 数据库设计原则、核心表结构、`jsonb` 使用策略、多租户隔离、组织树、线索冲突、广告原始数据、ETL 标准层、BI 元数据和虚拟账户生命周期数据模型。

## 2. 设计原则

- PostgreSQL 作为唯一主数据库。
- 所有租户业务表必须包含 `tenant_id`。
- 平台原始报表、导入原始字段和扩展配置使用 `jsonb` 保真存储。
- 强结构业务对象使用关系表和外键约束。
- 组织树采用邻接表为主，必要时增加闭包表或物化路径加速查询。
- ETL 后的数据进入标准事实表和维度表。
- BI 配置通过元数据表管理，不允许用户直接 SQL。
- 虚拟账户业务数据必须能按 `tenant_id` 和账户类型彻底清理。

## 3. 通用字段规范

核心业务表建议包含：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `uuid` | 主键。 |
| `tenant_id` | `uuid` | 租户 ID。 |
| `created_at` | `timestamptz` | 创建时间。 |
| `updated_at` | `timestamptz` | 更新时间。 |
| `deleted_at` | `timestamptz null` | 软删除时间，必要表使用。 |
| `created_by` | `uuid null` | 创建人。 |
| `updated_by` | `uuid null` | 更新人。 |

## 4. 租户与组织

### 4.1 tenants

```sql
CREATE TABLE tenants (
  id uuid PRIMARY KEY,
  name text NOT NULL,
  tenant_type text NOT NULL CHECK (tenant_type IN ('formal', 'virtual_demo')),
  status text NOT NULL CHECK (status IN ('active', 'suspended', 'expired', 'destroying', 'destroyed')),
  package_code text,
  meta jsonb NOT NULL DEFAULT '{}'::jsonb,
  expires_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
```

### 4.2 organizations

```sql
CREATE TABLE organizations (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  parent_id uuid REFERENCES organizations(id),
  name text NOT NULL,
  org_type text NOT NULL CHECK (org_type IN ('group', 'region', 'school', 'campus', 'department', 'other')),
  path ltree,
  sort_order int NOT NULL DEFAULT 0,
  status text NOT NULL DEFAULT 'active',
  meta jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_organizations_tenant_parent ON organizations(tenant_id, parent_id);
CREATE INDEX idx_organizations_path ON organizations USING gist(path);
```

`path` 可使用 PostgreSQL `ltree` 扩展。如果部署环境暂不启用 `ltree`，可使用 `text` 物化路径替代。

## 5. 用户、角色与权限引用

鉴权已完成，本系统只保留必要引用和业务范围映射。

### 5.1 user_org_scopes

```sql
CREATE TABLE user_org_scopes (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  user_id uuid NOT NULL,
  organization_id uuid NOT NULL REFERENCES organizations(id),
  role_code text NOT NULL,
  scope_type text NOT NULL CHECK (scope_type IN ('self', 'subtree')),
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, user_id, organization_id, role_code)
);
```

## 6. 团队与渠道

### 6.1 admission_teams

```sql
CREATE TABLE admission_teams (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  organization_id uuid NOT NULL REFERENCES organizations(id),
  name text NOT NULL,
  leader_user_id uuid,
  status text NOT NULL DEFAULT 'active',
  meta jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
```

### 6.2 team_members

```sql
CREATE TABLE team_members (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  team_id uuid NOT NULL REFERENCES admission_teams(id),
  user_id uuid NOT NULL,
  member_role text NOT NULL DEFAULT 'member',
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, team_id, user_id)
);
```

### 6.3 ad_channels

```sql
CREATE TABLE ad_channels (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  organization_id uuid NOT NULL REFERENCES organizations(id),
  platform text NOT NULL CHECK (platform IN ('douyin', 'tencent', 'baidu', 'xiaohongshu', 'other')),
  display_name text NOT NULL,
  status text NOT NULL DEFAULT 'active',
  config jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
```

## 7. 广告账户与原始报表

### 7.1 ad_platform_accounts

```sql
CREATE TABLE ad_platform_accounts (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  organization_id uuid REFERENCES organizations(id),
  channel_id uuid REFERENCES ad_channels(id),
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
```

`token_ref` 指向已加密 token 存储或密钥服务引用，不直接暴露 token 明文。

### 7.2 ad_sync_jobs

```sql
CREATE TABLE ad_sync_jobs (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  account_id uuid NOT NULL REFERENCES ad_platform_accounts(id),
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
```

### 7.3 ad_raw_reports

```sql
CREATE TABLE ad_raw_reports (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  account_id uuid NOT NULL REFERENCES ad_platform_accounts(id),
  sync_job_id uuid REFERENCES ad_sync_jobs(id),
  platform text NOT NULL,
  report_type text NOT NULL,
  report_date date NOT NULL,
  external_entity_type text,
  external_entity_id text,
  raw_payload jsonb NOT NULL,
  normalized_key text,
  payload_hash text NOT NULL,
  fetched_at timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, platform, report_type, report_date, payload_hash)
);

CREATE INDEX idx_ad_raw_reports_payload ON ad_raw_reports USING gin(raw_payload);
CREATE INDEX idx_ad_raw_reports_lookup ON ad_raw_reports(tenant_id, platform, report_type, report_date);
```

## 8. 线索与导入

### 8.1 lead_import_batches

```sql
CREATE TABLE lead_import_batches (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  organization_id uuid REFERENCES organizations(id),
  team_id uuid REFERENCES admission_teams(id),
  source_type text NOT NULL CHECK (source_type IN ('excel', 'api')),
  file_ref text,
  status text NOT NULL DEFAULT 'pending',
  total_rows int NOT NULL DEFAULT 0,
  success_rows int NOT NULL DEFAULT 0,
  duplicate_rows int NOT NULL DEFAULT 0,
  failed_rows int NOT NULL DEFAULT 0,
  mapping_config jsonb NOT NULL DEFAULT '{}'::jsonb,
  error_report_ref text,
  meta jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
```

### 8.2 leads

```sql
CREATE TABLE leads (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  organization_id uuid NOT NULL REFERENCES organizations(id),
  team_id uuid NOT NULL REFERENCES admission_teams(id),
  import_batch_id uuid REFERENCES lead_import_batches(id),
  system_lead_no text NOT NULL,
  student_name text,
  phone_hash text NOT NULL,
  phone_masked text,
  source_channel_text text,
  course_name text,
  current_stage text NOT NULL DEFAULT 'new',
  reported_at timestamptz NOT NULL,
  first_visit_at timestamptz,
  first_enroll_at timestamptz,
  deal_amount numeric(14,2),
  raw_fields jsonb NOT NULL DEFAULT '{}'::jsonb,
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, team_id, phone_hash)
);

CREATE UNIQUE INDEX idx_leads_system_no ON leads(tenant_id, system_lead_no);
CREATE INDEX idx_leads_phone_hash ON leads(tenant_id, phone_hash);
CREATE INDEX idx_leads_org_reported ON leads(tenant_id, organization_id, reported_at);
```

## 9. 冲突与归属

### 9.1 lead_conflict_groups

```sql
CREATE TABLE lead_conflict_groups (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  phone_hash text NOT NULL,
  status text NOT NULL CHECK (status IN ('open', 'resolved', 'ignored')),
  primary_lead_id uuid REFERENCES leads(id),
  conflict_count int NOT NULL DEFAULT 0,
  resolved_by uuid,
  resolved_at timestamptz,
  resolution_meta jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, phone_hash)
);
```

### 9.2 lead_conflict_items

```sql
CREATE TABLE lead_conflict_items (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  conflict_group_id uuid NOT NULL REFERENCES lead_conflict_groups(id),
  lead_id uuid NOT NULL REFERENCES leads(id),
  team_id uuid NOT NULL REFERENCES admission_teams(id),
  is_primary boolean NOT NULL DEFAULT false,
  conflict_role text NOT NULL DEFAULT 'candidate',
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, conflict_group_id, lead_id)
);
```

### 9.3 lead_attributions

```sql
CREATE TABLE lead_attributions (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  lead_id uuid NOT NULL REFERENCES leads(id),
  attribution_rule text NOT NULL CHECK (attribution_rule IN ('first_report', 'first_visit', 'first_enroll', 'manual')),
  owner_team_id uuid REFERENCES admission_teams(id),
  owner_organization_id uuid REFERENCES organizations(id),
  evidence jsonb NOT NULL DEFAULT '{}'::jsonb,
  decided_by uuid,
  decided_at timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now()
);
```

## 10. ETL 标准层

### 10.1 dim_date、dim_organization、dim_team、dim_channel、dim_course

维度表保存标准化分析维度。组织、团队和渠道维度可以从业务表定时快照，避免历史组织变更影响历史报表口径。

### 10.2 fact_ad_daily

```sql
CREATE TABLE fact_ad_daily (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  report_date date NOT NULL,
  organization_id uuid,
  team_id uuid,
  channel_id uuid,
  account_id uuid,
  platform text NOT NULL,
  campaign_id text,
  campaign_name text,
  adgroup_id text,
  adgroup_name text,
  spend numeric(14,2) NOT NULL DEFAULT 0,
  impressions bigint NOT NULL DEFAULT 0,
  clicks bigint NOT NULL DEFAULT 0,
  conversions bigint NOT NULL DEFAULT 0,
  source_raw_report_id uuid REFERENCES ad_raw_reports(id),
  etl_batch_id uuid,
  created_at timestamptz NOT NULL DEFAULT now()
);
```

### 10.3 fact_lead_daily

```sql
CREATE TABLE fact_lead_daily (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  report_date date NOT NULL,
  organization_id uuid NOT NULL,
  team_id uuid,
  channel_text text,
  course_name text,
  leads_count int NOT NULL DEFAULT 0,
  valid_leads_count int NOT NULL DEFAULT 0,
  visit_count int NOT NULL DEFAULT 0,
  enroll_count int NOT NULL DEFAULT 0,
  deal_amount numeric(14,2) NOT NULL DEFAULT 0,
  conflict_count int NOT NULL DEFAULT 0,
  etl_batch_id uuid,
  created_at timestamptz NOT NULL DEFAULT now()
);
```

### 10.4 fact_hourly_aggregate

```sql
CREATE TABLE fact_hourly_aggregate (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  hour_start timestamptz NOT NULL,
  organization_id uuid,
  team_id uuid,
  platform text,
  channel_id uuid,
  metrics jsonb NOT NULL DEFAULT '{}'::jsonb,
  etl_batch_id uuid,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, hour_start, organization_id, team_id, platform, channel_id)
);
```

## 11. BI 元数据

### 11.1 bi_datasets

```sql
CREATE TABLE bi_datasets (
  id uuid PRIMARY KEY,
  dataset_code text NOT NULL,
  name text NOT NULL,
  description text,
  source_type text NOT NULL CHECK (source_type IN ('table', 'view', 'materialized_view', 'service')),
  source_ref text NOT NULL,
  schema_def jsonb NOT NULL,
  version text NOT NULL,
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (dataset_code, version)
);
```

### 11.2 bi_metrics

```sql
CREATE TABLE bi_metrics (
  id uuid PRIMARY KEY,
  metric_code text NOT NULL,
  name text NOT NULL,
  definition text NOT NULL,
  dataset_code text NOT NULL,
  formula jsonb NOT NULL,
  unit text,
  version text NOT NULL,
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (metric_code, version)
);
```

### 11.3 bi_dashboards

```sql
CREATE TABLE bi_dashboards (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  dashboard_code text NOT NULL,
  name text NOT NULL,
  owner_user_id uuid,
  layout jsonb NOT NULL DEFAULT '{}'::jsonb,
  filters jsonb NOT NULL DEFAULT '{}'::jsonb,
  permission_rules jsonb NOT NULL DEFAULT '{}'::jsonb,
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
```

### 11.4 tenant_bi_extensions

```sql
CREATE TABLE tenant_bi_extensions (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  extension_code text NOT NULL,
  version text NOT NULL,
  status text NOT NULL CHECK (status IN ('enabled', 'disabled', 'rollback_pending')),
  manifest jsonb NOT NULL,
  enabled_at timestamptz,
  disabled_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, extension_code, version)
);
```

## 12. 虚拟账户生命周期

### 12.1 virtual_account_lifecycle

```sql
CREATE TABLE virtual_account_lifecycle (
  tenant_id uuid PRIMARY KEY REFERENCES tenants(id),
  created_by uuid NOT NULL,
  expires_at timestamptz NOT NULL,
  destroyed_at timestamptz,
  destroy_reason text,
  migration_status text,
  migration_target_tenant_id uuid,
  lifecycle_meta jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
```

### 12.2 virtual_account_cleanup_audits

```sql
CREATE TABLE virtual_account_cleanup_audits (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL,
  cleanup_task_id uuid NOT NULL,
  cleanup_status text NOT NULL CHECK (cleanup_status IN ('success', 'partial_failed', 'failed')),
  cleaned_resources jsonb NOT NULL DEFAULT '{}'::jsonb,
  error_summary jsonb NOT NULL DEFAULT '{}'::jsonb,
  business_data_retained boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now()
);
```

该表不得保存客户业务数据，只保存清理结果统计、资源类型和错误摘要。

## 13. 审计日志

```sql
CREATE TABLE audit_logs (
  id uuid PRIMARY KEY,
  tenant_id uuid,
  user_id uuid,
  action text NOT NULL,
  resource_type text NOT NULL,
  resource_id uuid,
  request_id text,
  ip_address inet,
  detail jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);
```

## 14. 数据清理策略

- 正式租户软删除后进入保留期，按合同与合规要求处理。
- 虚拟租户销毁时按 `tenant_id` 清理所有业务表、对象存储文件、Redis key 和未处理队列消息。
- 清理完成后仅保留 `virtual_account_cleanup_audits` 和必要系统日志。
- 清理任务必须可重试，并能证明无业务数据残留。
