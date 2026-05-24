# EduAdCRM 项目使用手册

## 1. 手册定位

本文档面向第一次接触 EduAdCRM 的产品、实施、开发、测试和运维人员，提供一份从项目立项背景到本地部署、配置、启动、功能验收、上线检查的完整使用手册。

本文档以当前代码版本和 git 历史为准。早期设计文档中曾出现“鉴权与广告 OAuth 回调由外部系统提供”的假设，当前实现已经修正为本项目自行实现用户鉴权、配置服务、广告 OAuth state 与回调处理。

## 2. 项目立项目标

EduAdCRM 是面向教育行业招生场景的广告 CRM 数据分析平台。它要解决的是教育集团、学校、校区和招生团队在广告投放与线索管理中常见的数据割裂问题。

典型痛点包括：

- 集团管理层无法实时了解各学校招生进度和广告预算使用效果。
- 学校无法及时判断招生团队是否实际投放广告、投放效果是否健康。
- 多个团队可能重复获取同一个学生或家长咨询，线索归属容易产生争议。
- 广告平台只反映曝光、点击和消耗，学校更关心有效线索、到校、报名、成交和 ROI。
- Excel 线索汇总和人工复盘滞后，无法支持小时级调整。
- 销售或实施人员需要现场演示真实广告接入，但客户数据必须可限期清理、可审计。

项目目标是形成从组织、团队、广告渠道、线索、冲突、归属、ETL、BI、报表到虚拟账户的完整闭环，让教育机构能按小时观察招生经营结果，并能通过受控配置扩展 BI 能力。

## 3. 项目意义

对业务侧：

- 让招生进度、广告消耗、团队效率、渠道 ROI 从事后总结变为过程可见。
- 让重复咨询和归属争议进入可审计流程，减少人工扯皮和寻租空间。
- 让投放人员能更早发现高成本渠道和低转化团队，及时调整预算与跟进策略。
- 让销售演示账户具备明确生命周期，避免演示后残留客户业务数据。

对技术侧：

- 使用 Go + Python 的组合，让控制面和数据处理各自使用合适工具。
- 使用 PostgreSQL `jsonb` 保留广告平台原始字段差异，再由 ETL 标准化。
- 使用 Redis 管理短期状态、黑名单和 OAuth state，使用 PGMQ 风格表承接异步事件。
- 使用指标目录、维度目录、数据集目录和扩展包机制，避免用户直接写 SQL。
- 通过发布门禁脚本把单测、构建和 HTTP 探针固化为上线前检查。

## 4. 当前实现总览

当前分支已经实现以下阶段：

| 阶段 | 主要能力 |
| --- | --- |
| 基础设施 | Monorepo、Go/Python/React 工程、Docker Compose、OpenAPI 摘要、迁移脚本。 |
| 配置服务 | `defaults.yaml`、配置发布、配置回滚、变更日志、配置 UI。 |
| 自有鉴权 | 注册、登录、refresh token、登出、profile、access token 24h、refresh token 720h。 |
| 组织控制面 | 组织树、组织移动防循环、团队、渠道、组织汇总。 |
| 广告 OAuth | 四平台授权 URL、一次性 state、回调、防重放、token_ref、同步任务。 |
| 广告同步 | 抖音、腾讯、百度、小红书演示同步，实体快照与原始报表保真。 |
| 线索导入 | CSV/XLSX 解析、字段映射建议、手机号标准化哈希、团队内去重、错误报告。 |
| 冲突归属 | 跨团队同手机号冲突识别、主咨询建议、手动裁定、归属记录。 |
| ETL 聚合 | 广告事实表、线索事实表、小时聚合事实、ETL 批次。 |
| 标准 BI | 集团总览、团队效率、渠道 ROI、招生漏斗、冲突、小时趋势。 |
| BI 配置 | 指标、维度、数据集、图表目录，租户看板和扩展包启停回滚。 |
| 报表中心 | Excel 报表任务、下载 token、行动建议、BI 口径一致性校验。 |
| 虚拟账户 | 1-7 天虚拟账户、迁移授权、手动销毁、到期清理、无业务数据残留审计。 |
| 上线准备 | `scripts/release-gate.ps1` 发布门禁和 `docs/13-release-runbook.md` 运行手册。 |

