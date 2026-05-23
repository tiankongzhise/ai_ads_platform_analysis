# 教育广告 CRM 数据分析平台系统架构设计

## 1. 文档目的

本文档定义系统总体架构、服务边界、技术选型、数据流、部署拓扑和关键架构决策。本文档以微服务模块化开发为核心原则，服务后续详细设计、开发、测试和部署。

## 2. 架构原则

- 前后端分离：前端只通过 API 与后端交互。
- 微服务模块化：按业务能力拆分服务，允许独立部署和测试。
- Go 与 Python 分工明确：Go 负责轻量控制面，Python 负责复杂数据处理。
- PostgreSQL 为主数据存储，充分使用关系模型、事务、索引和 `jsonb`。
- Redis 负责缓存、短期状态和幂等控制。
- PGMQ 负责异步任务和服务间事件流转。
- 广告平台数据先保真入库，再通过 ETL 标准化。
- 鉴权与广告回调作为已完成能力接入。

## 3. 总体架构

```text
┌──────────────────────────────────────────────────────────────┐
│                          前端层                               │
│ React + TypeScript + Vite + Ant Design Pro + ECharts          │
│ Web 管理端 / BI 看板 / 演示账户入口                            │
└───────────────────────────┬──────────────────────────────────┘
                            │ HTTPS / JSON API
┌───────────────────────────▼──────────────────────────────────┐
│                        API 网关层                              │
│ Nginx / API Gateway / TLS / 限流 / 路由 / 请求日志              │
└───────────────────────────┬──────────────────────────────────┘
                            │
┌───────────────────────────▼──────────────────────────────────┐
│                       后端微服务层                             │
│ Go 控制面服务 + Python 数据处理服务                            │
└──────┬───────────────┬──────────────┬───────────────┬────────┘
       │               │              │               │
       ▼               ▼              ▼               ▼
┌────────────┐ ┌────────────┐ ┌──────────────┐ ┌──────────────┐
│ PostgreSQL │ │   Redis    │ │ PGMQ Queues  │ │ Object Store │
│ 主库+JSONB │ │ 缓存/状态  │ │ 异步消息     │ │ 文件/报表    │
└────────────┘ └────────────┘ └──────────────┘ └──────────────┘
       ▲               ▲              ▲
       │               │              │
┌──────┴──────────────────────────────┴───────────────────────┐
│                    外部和既有能力                             │
│ 已完成鉴权服务 / 已完成广告 OAuth 回调 / 广告平台 API           │
└──────────────────────────────────────────────────────────────┘
```

## 4. 服务划分

### 4.1 Go 服务

#### organization-service

负责租户、组织树、组织成员、组织范围权限和组织层级查询。

#### team-service

负责招生团队、团队成员、团队与组织关系、团队渠道配置。

#### channel-service

负责广告渠道配置、平台账户引用、渠道状态、同步策略基础配置。

#### config-service

负责系统配置、套餐配置、BI 元数据配置、图表注册表、租户扩展包启停。

#### permission-adapter-service

对接已完成鉴权能力，统一解析用户、租户、角色、组织范围和权限点。

#### callback-adapter-service

对接已完成广告平台回调能力，在回调成功后创建平台账户引用并投递广告同步消息。

#### health-service

提供各服务健康检查、版本信息、依赖检查和 readiness/liveness 端点。

### 4.2 Python 服务

#### lead-ingestion-service

负责 Excel/CSV 解析、字段映射、线索校验、团队内去重、API 线索批量接收和导入错误报告。

#### conflict-analysis-service

负责手机号标准化、跨团队冲突识别、冲突组创建、主咨询裁定辅助和冲突统计。

#### ad-sync-service

负责广告平台数据获取基座和平台适配器。每个平台独立适配，但共享同步任务、错误分类、限流、原始入库和状态记录框架。

#### etl-service

负责将广告原始报表、线索、冲突、归属和招生阶段数据转换为标准事实表、维度表和 BI 数据集。

#### attribution-service

负责渠道归因、线索归属规则计算、首次上报/首次到校/首次报名等裁定口径计算。

#### bi-metric-service

负责指标计算、数据集查询、BI 缓存刷新、标准教育模板数据输出。

#### report-service

负责报表生成、Excel/PDF 导出、行动建议生成和报表文件存储。

#### virtual-account-service

负责虚拟演示账户创建、过期检测、清理、迁移和清理审计。

