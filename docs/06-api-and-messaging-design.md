# 教育广告 CRM 数据分析平台 API 与消息设计

## 1. 文档目的

本文档定义平台 REST API 分组、鉴权接入约定、错误模型、幂等机制、PGMQ 队列、事件消息结构、重试策略和核心状态流转。

## 2. API 设计原则

- API 使用 JSON over HTTPS。
- API 网关负责 TLS、路由、限流和请求日志。
- 业务服务通过已完成鉴权能力获取用户、租户、角色和组织范围。
- 所有写操作应支持幂等键或业务唯一约束。
- API 返回稳定错误码，便于前端统一处理。
- 外部 API 与内部服务 API 分开管理。

## 3. 通用请求头

| Header | 说明 |
| --- | --- |
| `Authorization` | 由已完成鉴权能力签发的访问凭证。 |
| `X-Tenant-Id` | 当前租户 ID，必须与鉴权上下文一致。 |
| `X-Request-Id` | 请求追踪 ID，缺省由网关生成。 |
| `Idempotency-Key` | 写操作幂等键，可选但推荐。 |
| `X-Org-Scope` | 当前操作组织范围，可由前端选择，也可由后端根据权限推导。 |

## 4. 通用响应格式

```json
{
  "success": true,
  "data": {},
  "request_id": "req_20260519_xxx"
}
```

错误响应：

```json
{
  "success": false,
  "error": {
    "code": "LEAD_DUPLICATED_IN_TEAM",
    "message": "当前团队已存在相同手机号线索",
    "details": {}
  },
  "request_id": "req_20260519_xxx"
}
```

## 5. API 分组

### 5.1 组织 API

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/orgs/tree` | 获取当前租户组织树。 |
| `POST` | `/api/orgs` | 创建组织。 |
| `PATCH` | `/api/orgs/{org_id}` | 更新组织。 |
| `POST` | `/api/orgs/{org_id}/move` | 调整组织父级。 |
| `GET` | `/api/orgs/{org_id}/summary` | 获取组织汇总信息。 |

### 5.2 团队 API

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/teams` | 查询团队列表。 |
| `POST` | `/api/teams` | 创建团队。 |
| `PATCH` | `/api/teams/{team_id}` | 更新团队。 |
| `POST` | `/api/teams/{team_id}/members` | 添加团队成员。 |
| `DELETE` | `/api/teams/{team_id}/members/{user_id}` | 移除成员。 |

### 5.3 渠道与广告账户 API

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/channels` | 查询广告渠道配置。 |
| `POST` | `/api/channels` | 创建渠道。 |
| `PATCH` | `/api/channels/{channel_id}` | 更新渠道。 |
| `GET` | `/api/ad-accounts` | 查询广告平台账户。 |
| `POST` | `/api/ad-accounts/{account_id}/sync` | 手动触发同步。 |
| `GET` | `/api/ad-sync/jobs` | 查询同步任务。 |

广告 OAuth 回调已完成，本系统只在回调成功后通过内部接口或消息接收账户绑定结果。

### 5.4 线索 API

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/api/leads/imports` | 创建 Excel 导入批次。 |
| `POST` | `/api/leads/imports/{batch_id}/upload` | 上传文件。 |
| `GET` | `/api/leads/imports/{batch_id}` | 查询导入状态。 |
| `POST` | `/api/leads/imports/{batch_id}/confirm` | 确认字段映射并开始导入。 |
| `POST` | `/api/leads` | API 上报单条线索。 |
| `POST` | `/api/leads/batch` | API 批量上报线索。 |
| `GET` | `/api/leads` | 查询线索列表。 |
| `PATCH` | `/api/leads/{lead_id}` | 更新线索阶段或补充信息。 |

### 5.5 冲突与归属 API

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/conflicts` | 查询冲突组列表。 |
| `GET` | `/api/conflicts/{group_id}` | 查询冲突详情。 |
| `POST` | `/api/conflicts/{group_id}/resolve` | 选择主咨询并解决冲突。 |
| `GET` | `/api/attributions/rules` | 查询归属规则配置。 |
| `POST` | `/api/attributions/calculate` | 触发归属重算。 |

### 5.6 BI 与分析 API

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/analytics/group-overview` | 集团总览。 |
| `GET` | `/api/analytics/school-compare` | 学校对比。 |
| `GET` | `/api/analytics/team-efficiency` | 团队效率。 |
| `GET` | `/api/analytics/channel-roi` | 渠道 ROI。 |
| `GET` | `/api/analytics/funnel` | 招生漏斗。 |
| `GET` | `/api/analytics/conflicts` | 线索冲突分析。 |
| `GET` | `/api/analytics/hourly-trend` | 小时趋势。 |

### 5.7 BI 配置 API

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/bi/catalog/metrics` | 指标目录。 |
| `GET` | `/api/bi/catalog/dimensions` | 维度目录。 |
| `GET` | `/api/bi/catalog/datasets` | 数据集目录。 |
| `GET` | `/api/bi/catalog/charts` | 图表注册表。 |
| `GET` | `/api/bi/dashboards` | 看板列表。 |
| `POST` | `/api/bi/dashboards` | 创建看板。 |
| `PATCH` | `/api/bi/dashboards/{dashboard_id}` | 更新看板布局和筛选。 |
| `POST` | `/api/bi/extensions/{extension_code}/enable` | 启用租户扩展包。 |

### 5.8 报表 API

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/api/reports` | 创建报表生成任务。 |
| `GET` | `/api/reports` | 查询报表列表。 |
| `GET` | `/api/reports/{report_id}` | 查询报表详情。 |
| `GET` | `/api/reports/{report_id}/download` | 下载报表文件。 |
| `GET` | `/api/reports/{report_id}/insights` | 查询行动建议。 |