## 5. 建设历程与 git 历史脉络

项目最初从文档与广告平台 SDK 调研开始，随后在 `codex/dev` 分支上按功能阶段小步落地。git 历史中的主要演进路径如下：

| 历史阶段 | 代表提交 | 说明 |
| --- | --- | --- |
| 文档与调研 | `3082d2ff` 之前 | 建立项目概述、平台调研、OAuth token 生命周期、报表同步模型和微服务拆分方案。 |
| 工程骨架 | `29aa5fc8` | 初始化 monorepo、基础目录、Docker Compose、Go/Python/前端工程。 |
| 配置与鉴权 | `e465ed1a`、`d470681e`、`a9ae34c9` | 实现配置服务、自有用户鉴权、可配置 token 生命周期。 |
| 广告 OAuth | `7e1cf257`、`46f83f27` | 实现四平台 OAuth 授权链接、state 校验、回调和同步任务投递。 |
| 控制面 | `277dedf1`、`23eb2f35`、`75b2c5fd` | 完成组织、团队、渠道、PostgreSQL 仓储、Redis 与 PGMQ 基座。 |
| 广告同步 | `e649946f`、`1c021552` | 完成抖音、腾讯、百度、小红书同步适配和本地演示数据。 |
| 线索与冲突 | `2770d378`、`aeb3a9a3` | 完成线索导入、去重、跨团队冲突和归属裁定。 |
| 数据洞察 | `68506305`、`a031792a`、`76127dbb` | 完成 ETL、标准 BI、BI 配置与扩展包。 |
| 报表与虚拟账户 | `b20a2e3a`、`a18e881a` | 完成报表任务、行动建议、虚拟账户生命周期和清理审计。 |
| 上线准备 | `9ad8939b`、`f12335cd` | 完成发布门禁脚本、HTTP 探针和上线运行手册。 |

每个垂直阶段都追加了 `垂直功能完成：...` 提交，用于记录该阶段接口、数据模型、前端页面和验证结果。因此排查功能来源时，可以优先从这些垂直完成提交向前追溯。

## 6. 技术架构

### 6.1 Monorepo 结构

```text
apps/frontend-web              React + TypeScript + Vite 前端
services/control-plane         Go 控制面：配置、鉴权、组织、团队、渠道
services/ad-integration        Go 广告集成：OAuth、广告账户、同步、平台适配
services/lead-lifecycle        Python 线索导入、冲突、归属
services/data-insight          Python ETL、BI、BI 配置、报表
services/virtual-account       Python 虚拟账户生命周期
services/platform-ops          Python 运维服务骨架
packages/contracts/openapi     API 摘要契约
config/defaults.yaml           默认配置
infra/docker-compose.yml       PostgreSQL、Redis、MinIO
migrations                     PostgreSQL 迁移脚本
scripts/release-gate.ps1       发布门禁脚本
docs                           项目设计、测试、运维和本文档
vendor/ad-platform-sdks        广告平台 SDK 源码快照参考
```

### 6.2 服务与端口

| 服务 | 技术 | 端口 | 作用 |
| --- | --- | --- | --- |
| frontend-web | React/Vite | 5173 | 管理端 UI。 |
| control-plane | Go | 8080 | 配置、鉴权、组织、团队、渠道。 |
| ad-integration | Go | 8081 | 广告 OAuth、账户、同步。 |
| lead-lifecycle | Python/uv | 8090 | 线索导入、冲突、归属。 |
| data-insight | Python/uv | 8091 | ETL、BI、BI 配置、报表。 |
| virtual-account | Python/uv | 8092 | 虚拟账户生命周期与清理审计。 |

### 6.3 数据与中间件

- PostgreSQL：正式持久化设计，迁移脚本位于 `migrations/`。
- Redis：access token 黑名单、OAuth state、短期状态和缓存。
- PGMQ 风格表：当前通过 `ad_sync.pgmq_messages` 承接 `ad.sync.requested`。
- MinIO：对象存储占位，用于导入文件、错误报告、报表和虚拟账户文件。
- 默认开发模式：多数服务使用内存仓储，便于本地快速演示；control-plane 和 ad-integration 已支持 PostgreSQL/Redis 适配。