## 5. 关键数据流

### 5.1 线索上报流程

```text
Excel/API 上报
  → API 网关
  → lead-ingestion-service
  → 写入导入批次和原始字段 jsonb
  → PGMQ: lead.ingested
  → conflict-analysis-service
  → 创建或更新冲突组
  → attribution-service
  → 更新归属候选
  → PGMQ: bi.refresh.requested
  → bi-metric-service 刷新 BI
```

### 5.2 广告数据同步流程

```text
广告 OAuth 回调已完成
  → callback-adapter-service 接收回调结果
  → 创建或更新 ad_platform_accounts
  → PGMQ: ad.sync.requested
  → ad-sync-service 调用平台 API
  → 原始报表写入 ad_raw_reports(jsonb)
  → PGMQ: etl.ad_raw.ready
  → etl-service 标准化
  → PGMQ: bi.refresh.requested
```

### 5.3 小时聚合流程

```text
定时调度或数据变更
  → PGMQ: aggregate.hourly.requested
  → etl-service 读取标准事实表
  → 写入 hourly_aggregate_fact
  → bi-metric-service 刷新热点指标
  → Redis 写入 BI 热点缓存
```

### 5.4 虚拟账户清理流程

```text
到期或用户手动销毁
  → virtual-account-service 创建清理任务
  → PGMQ: virtual_account.cleanup.requested
  → 清理业务表、token、文件、缓存、队列残留
  → 记录不含业务数据的清理审计
  → 标记虚拟账户 destroyed
```

## 6. 数据存储架构

### 6.1 PostgreSQL

用于保存：

- 租户、组织、团队、渠道、账户。
- 线索、冲突、归属裁定。
- 广告原始报表和 ETL 标准数据。
- BI 元数据、看板、图表配置。
- 虚拟账户生命周期和清理审计。
- PGMQ 队列表。

### 6.2 Redis

用于保存：

- 热点 BI 缓存。
- 导入和同步进度。
- OAuth 临时状态。
- 幂等键。
- 虚拟账户过期提示。
- 短期任务状态。

### 6.3 对象存储

用于保存：

- 上传的 Excel/CSV 文件。
- 导入错误报告。
- 导出的报表文件。
- 临时演示账户文件。

## 7. API 架构

外部 API 按业务域分组：

- `/api/orgs`
- `/api/teams`
- `/api/channels`
- `/api/leads`
- `/api/conflicts`
- `/api/ad-accounts`
- `/api/ad-sync`
- `/api/analytics`
- `/api/bi`
- `/api/reports`
- `/api/virtual-accounts`
- `/api/settings`

所有 API 请求应携带鉴权上下文。服务内部不自行实现登录注册，只消费已完成鉴权能力产生的用户与租户上下文。

## 8. 消息架构

PGMQ 队列按业务事件拆分：

- `lead_ingest_queue`
- `lead_conflict_queue`
- `ad_sync_queue`
- `etl_queue`
- `hourly_aggregate_queue`
- `bi_refresh_queue`
- `report_generate_queue`
- `virtual_account_cleanup_queue`
- `virtual_account_migration_queue`

消息必须包含 `message_id`、`tenant_id`、`trace_id`、`event_type`、`schema_version`、`occurred_at` 和业务载荷。

## 9. 部署拓扑

### 9.1 开发环境

- 本地 Docker Compose 启动 PostgreSQL、Redis、对象存储和核心服务。
- 服务可按模块单独启动。
- 使用测试广告平台 Mock 或沙箱环境。

### 9.2 测试环境

- 独立 PostgreSQL 和 Redis。
- 自动执行迁移和种子数据。
- 支持 API 契约测试、队列集成测试和前端 E2E 测试。

### 9.3 生产环境

- API 网关统一入口。
- Go 服务和 Python 服务分别部署。
- PostgreSQL 主从或云数据库高可用。
- Redis 使用持久化和高可用配置。
- 对象存储使用云服务或高可用 MinIO。
- 后台 worker 可按队列横向扩容。

## 10. 架构演进

第一阶段保持服务边界清晰但避免过细拆分。若后期数据量增长，可优先扩展：

- `ad-sync-service` 按平台拆分。
- `etl-service` 与 `bi-metric-service` 单独扩容。
- 高频 BI 查询引入只读副本或 OLAP 存储。
- PGMQ 热点队列迁移至专用消息队列。
