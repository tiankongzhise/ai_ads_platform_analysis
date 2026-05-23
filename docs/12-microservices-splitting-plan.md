# 教育广告 CRM 数据分析平台微服务拆分方案

## 1. 文档目的

本文档回答“如果以微服务形式开发本项目，应拆分成哪些微服务”的问题，并给出拆分依据、目标服务清单、数据归属、通信方式、MVP 落地方式和后续演进路线。

本文档基于仓库现有文档：

- `readme.md`
- `docs/03-requirements-specification.md`
- `docs/04-system-architecture.md`
- `docs/05-database-design.md`
- `docs/06-api-and-messaging-design.md`
- `docs/07-detailed-module-design.md`
- `docs/10-deployment-and-operations.md`

## 2. 调研结论

### 2.1 外部资料结论

| 资料 | 关键观点 | 本项目采用方式 |
| --- | --- | --- |
| [Martin Fowler - Microservices](https://martinfowler.com/articles/microservices.html) | 微服务应是可独立部署的服务集合，围绕业务能力组织，使用轻量通信，并倾向去中心化数据管理。 | 不按 Controller、Repository、工具类等技术层拆服务，而按组织、线索、广告同步、ETL、BI、报表等业务能力拆分。 |
| [AWS Prescriptive Guidance - Decompose by business capability](https://docs.aws.amazon.com/prescriptive-guidance/latest/modernization-decomposing-monoliths/decompose-business-capability.html) | 按业务能力拆分能形成更稳定的微服务架构，但需要深入理解业务。 | 以教育广告招生闭环为主线，先识别业务能力，再决定服务边界。 |
| [Microsoft Azure Architecture Center - Use domain analysis to model microservices](https://learn.microsoft.com/en-us/azure/architecture/microservices/model/domain-analysis) | 微服务边界没有机械公式，应通过领域分析和限界上下文识别，保证高内聚、低耦合。 | 将本项目划分为租户组织、招生协作、广告集成、数据加工、BI 分析、演示账户生命周期等限界上下文。 |
| [Microsoft Azure Architecture Center - Data considerations for microservices](https://learn.microsoft.com/en-us/azure/architecture/microservices/design/data-considerations) | 每个微服务应管理自己的私有数据，避免多个服务共享同一数据表造成耦合。 | 第一阶段可共用 PostgreSQL 集群，但必须按服务拆 schema、账号和迁移边界，禁止跨服务直接读写表。 |
| [AWS Prescriptive Guidance - Database-per-service pattern](https://docs.aws.amazon.com/prescriptive-guidance/latest/modernization-data-persistence/database-per-service.html) | 数据库按服务隔离能降低耦合，但跨服务查询和事务需要 API Composition、CQRS、Saga 等模式配合。 | 广告、线索、冲突、归属、ETL、BI 等服务分别拥有数据；跨域查询通过 BI 读模型和事件驱动同步完成。 |

### 2.2 对本项目的核心判断

本项目不是一个简单 CRUD 后台，而是“广告数据接入 + 线索上报 + 冲突裁定 + 归因归属 + ETL + BI + 报表 + 虚拟演示账户”的业务闭环。适合采用微服务，但不适合第一版就把每张表、每个平台、每个接口都拆成独立服务。

推荐策略：

- 目标架构按业务能力拆成清晰微服务。
- MVP 阶段按“逻辑边界清晰、物理部署适度合并”的方式落地。
- 数据边界从第一天就按服务隔离，避免后期拆库困难。
- 高吞吐、异步、计算密集模块优先独立部署，例如广告同步、线索导入、ETL、BI、报表。
- 广告平台适配先放在 `ad-sync-service` 内部插件化，后期再按平台拆分。

## 3. 拆分原则

### 3.1 按业务能力拆分

服务边界围绕“业务对象 + 业务规则 + 数据所有权”划分，而不是围绕技术层划分。例如：

- 组织树和组织权限范围属于租户组织能力。
- 线索导入、手机号哈希、团队内去重属于线索接入能力。
- 跨团队重复手机号、主咨询裁定属于冲突处理能力。
- 平台 API 拉取、限流、原始报表入库属于广告同步能力。
- 指标计算、BI 查询、热点缓存属于 BI 指标能力。

### 3.2 Go 与 Python 分工

沿用现有技术约束：

- Go 服务负责控制面、配置类、权限适配、租户组织、团队、渠道、审计、调度等轻量高并发模块。
- Python 服务负责 Excel 解析、广告数据同步适配、冲突分析、归因、ETL、BI 指标、报表生成等数据处理模块。

### 3.3 数据归属优先

每个服务必须有明确的数据归属。其他服务不得直接访问其私有表，只能通过 API、事件或只读数据产品使用数据。

第一阶段可使用同一个 PostgreSQL 集群，但建议：

- 每个服务一个独立 schema。
- 每个服务一个独立数据库账号。
- 每个服务只执行自己的迁移脚本。
- 禁止业务代码跨 schema join。
- BI/ETL 跨域分析通过事件同步和标准事实表完成。

### 3.4 异步优先，API 粗粒度

线索导入、广告同步、ETL、小时聚合、BI 刷新、报表生成、虚拟账户清理都应走 PGMQ 异步队列。同步 API 只用于用户实时交互、配置读取、任务创建和状态查询。

### 3.5 避免过度拆分

以下拆法不推荐：

- 按 Controller、Service、Repository 分服务。
- 一张表一个服务。
- 每个小工具函数一个服务。
- 第一版就把抖音、腾讯、百度、小红书拆成四个独立广告同步服务。
- 多个服务共享同一个“common business library”并把核心业务规则放进去。

## 4. 领域上下文划分

| 限界上下文 | 核心业务问题 | 候选微服务 |
| --- | --- | --- |
| 租户与组织上下文 | 谁是租户，组织树如何划分，用户在什么组织范围内操作。 | `tenant-org-service`、`permission-adapter-service` |
| 招生协作上下文 | 招生团队如何管理，线索如何上报，重复咨询如何处理，归属如何裁定。 | `team-service`、`lead-ingestion-service`、`lead-conflict-service`、`attribution-service` |
| 广告集成上下文 | 广告账户如何接入，平台数据如何同步，平台差异如何保真保存。 | `channel-account-service`、`callback-adapter-service`、`ad-sync-service` |
| 数据加工上下文 | 原始广告、线索、冲突、归属如何转成标准事实和维度。 | `etl-aggregation-service` |
| BI 与商业化上下文 | 指标、维度、数据集、图表、看板和定制扩展如何管理和查询。 | `bi-config-service`、`bi-metric-service` |
| 报表上下文 | 用户如何生成、下载、追踪报表和行动建议。 | `report-service` |
| 虚拟演示账户上下文 | 演示账户如何创建、接入、迁移、销毁和验证清理。 | `virtual-account-service` |
| 平台治理上下文 | 调度、审计、健康检查、观测性、告警如何统一。 | `scheduler-service`、`audit-service`、服务内健康端点 |

## 5. 目标微服务清单

### 5.1 接入与平台支撑服务

| 服务 | 建议语言 | 职责 | 数据归属 | 说明 |
| --- | --- | --- | --- | --- |
| `api-gateway` / `bff-service` | Nginx / Go | TLS、路由、限流、请求日志、前端聚合 API。 | 不持有核心业务数据。 | 网关是平台入口，不建议承载业务规则。复杂前端聚合可放 BFF。 |
| `permission-adapter-service` | Go | 对接已完成鉴权能力，解析用户、租户、角色、组织范围和权限点。 | 权限缓存、鉴权上下文映射。 | 本项目不重复建设登录注册，只消费既有鉴权结果。 |
| `scheduler-service` | Go | 定时投递广告同步、小时聚合、BI 刷新、虚拟账户过期扫描等任务。 | 调度任务、执行记录。 | 避免把定时逻辑散落在多个服务。 |
| `audit-service` | Go | 记录敏感操作、配置变更、冲突裁定、虚拟账户迁移和清理审计索引。 | `audit_logs` 或审计专用 schema。 | 业务服务通过事件或 API 写审计，不直接写审计表。 |

### 5.2 控制面业务服务

| 服务 | 建议语言 | 职责 | 数据归属 | 主要 API / 事件 |
| --- | --- | --- | --- | --- |
| `tenant-org-service` | Go | 租户、组织树、组织移动、组织状态、组织范围查询。 | `tenants`、`organizations`、组织快照。 | `/api/orgs/*`；发布 `org.changed`。 |
| `team-service` | Go | 招生团队、团队成员、团队与组织关系、团队可用渠道。 | `admission_teams`、`team_members`、团队渠道关系。 | `/api/teams/*`；发布 `team.changed`。 |
| `channel-account-service` | Go | 广告渠道配置、平台账户引用、账户状态、同步策略基础配置。 | `ad_channels`、`ad_platform_accounts` 的账户引用部分。 | `/api/channels/*`、`/api/ad-accounts/*`；发布 `ad.account.bound`。 |
| `callback-adapter-service` | Go | 接收既有广告 OAuth 回调结果，校验租户和组织范围，创建或更新广告账户引用。 | 回调接入记录、回调幂等记录。 | 消费既有回调结果；发布 `ad.sync.requested`。 |
| `bi-config-service` | Go | 指标目录、维度目录、数据集目录、图表注册表、看板布局、租户 BI 扩展包启停。 | `bi_datasets`、`bi_metrics`、`bi_dashboards`、`tenant_bi_extensions`。 | `/api/bi/catalog/*`、`/api/bi/dashboards/*`。 |

### 5.3 数据面业务服务

| 服务 | 建议语言 | 职责 | 数据归属 | 主要 API / 事件 |
| --- | --- | --- | --- | --- |
| `lead-ingestion-service` | Python | Excel/CSV 解析、字段映射、API 线索上报、手机号标准化、团队内去重、导入错误报告。 | `lead_import_batches`、`leads`、导入文件引用。 | `/api/leads/*`；发布 `lead.ingested`、`lead.import.completed`。 |
| `lead-conflict-service` | Python | 跨团队冲突识别、冲突组维护、主咨询裁定、冲突统计。 | `lead_conflict_groups`、`lead_conflict_items`。 | `/api/conflicts/*`；消费 `lead.ingested`，发布 `lead.conflict.detected`、`lead.conflict.resolved`。 |
| `attribution-service` | Python | 首次上报、首次到校、首次报名、手动裁定等归属规则计算；渠道归因辅助。 | `lead_attributions`、归因规则版本。 | `/api/attributions/*`；消费 `lead.ingested`、`lead.conflict.resolved`，发布 `lead.attribution.updated`。 |
| `ad-sync-service` | Python | 广告平台 API 客户端、平台适配器、限流重试、同步任务、原始报表 `jsonb` 入库。 | `ad_sync_jobs`、`ad_raw_reports`、平台同步状态。 | 消费 `ad.sync.requested`；发布 `ad.raw.ready`、`ad.sync.failed`。 |
| `etl-aggregation-service` | Python | 广告原始报表、线索、冲突、归属数据标准化；维度快照；小时聚合。 | `fact_ad_daily`、`fact_lead_daily`、`fact_hourly_aggregate`、维度表、ETL 批次。 | 消费 `ad.raw.ready`、`lead.attribution.updated`；发布 `etl.batch.completed`、`bi.refresh.requested`。 |
| `bi-metric-service` | Python | 指标计算、数据集查询、标准教育 BI 模板数据输出、Redis 热点缓存刷新。 | BI 查询缓存、指标计算结果、读模型状态。 | `/api/analytics/*`、`/api/bi/query`；消费 `bi.refresh.requested`。 |
| `report-service` | Python | 报表任务、Excel/PDF 导出、行动建议、报表文件写对象存储。 | `report_jobs`、`report_files`、报表参数和指标版本。 | `/api/reports/*`；消费 `report.generate.requested`。 |
| `virtual-account-service` | Python / Go | 虚拟演示账户创建、过期、迁移、销毁、清理编排和清理验证。 | `virtual_account_lifecycle`、迁移授权、清理任务、清理审计。 | `/api/virtual-accounts/*`；发布 `virtual_account.cleanup.requested`。 |

## 6. 数据归属方案

| 数据域 | 归属服务 | 其他服务使用方式 |
| --- | --- | --- |
| 租户、组织树 | `tenant-org-service` | API 查询组织范围；通过 `org.changed` 更新读模型。 |
| 用户组织范围、权限上下文 | `permission-adapter-service` | API 或中间件调用；不复制权限业务规则。 |
| 招生团队、团队成员 | `team-service` | API 校验团队有效性；事件同步团队维度快照。 |
| 广告渠道、账户引用 | `channel-account-service` | `ad-sync-service` 通过 API 或事件获取账户同步上下文。 |
| 线索、导入批次 | `lead-ingestion-service` | 冲突、归因、ETL 通过事件或受控读取模型使用。 |
| 冲突组、冲突项 | `lead-conflict-service` | BI 和归因通过事件同步冲突结果。 |
| 归属裁定 | `attribution-service` | ETL 通过事件获取最新归属口径。 |
| 广告同步任务、原始报表 | `ad-sync-service` | ETL 消费 `ad.raw.ready` 后转换，不直接修改原始表。 |
| 标准事实表、维度快照 | `etl-aggregation-service` | `bi-metric-service` 和 `report-service` 通过数据集 API 或只读读模型访问。 |
| 指标、维度、看板、扩展包 | `bi-config-service` | `bi-metric-service` 读取已发布版本，不读取草稿配置。 |
| BI 热点缓存、查询结果 | `bi-metric-service` | 前端通过分析 API 访问。 |
| 报表任务、报表文件引用 | `report-service` | 前端查询状态和下载文件。 |
| 虚拟账户生命周期、清理审计 | `virtual-account-service` | 各业务服务消费清理事件后自行删除本服务数据，并回传清理结果。 |
| 审计日志 | `audit-service` | 业务服务写审计事件，不直接写审计表。 |

## 7. 服务通信设计

### 7.1 同步通信

适合使用 REST/JSON API 的场景：

- 前端查询列表、详情、状态。
- 创建组织、团队、渠道、看板、报表任务。
- 服务间做少量、粗粒度的引用校验，例如校验团队是否有效。
- BFF 聚合多个只读 API 给前端。

同步 API 设计要求：

- 使用 OpenAPI 管理契约。
- 请求携带 `X-Tenant-Id`、`X-Request-Id`、`Idempotency-Key`。
- 服务不得信任前端传入的租户和组织范围，必须与鉴权上下文校验。
- 跨服务 API 必须粗粒度，避免循环调用和聊天式 RPC。

### 7.2 异步事件

适合使用 PGMQ 的场景：

- 线索导入完成后触发冲突识别。
- 冲突裁定后触发归属更新。
- 广告同步完成后触发 ETL。
- ETL 完成后触发 BI 刷新。
- 用户提交报表后异步生成文件。
- 虚拟账户到期后编排各服务清理数据。

推荐事件清单：

| 事件 | 生产者 | 消费者 |
| --- | --- | --- |
| `org.changed` | `tenant-org-service` | `etl-aggregation-service`、`bi-metric-service` |
| `team.changed` | `team-service` | `etl-aggregation-service`、`bi-metric-service` |
| `ad.account.bound` | `channel-account-service` / `callback-adapter-service` | `ad-sync-service` |
| `ad.sync.requested` | `callback-adapter-service` / `scheduler-service` | `ad-sync-service` |
| `ad.raw.ready` | `ad-sync-service` | `etl-aggregation-service` |
| `lead.ingested` | `lead-ingestion-service` | `lead-conflict-service`、`attribution-service` |
| `lead.conflict.resolved` | `lead-conflict-service` | `attribution-service`、`etl-aggregation-service` |
| `lead.attribution.updated` | `attribution-service` | `etl-aggregation-service` |
| `etl.batch.completed` | `etl-aggregation-service` | `bi-metric-service`、`report-service` |
| `bi.refresh.requested` | `etl-aggregation-service` / `scheduler-service` | `bi-metric-service` |
| `report.generate.requested` | `report-service` API | `report-service` worker |
| `virtual_account.cleanup.requested` | `virtual-account-service` / `scheduler-service` | 所有持有租户业务数据的服务 |
| `virtual_account.cleanup.completed` | 各业务服务 | `virtual-account-service`、`audit-service` |

### 7.3 事件格式

```json
{
  "message_id": "uuid",
  "schema_version": "1.0",
  "event_type": "lead.ingested",
  "tenant_id": "uuid",
  "trace_id": "req_20260524_xxx",
  "occurred_at": "2026-05-24T10:00:00Z",
  "producer": "lead-ingestion-service",
  "payload": {}
}
```

所有消费者必须实现幂等处理。推荐按 `message_id` 和业务幂等键建立 inbox 表或处理记录。

## 8. 关键业务流程

### 8.1 线索上报流程

```text
前端 / API 调用
  -> api-gateway
  -> lead-ingestion-service 创建导入批次或接收 API 线索
  -> lead-ingestion-service 标准化手机号、团队内去重、写入 leads
  -> 发布 lead.ingested
  -> lead-conflict-service 识别跨团队冲突
  -> attribution-service 计算归属候选
  -> etl-aggregation-service 生成事实表和小时聚合
  -> bi-metric-service 刷新热点指标
```

### 8.2 广告数据同步流程

```text
广告 OAuth 回调已完成
  -> callback-adapter-service 接收可信回调结果
  -> channel-account-service 创建或更新广告账户引用
  -> 发布 ad.sync.requested
  -> ad-sync-service 调用平台 API，按平台限流重试
  -> ad-sync-service 原始报表 jsonb 入库
  -> 发布 ad.raw.ready
  -> etl-aggregation-service 标准化为事实表
  -> bi-metric-service 刷新广告消耗、ROI、趋势等指标
```

### 8.3 虚拟演示账户销毁流程

在微服务架构下，不建议 `virtual-account-service` 直接删除其他服务私有表。推荐使用清理编排：

```text
虚拟账户到期或用户手动销毁
  -> virtual-account-service 创建清理任务
  -> 发布 virtual_account.cleanup.requested
  -> 各业务服务删除本服务内该 tenant_id 的业务数据
  -> 各服务删除本服务对象存储文件、Redis key、未处理队列任务
  -> 各服务发布 virtual_account.cleanup.completed
  -> virtual-account-service 汇总清理结果
  -> audit-service 记录不含业务数据的清理审计
  -> virtual-account-service 标记 destroyed
```

## 9. MVP 阶段落地建议

目标微服务清单较完整，但第一阶段不建议一次性物理拆成十几个仓库或十几个独立发布单元。建议采用“逻辑拆分清晰、物理部署适度合并”的 MVP 方案。

### 9.1 MVP 物理服务

| MVP 服务 | 包含逻辑服务 | 说明 |
| --- | --- | --- |
| `gateway-bff` | `api-gateway`、前端聚合 API | 统一入口，承接前端请求。 |
| `control-plane-service` | `tenant-org-service`、`team-service`、`channel-account-service`、`permission-adapter-service`、`bi-config-service` | Go 实现，先合并低计算量控制面能力，内部包边界按目标服务划分。 |
| `ad-integration-service` | `callback-adapter-service`、`ad-sync-service` | Go 接回调，Python worker 做平台同步；可同一发布单元。 |
| `lead-lifecycle-service` | `lead-ingestion-service`、`lead-conflict-service`、`attribution-service` | Python 实现，打通线索、冲突、归属闭环。 |
| `data-insight-service` | `etl-aggregation-service`、`bi-metric-service`、`report-service` | Python 实现，打通 ETL、BI 查询和报表输出。 |
| `virtual-account-service` | 虚拟账户生命周期、迁移、清理编排 | 独立部署，因为数据清理和合规风险高。 |
| `platform-ops-service` | `scheduler-service`、`audit-service` | 统一调度和审计，可后续拆开。 |

### 9.2 MVP 数据隔离

即使物理服务合并，数据库也应按目标边界拆 schema：

| Schema | 目标服务 |
| --- | --- |
| `tenant_org` | `tenant-org-service` |
| `team` | `team-service` |
| `channel` | `channel-account-service` |
| `lead` | `lead-ingestion-service` |
| `conflict` | `lead-conflict-service` |
| `attribution` | `attribution-service` |
| `ad_sync` | `ad-sync-service` |
| `etl` | `etl-aggregation-service` |
| `bi_config` | `bi-config-service` |
| `bi_metric` | `bi-metric-service` |
| `report` | `report-service` |
| `virtual_account` | `virtual-account-service` |
| `audit` | `audit-service` |

这样后续拆分物理服务或独立数据库时，迁移成本可控。

## 10. 演进路线

### 10.1 第一阶段：核心闭环

目标：打通广告数据和招生线索的数据闭环。

优先服务：

- `control-plane-service`
- `ad-integration-service`
- `lead-lifecycle-service`
- `data-insight-service`
- `virtual-account-service`
- `platform-ops-service`

完成能力：

- 租户、组织、团队、渠道基础管理。
- 广告账户回调接入和广告同步。
- Excel/API 线索上报。
- 团队内去重、跨团队冲突识别。
- 基础归属规则。
- ETL 标准事实表。
- 集团总览、团队效率、渠道 ROI。
- 虚拟账户创建和销毁。

### 10.2 第二阶段：按业务压力拆分

触发条件：

- 某个模块发布频率明显高于其他模块。
- 某个模块资源消耗独立，例如 ETL、BI、广告同步。
- 某个模块需要独立扩容或独立故障隔离。
- 团队开始按业务能力分组。

优先拆分：

- 从 `lead-lifecycle-service` 拆出 `lead-conflict-service` 和 `attribution-service`。
- 从 `data-insight-service` 拆出 `etl-aggregation-service`、`bi-metric-service`、`report-service`。
- 从 `control-plane-service` 拆出 `bi-config-service`。

### 10.3 第三阶段：平台化和高吞吐

触发条件：

- 单个平台同步量大、限流规则复杂或 SDK 依赖冲突。
- BI 查询明显高于 OLTP 负载。
- 客户定制 BI 增多。
- 虚拟账户清理需要更强审计和自动验证。

演进方向：

- `ad-sync-service` 按平台拆为 `ad-sync-douyin-service`、`ad-sync-tencent-service`、`ad-sync-baidu-service`、`ad-sync-xiaohongshu-service`。
- BI 查询引入只读副本、物化视图或 OLAP 存储。
- PGMQ 热点队列迁移到专用消息队列。
- `audit-service` 和 `scheduler-service` 独立扩展。

## 11. 服务拆分决策矩阵

| 问题 | 如果答案是“是” | 决策 |
| --- | --- | --- |
| 该模块是否有独立业务语言和规则？ | 例如冲突、归因、广告同步都有独立规则。 | 倾向独立服务。 |
| 该模块是否需要独立扩容？ | ETL、BI、报表、广告同步资源消耗明显。 | 倾向独立服务。 |
| 该模块是否有不同失败模式？ | 广告平台限流失败不应影响组织管理。 | 倾向独立服务。 |
| 该模块是否需要频繁迭代？ | 平台适配、BI 指标、报表模板迭代频繁。 | 倾向独立服务。 |
| 拆分后是否会出现大量同步互调？ | 如果线索和冲突每一步都需要同步互查。 | 暂缓拆分或通过事件读模型解耦。 |
| 是否只是同一业务对象的简单 CRUD？ | 团队和组织早期改动少、规则轻。 | MVP 可合并部署。 |

## 12. 风险与应对

| 风险 | 表现 | 应对 |
| --- | --- | --- |
| 过度拆分 | 服务数量多、联调困难、一次需求改多个服务。 | MVP 物理合并，目标边界先体现在代码包、schema 和契约上。 |
| 共享数据库耦合 | 服务跨表 join、迁移互相影响。 | 每服务 schema、账号、迁移脚本隔离；跨域查询走读模型。 |
| 分布式事务复杂 | 线索、冲突、归属、BI 刷新无法一个事务完成。 | 使用事件驱动、幂等消费、重试和补偿任务。 |
| BI 跨域查询慢 | 多服务 API 聚合成本高。 | ETL 维护标准事实表和 BI 读模型，前端不直接跨服务拼大查询。 |
| 平台 API 不稳定 | 限流、字段变化、同步失败。 | 平台适配器插件化，原始 `jsonb` 保真，错误分类和重试策略版本化。 |
| 虚拟账户清理不完整 | 某个服务遗留业务数据。 | 清理事件广播，各服务自清理，统一汇总验证和审计。 |

## 13. 最终建议

本项目建议采用“目标微服务 + MVP 合并部署”的方案。

目标微服务边界如下：

1. `permission-adapter-service`
2. `scheduler-service`
3. `audit-service`
4. `tenant-org-service`
5. `team-service`
6. `channel-account-service`
7. `callback-adapter-service`
8. `bi-config-service`
9. `lead-ingestion-service`
10. `lead-conflict-service`
11. `attribution-service`
12. `ad-sync-service`
13. `etl-aggregation-service`
14. `bi-metric-service`
15. `report-service`
16. `virtual-account-service`

MVP 阶段建议物理部署为 7 个服务：

1. `gateway-bff`
2. `control-plane-service`
3. `ad-integration-service`
4. `lead-lifecycle-service`
5. `data-insight-service`
6. `virtual-account-service`
7. `platform-ops-service`

这样既能保持微服务边界和未来演进空间，又能控制第一阶段开发、联调、部署和运维成本。