## 7. 快速上手指南

### 7.1 前置依赖

建议在 Windows PowerShell 中操作。需要本机可用：

- Go
- Node.js 与 npm
- Python 与 uv
- Docker Desktop
- 可选：PostgreSQL 客户端 `psql`

### 7.2 拉起基础组件

```powershell
docker compose -f infra/docker-compose.yml up -d
```

启动后会得到：

- PostgreSQL: `127.0.0.1:5432`
- Redis: `127.0.0.1:6379`
- MinIO API: `127.0.0.1:9000`
- MinIO Console: `127.0.0.1:9001`

### 7.3 启动五个后端服务

分别打开 5 个 PowerShell 终端。

终端 1：

```powershell
go run ./services/control-plane/cmd/control-plane
```

终端 2：

```powershell
go run ./services/ad-integration/cmd/ad-integration
```

终端 3：

```powershell
$env:UV_CACHE_DIR=".cache\uv"
$env:PYTHONPATH="services\lead-lifecycle\src"
uv run python services/lead-lifecycle/src/lead_lifecycle/main.py
```

终端 4：

```powershell
$env:UV_CACHE_DIR=".cache\uv"
$env:PYTHONPATH="services\data-insight\src"
uv run python services/data-insight/src/data_insight/main.py
```

终端 5：

```powershell
$env:UV_CACHE_DIR=".cache\uv"
$env:PYTHONPATH="services\virtual-account\src"
uv run python services/virtual-account/src/virtual_account/main.py
```

### 7.4 启动前端

第 6 个 PowerShell 终端：

```powershell
npm.cmd --prefix apps/frontend-web run dev
```

浏览器打开：

```text
http://127.0.0.1:5173
```

### 7.5 最短体验路径

1. 打开“用户鉴权”，注册一个租户管理员。
2. 打开“系统配置”，确认 `auth.access_token_ttl=24h`、`auth.refresh_token_ttl=720h`。
3. 打开“组织配置”，创建组织、团队和渠道。
4. 打开“广告授权”，选择广告平台并运行同步。
5. 打开“线索导入”，创建批次、上传示例 CSV、确认导入。
6. 打开“线索冲突”，查看同手机号跨团队冲突并裁定主咨询。
7. 打开“ETL 聚合”，运行演示 ETL。
8. 打开“标准 BI”，查看集团总览、团队效率、渠道 ROI 等指标。
9. 打开“BI 配置”，创建看板、启用扩展包。
10. 打开“报表中心”，生成 Excel 报表并查看行动建议和一致性校验。
11. 打开“虚拟账户”，创建虚拟账户、迁移授权、销毁并查看清理审计。

完成以上步骤后，可以看到一个从租户注册、组织配置、广告同步、线索导入、冲突裁定、BI 分析、报表导出到虚拟账户销毁的完整演示闭环。

## 8. 完整部署与配置

### 8.1 执行数据库迁移

如果要使用 PostgreSQL，请先启动 Docker Compose，然后设置连接串：

```powershell
$env:DATABASE_URL="postgres://eduadcrm:eduadcrm@127.0.0.1:5432/eduadcrm?sslmode=disable"
```

依次执行：

```powershell
psql $env:DATABASE_URL -f migrations/001_core_config_auth.sql
psql $env:DATABASE_URL -f migrations/002_ad_oauth.sql
psql $env:DATABASE_URL -f migrations/003_lead_lifecycle.sql
psql $env:DATABASE_URL -f migrations/004_lead_conflict_attribution.sql
psql $env:DATABASE_URL -f migrations/005_etl_facts.sql
psql $env:DATABASE_URL -f migrations/006_bi_config.sql
psql $env:DATABASE_URL -f migrations/007_reports.sql
psql $env:DATABASE_URL -f migrations/008_virtual_accounts.sql
```

各迁移作用：