### 5.9 虚拟演示账户 API

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/api/virtual-accounts` | 创建虚拟演示账户。 |
| `GET` | `/api/virtual-accounts/{tenant_id}` | 查看虚拟账户状态。 |
| `POST` | `/api/virtual-accounts/{tenant_id}/destroy` | 手动销毁。 |
| `POST` | `/api/virtual-accounts/{tenant_id}/migration-authorizations` | 授权迁移到正式账户。 |
| `POST` | `/api/virtual-accounts/{tenant_id}/migrate` | 执行迁移。 |
| `GET` | `/api/virtual-accounts/{tenant_id}/cleanup-audits` | 查询清理审计。 |

## 6. 幂等设计

### 6.1 API 幂等

以下操作必须支持幂等：

- API 线索上报。
- Excel 导入确认。
- 手动广告同步。
- 冲突裁定。
- 报表生成。
- 虚拟账户销毁。
- 虚拟账户迁移。

幂等键优先使用 `Idempotency-Key`，同时结合业务唯一约束。例如团队内线索去重使用 `(tenant_id, team_id, phone_hash)`。

### 6.2 消息幂等

消息消费者必须按 `message_id` 或业务幂等键记录处理结果。重复消息不得导致重复写入、重复报表或重复清理。

## 7. PGMQ 队列设计

| 队列 | 生产者 | 消费者 | 用途 |
| --- | --- | --- | --- |
| `lead_ingest_queue` | lead API / import service | lead-ingestion-service | 线索导入和 API 上报处理。 |
| `lead_conflict_queue` | lead-ingestion-service | conflict-analysis-service | 冲突识别。 |
| `ad_sync_queue` | callback-adapter / scheduler | ad-sync-service | 广告数据同步。 |
| `etl_queue` | ad-sync / lead / attribution | etl-service | 原始数据标准化。 |
| `hourly_aggregate_queue` | scheduler / etl-service | etl-service | 小时聚合。 |
| `bi_refresh_queue` | etl-service | bi-metric-service | BI 指标刷新。 |
| `report_generate_queue` | report API / scheduler | report-service | 报表生成。 |
| `virtual_account_cleanup_queue` | virtual-account-service / scheduler | virtual-account-service | 虚拟账户清理。 |
| `virtual_account_migration_queue` | virtual-account-service | virtual-account-service | 虚拟账户迁移。 |

## 8. 事件消息格式

```json
{
  "message_id": "uuid",
  "schema_version": "1.0",
  "event_type": "ad.sync.requested",
  "tenant_id": "uuid",
  "trace_id": "req_20260519_xxx",
  "occurred_at": "2026-05-19T12:00:00Z",
  "producer": "callback-adapter-service",
  "payload": {}
}
```

### 8.1 ad.sync.requested

```json
{
  "account_id": "uuid",
  "platform": "douyin",
  "report_types": ["campaign_daily", "adgroup_daily"],
  "date_from": "2026-05-12",
  "date_to": "2026-05-19",
  "reason": "oauth_callback"
}
```

### 8.2 lead.ingested

```json
{
  "lead_id": "uuid",
  "team_id": "uuid",
  "organization_id": "uuid",
  "phone_hash": "sha256",
  "reported_at": "2026-05-19T12:00:00Z"
}
```

### 8.3 virtual_account.cleanup.requested

```json
{
  "virtual_tenant_id": "uuid",
  "reason": "expired",
  "requested_by": "system",
  "delete_business_data": true
}
```

## 9. 重试与失败处理

- 可重试错误：网络超时、平台限流、数据库临时连接失败。
- 不可重试错误：权限不足、账户被禁用、参数非法、数据格式永久错误。
- 消息默认最多重试 3 次，平台同步可按平台规则增加指数退避。
- 超过重试次数进入失败状态，并生成告警事件。
- 虚拟账户清理失败必须可重试，直至全部业务资源清理完成或人工介入。

## 10. 核心状态流转

### 10.1 广告同步任务

```text
pending → running → success
                  ↘ failed → retrying → running
                  ↘ cancelled
```

### 10.2 线索导入批次

```text
created → uploaded → mapping_confirmed → processing → success
                                              ↘ partial_failed
                                              ↘ failed
```

### 10.3 冲突组

```text
open → resolved
open → ignored
resolved → reopened
```

### 10.4 虚拟账户

```text
active → expiring → destroying → destroyed
active → migration_authorized → migrating → migrated → destroying → destroyed
```

## 11. 鉴权接入约定

- 登录、注册、token 刷新由既有鉴权能力负责。
- 本系统通过鉴权中间件获得 `tenant_id`、`user_id`、角色和组织范围。
- 服务不得信任前端传入的租户和组织范围，必须与鉴权上下文校验。
- 回调适配服务只消费已完成广告回调能力提供的可信回调结果。