| 文件 | 作用 |
| --- | --- |
| `001_core_config_auth.sql` | 配置、租户、用户、凭证、refresh token、登录审计。 |
| `002_ad_oauth.sql` | 广告 OAuth token、账户、同步任务、实体、原始报表、PGMQ 消息。 |
| `003_lead_lifecycle.sql` | 线索导入批次、线索、导入错误。 |
| `004_lead_conflict_attribution.sql` | 冲突组、冲突项、归属裁定。 |
| `005_etl_facts.sql` | ETL 批次、广告日事实、线索日事实、小时聚合。 |
| `006_bi_config.sql` | BI 数据集、指标、看板、租户扩展包。 |
| `007_reports.sql` | 报表任务、下载日志、行动建议。 |
| `008_virtual_accounts.sql` | 虚拟账户、迁移授权、清理日志。 |

### 8.2 配置文件

默认配置在 `config/defaults.yaml`：

```yaml
auth:
  access_token_ttl: 24h
  refresh_token_ttl: 720h
  refresh_token_rotation: true
  logout_blacklist_enabled: true
  issuer: eduadcrm
  jwt_secret: dev-only-change-me
  password_iterations: 210000
config:
  cache_refresh_interval: 30s
oauth:
  state_ttl: 10m
```

本地覆盖文件可使用 `config/local.yaml`，该文件用于本机私有配置，不应提交真实密钥。

control-plane 支持环境变量：

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `EDUADCRM_CONFIG_FILE` | `config/defaults.yaml` | 默认配置文件路径。 |
| `EDUADCRM_LOCAL_CONFIG_FILE` | `config/local.yaml` | 本地覆盖配置路径。 |
| `CONTROL_PLANE_ADDR` | `:8080` | control-plane 监听地址。 |
| `CONTROL_PLANE_STORE` | `memory` | 设置为 `postgres` 后使用 PostgreSQL。 |
| `DATABASE_URL` | 空 | PostgreSQL 连接串。 |
| `REDIS_ADDR` | 空 | Redis 地址，启用 access token 黑名单。 |

ad-integration 支持：

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `AD_INTEGRATION_ADDR` | `:8081` | ad-integration 监听地址。 |
| `DATABASE_URL` | 空 | 启用 PostgreSQL 仓储和 PGMQ 消息写入。 |
| `REDIS_ADDR` | 空 | 启用 Redis OAuth state 存储。 |

### 8.3 使用 PostgreSQL/Redis 启动 Go 服务

control-plane：

```powershell
$env:CONTROL_PLANE_STORE="postgres"
$env:DATABASE_URL="postgres://eduadcrm:eduadcrm@127.0.0.1:5432/eduadcrm?sslmode=disable"
$env:REDIS_ADDR="127.0.0.1:6379"
go run ./services/control-plane/cmd/control-plane
```

ad-integration：

```powershell
$env:DATABASE_URL="postgres://eduadcrm:eduadcrm@127.0.0.1:5432/eduadcrm?sslmode=disable"
$env:REDIS_ADDR="127.0.0.1:6379"
go run ./services/ad-integration/cmd/ad-integration
```

启用后：

- 登出时 access token hash 会写入 Redis 黑名单。
- OAuth state 会写入 Redis，并在回调时一次性消费。
- OAuth 回调成功后 `ad.sync.requested` 会写入 `ad_sync.pgmq_messages`。
- 广告账户、实体和原始报表会写入 `ad_sync` schema。

## 9. 页面功能与逐步验收

### 9.1 系统配置

入口：前端左侧“系统配置”。

能实现：

- 查看配置定义和当前值。
- 发布白名单配置。
- 回滚配置。
- 查看配置变更日志。

重点配置：

- `auth.access_token_ttl` 默认 `24h`
- `auth.refresh_token_ttl` 默认 `720h`
- `auth.refresh_token_rotation` 默认 `true`
- `auth.logout_blacklist_enabled` 默认 `true`

验收方式：

1. 打开“系统配置”。
2. 修改 `auth.access_token_ttl` 为 `12h`。
3. 页面提示发布成功，配置表显示新版本。
4. 点击回滚，配置恢复到上一值或默认值。
5. “配置变更日志”出现 publish 和 rollback 记录。

接口验收：

```powershell
curl.exe http://127.0.0.1:8080/api/config
```

### 9.2 用户鉴权

入口：前端左侧“用户鉴权”。

能实现：

- 注册租户管理员。
- 邮箱密码登录。
- 签发 access token 与 refresh token。
- access token 默认 24 小时过期，refresh token 默认 30 天过期。
- refresh token 轮换。
- 登出时撤销 refresh token，并按配置写入 access token 黑名单。

验收方式：

1. 在“注册租户”填写机构名称、邮箱、姓名、密码。
2. 注册成功后页面显示 access token 和 refresh token 过期时间。
3. 切换到“登录”页，用同一邮箱密码登录。
4. 确认 token 过期时间符合当前配置。

接口验收：

```powershell
curl.exe -X POST http://127.0.0.1:8080/api/auth/register -H "Content-Type: application/json" -d "{\"tenant_name\":\"演示教育集团\",\"email\":\"admin@example.com\",\"display_name\":\"管理员\",\"password\":\"password123\"}"
curl.exe -X POST http://127.0.0.1:8080/api/auth/login -H "Content-Type: application/json" -d "{\"email\":\"admin@example.com\",\"password\":\"password123\"}"
```

### 9.3 组织、团队与渠道

入口：前端左侧“组织配置”。

能实现：

- 创建组织节点。
- 查询组织树。
- 创建团队。
- 创建广告渠道。
- 组织移动防循环。
- 组织子树汇总。

验收方式：

1. 先在“用户鉴权”注册或登录，前端会保存 access token。
2. 在“组织配置”创建一个学校组织。
3. 使用该组织 ID 创建招生团队。
4. 使用该组织 ID 创建广告渠道。
5. 页面中的组织树、团队表、渠道表都出现对应数据。

注意：组织相关接口需要鉴权。用 curl 调用时，需要携带 `Authorization: Bearer <access_token>`。

### 9.4 广告授权与广告同步

入口：前端左侧“广告授权”。

能实现：

- 为抖音、腾讯、百度、小红书生成授权链接和一次性 state。
- 模拟广告平台回调，创建广告账户引用和 token_ref。
- 手动运行广告同步。
- 查看同步任务、广告实体、原始报表。

技术细节：

- OAuth state 默认 10 分钟有效。
- 回调时 state 会被一次性消费，避免重放。
- token 明文不进入业务对象，业务侧只使用 `token_ref`。
- 没有真实平台密钥时，同步服务生成与 vendor SDK 字段形态一致的确定性演示数据。

验收方式：

1. 在“广告授权”选择平台。
2. 点击生成授权链接。
3. 点击运行同步。
4. 页面显示同步完成、实体数量和报表行数。
5. 实体表能看到账户、计划、单元等快照。
6. 原始报表表能看到平台 raw 字段。

接口验收：

```powershell
curl.exe -X POST http://127.0.0.1:8081/api/ad-sync/run -H "Content-Type: application/json" -d "{\"platform\":\"douyin\",\"date_from\":\"2026-05-20\",\"date_to\":\"2026-05-20\"}"
curl.exe http://127.0.0.1:8081/api/ad-sync/entities?platform=douyin
curl.exe http://127.0.0.1:8081/api/ad-sync/raw-reports?platform=douyin
```

预期效果：

- 抖音/腾讯会返回账户、计划、单元和日报/小时报演示数据。
- 百度会返回账户日报、计划日报、单元日报演示数据。
- 小红书会返回账户日报、计划日报、单元日报、账户实时报演示数据。

### 9.5 线索导入

入口：前端左侧“线索导入”。

能实现：

- 创建导入批次。
- 上传 CSV 或 XLSX。
- 自动识别字段映射。
- 手机号标准化与 hash。
- 团队内相同手机号去重。
- 生成系统线索号。
- 输出导入错误报告。

支持字段别名：

| 目标字段 | 可识别表头示例 |
| --- | --- |
| `student_name` | `学生姓名`、`姓名`、`客户姓名`、`student_name` |
| `phone` | `手机号`、`手机`、`联系电话`、`phone` |
| `source_channel` | `渠道`、`来源`、`线索来源`、`source_channel` |
| `stage` | `线索阶段`、`跟进阶段`、`stage` |
| `reported_at` | `上报时间`、`创建时间`、`咨询时间` |

前端默认 CSV：

```csv
学生姓名,手机号,渠道,线索阶段
张三,13800138000,抖音,new
李四,13800138000,腾讯,new
王五,13900139000,百度,visited
```

验收方式：

1. 创建导入批次。
2. 上传默认 CSV。
3. 查看预览行和字段映射建议。
4. 点击确认导入。
5. 线索列表出现成功导入的线索。
6. 同一团队重复手机号会进入错误报告。

接口验收：

```powershell
curl.exe -X POST http://127.0.0.1:8090/api/leads/imports -H "Content-Type: application/json" -d "{\"tenant_id\":\"demo-tenant\",\"organization_id\":\"demo-org\",\"team_id\":\"team-a\",\"channel_id\":\"douyin\"}"
```

### 9.6 线索冲突与归属

入口：前端左侧“线索冲突”。

能实现：

- 不同团队上报同一手机号时创建冲突组。
- 按首次上报生成主咨询建议。
- 管理者可手动裁定主咨询。
- 非主咨询被标记为冲突量。
- 生成归属记录。

验收方式：

1. 用不同团队导入相同手机号线索。
2. 打开“线索冲突”。
3. 查看 open 冲突组。
4. 选择某条线索为主咨询并裁定。
5. 冲突组状态变为 resolved。
6. 冲突项中主咨询和冲突角色清晰展示。

接口验收：

```powershell
curl.exe http://127.0.0.1:8090/api/conflicts
curl.exe http://127.0.0.1:8090/api/attributions/rules
```

### 9.7 ETL 与小时聚合

入口：前端左侧“ETL 聚合”。

能实现：

- 运行演示 ETL。
- 生成 `fact_ad_daily`。
- 生成 `fact_lead_daily`。
- 生成 `fact_hourly_aggregate`。
- 查看 ETL 批次结果。

验收方式：

1. 打开“ETL 聚合”。
2. 点击运行演示 ETL。
3. ETL 批次状态为 success。
4. 广告日事实、线索日事实、小时聚合表都有数据。

接口验收：

```powershell
curl.exe -X POST http://127.0.0.1:8091/api/etl/run -H "Content-Type: application/json" -d "{}"
curl.exe http://127.0.0.1:8091/api/etl/facts/ad-daily
curl.exe http://127.0.0.1:8091/api/etl/facts/lead-daily
curl.exe http://127.0.0.1:8091/api/etl/facts/hourly
```

### 9.8 标准 BI

入口：前端左侧“标准 BI”。

能实现：

- 集团总览：消耗、线索、报名、线索成本、转化率。
- 团队效率：线索、有效线索、到校、报名、冲突、有效率。
- 渠道 ROI：渠道消耗、线索数、线索成本。
- 招生漏斗：线索、有效、到校、报名。
- 冲突统计：冲突数和冲突率。
- 小时趋势：广告与线索小时聚合。

验收方式：

1. 先运行 ETL。
2. 打开“标准 BI”。
3. 确认集团总览和各表格出现指标。
4. 渠道 ROI、漏斗和小时趋势能展示数据。

接口验收：

```powershell
curl.exe http://127.0.0.1:8091/api/analytics/group-overview
curl.exe http://127.0.0.1:8091/api/analytics/team-efficiency
curl.exe http://127.0.0.1:8091/api/analytics/channel-roi
curl.exe http://127.0.0.1:8091/api/analytics/funnel
curl.exe http://127.0.0.1:8091/api/analytics/hourly-trend
```

### 9.9 BI 配置与扩展包

入口：前端左侧“BI 配置”。

能实现：

- 查看指标目录。
- 查看维度目录。
- 查看数据集目录。
- 查看图表目录。
- 创建租户看板布局。
- 切换看板状态。
- 启用、禁用、回滚租户 BI 扩展包。

验收方式：

1. 打开“BI 配置”。
2. 页面展示指标、维度、数据集和图表数量。
3. 创建一个看板。
4. 看板列表出现新看板。
5. 启用“高级 ROI 扩展”。
6. 扩展包列表出现 enabled 状态。
7. 点击禁用或回滚，状态变为 disabled 或 rollback_pending。

接口验收：

```powershell
curl.exe http://127.0.0.1:8091/api/bi/catalog/metrics
curl.exe -X POST http://127.0.0.1:8091/api/bi/dashboards -H "Content-Type: application/json" -d "{\"tenant_id\":\"demo-tenant\",\"dashboard_code\":\"growth_ops_dashboard\",\"name\":\"招生增长运营看板\"}"
curl.exe -X POST http://127.0.0.1:8091/api/bi/extensions/advanced_roi_pack/enable -H "Content-Type: application/json" -d "{\"tenant_id\":\"demo-tenant\"}"
```

### 9.10 报表中心

入口：前端左侧“报表中心”。

能实现：

- 生成标准汇总、集团总览、团队效率、渠道 ROI、招生漏斗、冲突治理、小时趋势报表。
- 同步生成 Excel-compatible `.xlsx` 文件。
- 使用下载 token 下载报表。
- 生成行动建议。
- 校验报表指标与 BI 指标同口径一致。

验收方式：

1. 打开“报表中心”。
2. 选择“标准汇总”。
3. 点击生成。
4. 任务列表出现 success 状态和下载按钮。
5. 行动建议表至少出现一条建议。
6. BI 口径一致性显示 passed。
7. 点击下载，浏览器下载 `.xlsx` 文件。

接口验收：

```powershell
curl.exe http://127.0.0.1:8091/api/reports/advice
curl.exe http://127.0.0.1:8091/api/reports/consistency
curl.exe -X POST http://127.0.0.1:8091/api/reports -H "Content-Type: application/json" -d "{\"tenant_id\":\"demo-tenant\",\"report_type\":\"standard_summary\",\"format\":\"xlsx\",\"created_by\":\"demo-manager\"}"
```

预期效果：

- 报表任务 `status=success`。
- `consistency.status=passed`。
- `download_url` 能下载 Excel 文件。

### 9.11 虚拟账户

入口：前端左侧“虚拟账户”。

能实现：

- 创建 1-7 天虚拟账户，默认 7 天。
- 迁移授权到正式用户。
- 手动销毁虚拟账户。
- 清理到期账户。
- 查看清理审计日志。
- 清理时记录 token、文件、cache、queue 删除数量。
- 若存在业务数据残留，清理进入 blocked，不删除资源。

验收方式：

1. 打开“虚拟账户”。
2. 创建 7 天虚拟账户。
3. 账户列表显示 active 和到期时间。
4. 点击迁移授权，状态变为 migration_authorized。
5. 点击销毁，状态变为 destroyed，`cleanup_verified=true`。
6. 清理审计中出现 token、文件、cache、queue 删除数量。
7. `no_business_data=true`。

接口验收：

```powershell
curl.exe -X POST http://127.0.0.1:8092/api/virtual-accounts -H "Content-Type: application/json" -d "{\"tenant_id\":\"demo-tenant\",\"owner_user_id\":\"demo-user\",\"display_name\":\"7 天虚拟投放账户\",\"purpose\":\"短期投放联调\"}"
curl.exe http://127.0.0.1:8092/api/virtual-accounts
curl.exe http://127.0.0.1:8092/api/virtual-accounts/cleanup-logs
```

## 10. 发布前验收

基础门禁：

```powershell
powershell -ExecutionPolicy Bypass -File scripts/release-gate.ps1
```

会执行：

- Go 控制面和广告集成服务单测。
- lead-lifecycle 单测。
- data-insight 单测。
- virtual-account 单测。
- 前端生产构建。

HTTP 门禁：

```powershell
powershell -ExecutionPolicy Bypass -File scripts/release-gate.ps1 -Http
```

需要先启动 8080、8081、8090、8091、8092 五个后端服务。会额外检查：

- 五个服务健康检查。
- 广告同步探针。
- 线索导入探针。
- BI 与报表探针。
- 虚拟账户清理探针。

通过后输出：

```text
Release gate passed.
```

## 11. 最终呈现效果

当所有服务启动并完成快速体验路径后，系统最终呈现的是一个教育招生广告经营工作台：

- 左侧菜单覆盖配置、鉴权、组织、广告、线索、冲突、ETL、BI、报表和虚拟账户。
- 管理员可以注册租户并配置 token 生命周期等基础参数。
- 学校或集团可以维护组织树、团队和广告渠道。
- 广告平台可被授权并同步演示数据，原始报表保留平台字段。
- 招生线索可通过 CSV/XLSX 导入，并自动去重、生成线索号。
- 跨团队重复手机号会进入冲突裁定。
- ETL 会将广告与线索数据转换为事实表。
- 标准 BI 展示招生进度、团队效率、渠道 ROI、漏斗和小时趋势。
- BI 配置页面展示受控指标、维度、数据集、图表和扩展包。
- 报表中心生成 Excel 报表、行动建议和一致性校验。
- 虚拟账户页面完成演示账户创建、迁移授权、销毁和清理审计。

最终验收标准：

- `scripts/release-gate.ps1` 通过。
- `scripts/release-gate.ps1 -Http` 通过。
- 前端所有菜单均可打开。
- 核心页面至少能完成一次创建或运行操作。
- BI 与报表一致性为 passed。
- 虚拟账户销毁后 `cleanup_verified=true`，清理日志 `no_business_data=true`。

## 12. 常见问题

### 12.1 前端请求 404 或连接失败

确认对应后端服务已经启动：

```powershell
curl.exe http://127.0.0.1:8080/healthz
curl.exe http://127.0.0.1:8081/healthz
curl.exe http://127.0.0.1:8090/healthz
curl.exe http://127.0.0.1:8091/healthz
curl.exe http://127.0.0.1:8092/healthz
```

### 12.2 组织配置接口提示未登录

先在“用户鉴权”注册或登录。前端会将 access token 保存到 localStorage，并自动加到请求头。

### 12.3 Python 服务无法 import 包

使用源码入口启动时请设置 `PYTHONPATH`：

```powershell
$env:PYTHONPATH="services\data-insight\src"
uv run python services/data-insight/src/data_insight/main.py
```

lead-lifecycle 和 virtual-account 同理。

### 12.4 Vite build 提示 chunk 较大

这是当前单文件前端演示实现带来的构建提醒，不影响构建成功。后续生产化可按路由拆分和动态导入优化。

### 12.5 PostgreSQL 迁移失败

确认 Docker Compose 已启动，连接串正确：

```powershell
docker compose -f infra/docker-compose.yml ps
$env:DATABASE_URL="postgres://eduadcrm:eduadcrm@127.0.0.1:5432/eduadcrm?sslmode=disable"
```

### 12.6 真实广告平台如何接入

当前同步逻辑在无真实密钥时使用本地演示数据，并保留 vendor SDK 字段形态。接真实平台时，需要将 ad-integration 的平台 client 替换为真实 SDK/REST 调用，并配置平台 app id、secret、回调地址和 token 加密密钥。

## 13. 相关文档

- `README.dev.md`：开发入口和命令摘要。
- `docs/01-project-overview.md`：项目概述。
- `docs/03-requirements-specification.md`：需求规格。
- `docs/04-system-architecture.md`：系统架构。
- `docs/05-database-design.md`：数据库设计。
- `docs/06-api-and-messaging-design.md`：API 与消息设计。
- `docs/09-test-plan.md`：测试计划。
- `docs/10-deployment-and-operations.md`：部署与运维设计。
- `docs/13-release-runbook.md`：发布运行手册。
- `docs/ad-platform-research/`：广告平台 OAuth 和报表同步调研。

## 14. 当前版本说明

当前版本是可运行演示与开发基座，不是最终生产部署形态。重要边界：

- control-plane 和 ad-integration 已具备 PostgreSQL/Redis 适配。
- lead-lifecycle、data-insight、virtual-account 当前主要使用内存服务实现闭环，迁移脚本已定义目标表结构。
- 广告同步默认使用演示数据，不直接访问真实平台 API。
- 前端是单页管理台，当前强调功能闭环和验收路径。
- 发布前应至少运行基础门禁；联调环境应运行 HTTP 门禁。
