# EduAdCRM 开发入口

本仓库从文档仓库开始新建 monorepo。当前实现优先落地配置服务、自有鉴权和广告 OAuth 回调骨架。

完整项目使用手册见 `docs/14-project-user-manual.md`，覆盖立项目标、项目意义、技术细节、部署配置、启动步骤、功能验收、最终效果和快速上手路径。

## 本地启动

```powershell
docker compose -f infra/docker-compose.yml up -d
go test ./...
go run ./services/control-plane/cmd/control-plane
go run ./services/ad-integration/cmd/ad-integration
npm.cmd --prefix apps/frontend-web run dev
```

发布前全量门禁：

```powershell
powershell -ExecutionPolicy Bypass -File scripts/release-gate.ps1
powershell -ExecutionPolicy Bypass -File scripts/release-gate.ps1 -Http
```

`-Http` 模式需要 8080、8081、8090、8091、8092 五个后端服务已启动。详细发布、备份和回滚流程见 `docs/13-release-runbook.md`。

默认 control-plane 使用内存存储。切换到 PostgreSQL：

```powershell
$env:CONTROL_PLANE_STORE="postgres"
$env:DATABASE_URL="postgres://eduadcrm:eduadcrm@127.0.0.1:5432/eduadcrm?sslmode=disable"
go run ./services/control-plane/cmd/control-plane
```

启用 Redis access token 黑名单：

```powershell
$env:REDIS_ADDR="127.0.0.1:6379"
go run ./services/control-plane/cmd/control-plane
```

执行迁移：

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

运行 PostgreSQL 集成测试：

```powershell
$env:CONTROL_PLANE_INTEGRATION_DATABASE_URL=$env:DATABASE_URL
go test ./services/control-plane/internal/store -run TestPostgresStoreIntegration
```

Python 服务统一使用 `uv` 管理。若默认 uv 缓存不可用，可先设置：

```powershell
$env:UV_CACHE_DIR="C:\tmp\uv-cache"
uv run python --version
```

运行 Python 服务示例：

```powershell
$env:UV_CACHE_DIR="C:\tmp\uv-cache"
uv run --package eduadcrm-lead-lifecycle lead-lifecycle
```

## 端口

- control-plane: `http://localhost:8080`
- ad-integration: `http://localhost:8081`
- lead-lifecycle: `http://localhost:8090`
- data-insight: `http://localhost:8091`
- frontend-web: `http://localhost:5173`

## Redis 与 PGMQ

ad-integration 设置 `REDIS_ADDR` 后，OAuth state 会写入 Redis，并在回调时通过 `GETDEL` 原子消费，避免服务重启丢失 state 或并发重放。

设置 `DATABASE_URL` 后，OAuth 回调创建的 `ad.sync.requested` 会写入 `ad_sync.pgmq_messages`：

```powershell
$env:REDIS_ADDR="127.0.0.1:6379"
$env:DATABASE_URL="postgres://eduadcrm:eduadcrm@127.0.0.1:5432/eduadcrm?sslmode=disable"
go run ./services/ad-integration/cmd/ad-integration
```

## 广告平台同步验收

开发环境没有真实平台密钥时，ad-integration 会使用本地 SDK 字段形态客户端生成确定性演示数据。可在前端“广告授权”页面选择抖音、腾讯、百度或小红书并点击“运行同步”，也可以直接调用：

```powershell
curl.exe -X POST http://127.0.0.1:8081/api/ad-sync/run -H "Content-Type: application/json" -d "{\"platform\":\"douyin\",\"date_from\":\"2026-05-20\",\"date_to\":\"2026-05-20\"}"
curl.exe http://127.0.0.1:8081/api/ad-sync/entities?platform=douyin
curl.exe http://127.0.0.1:8081/api/ad-sync/raw-reports?platform=douyin
curl.exe -X POST http://127.0.0.1:8081/api/ad-sync/run -H "Content-Type: application/json" -d "{\"platform\":\"baidu\",\"date_from\":\"2026-05-20\",\"date_to\":\"2026-05-20\"}"
curl.exe -X POST http://127.0.0.1:8081/api/ad-sync/run -H "Content-Type: application/json" -d "{\"platform\":\"xiaohongshu\",\"date_from\":\"2026-05-20\",\"date_to\":\"2026-05-20\"}"
```

百度 P1 默认同步账户日报、计划日报、单元日报，raw 字段保留 `ApiRequestHeader`、`OAuthAuthorizedToolAPI`、`ReportService`、`userName`、`campaignId`、`adgroupId` 等 SDK 形态。小红书 P1 默认同步账户日报、计划日报、单元日报、账户实时报，raw 字段保留 `realtime.AdvertiserRequest`、`realtime.CampaignRequest`、`offline.Request`、`DataReportDTO`、`fee`、`leads` 等字段。设置 `DATABASE_URL` 后，账户引用、同步任务、账户/计划/单元快照和原始日报/小时报会写入 `ad_sync` schema。

data-insight 提供 Python/uv 管理的数据面适配骨架，用于查看百度和小红书字段形态预览：

```powershell
$env:UV_CACHE_DIR=".cache\uv"
$env:PYTHONPATH="services\data-insight\src"
uv run python -c "from data_insight.adapters import baidu, xiaohongshu; print(baidu.request_shape()); print(xiaohongshu.request_shape())"
```

## 线索导入验收

lead-lifecycle 支持创建导入批次、上传 CSV/XLSX、字段映射建议、手机号标准化哈希、团队内去重和错误报告。前端“线索导入”页面可直接使用示例 CSV 体验闭环，也可以直接调用：

```powershell
$env:UV_CACHE_DIR=".cache\uv"
$env:PYTHONPATH="services\lead-lifecycle\src"
uv run python -m unittest services/lead-lifecycle/src/lead_lifecycle/service_test.py
uv run --package eduadcrm-lead-lifecycle lead-lifecycle
curl.exe -X POST http://127.0.0.1:8090/api/leads/imports -H "Content-Type: application/json" -d "{\"tenant_id\":\"demo-tenant\",\"organization_id\":\"demo-org\",\"team_id\":\"demo-team\",\"channel_id\":\"demo-channel\"}"
```

当前开发期使用内存仓储；`003_lead_lifecycle.sql` 已定义 PostgreSQL 表结构，后续阶段接入持久化 repository。

## 冲突与归属验收

不同团队上报相同手机号时，lead-lifecycle 会创建 open 冲突组，并按首次上报生成主咨询建议；管理者可通过前端“线索冲突”页面手动裁定主咨询。

```powershell
curl.exe http://127.0.0.1:8090/api/conflicts
curl.exe -X POST http://127.0.0.1:8090/api/conflicts/{group_id}/resolve -H "Content-Type: application/json" -d "{\"primary_lead_id\":\"lead-id\",\"rule\":\"manual\",\"resolved_by\":\"demo-manager\"}"
curl.exe http://127.0.0.1:8090/api/attributions/rules
```

## ETL 与小时聚合验收

data-insight 支持将广告原始报表和线索数据标准化为 `fact_ad_daily`、`fact_lead_daily` 和 `fact_hourly_aggregate`。前端“ETL 聚合”页面可触发演示 ETL，也可以直接调用：

```powershell
$env:UV_CACHE_DIR=".cache\uv"
$env:PYTHONPATH="services\data-insight\src"
uv run python -m unittest services/data-insight/src/data_insight/etl_test.py
uv run --package eduadcrm-data-insight data-insight
curl.exe -X POST http://127.0.0.1:8091/api/etl/run -H "Content-Type: application/json" -d "{}"
curl.exe http://127.0.0.1:8091/api/etl/facts/hourly
```

## 标准 BI 验收

data-insight 基于 ETL 标准事实表提供集团总览、团队效率、渠道 ROI、招生漏斗、冲突和小时趋势接口。前端“标准 BI”页面可直接查看，也可以直接调用：

```powershell
curl.exe http://127.0.0.1:8091/api/analytics/group-overview
curl.exe http://127.0.0.1:8091/api/analytics/team-efficiency
curl.exe http://127.0.0.1:8091/api/analytics/channel-roi
curl.exe http://127.0.0.1:8091/api/analytics/funnel
curl.exe http://127.0.0.1:8091/api/analytics/hourly-trend
```

## BI 配置与扩展包验收

data-insight 提供指标、维度、数据集、图表目录，以及租户看板布局和扩展包启用、禁用、回滚接口。前端“BI 配置”页面可创建看板、切换看板状态、管理租户扩展包，也可以直接调用：

```powershell
curl.exe http://127.0.0.1:8091/api/bi/catalog/metrics
curl.exe http://127.0.0.1:8091/api/bi/catalog/dimensions
curl.exe http://127.0.0.1:8091/api/bi/catalog/datasets
curl.exe http://127.0.0.1:8091/api/bi/catalog/charts
curl.exe "http://127.0.0.1:8091/api/bi/dashboards?tenant_id=demo-tenant"
curl.exe -X POST http://127.0.0.1:8091/api/bi/dashboards -H "Content-Type: application/json" -d "{\"tenant_id\":\"demo-tenant\",\"dashboard_code\":\"growth_ops_dashboard\",\"name\":\"招生增长运营看板\"}"
curl.exe -X POST http://127.0.0.1:8091/api/bi/extensions/advanced_roi_pack/enable -H "Content-Type: application/json" -d "{\"tenant_id\":\"demo-tenant\"}"
curl.exe -X POST http://127.0.0.1:8091/api/bi/extensions/advanced_roi_pack/disable -H "Content-Type: application/json" -d "{\"tenant_id\":\"demo-tenant\"}"
curl.exe -X POST http://127.0.0.1:8091/api/bi/extensions/advanced_roi_pack/rollback -H "Content-Type: application/json" -d "{\"tenant_id\":\"demo-tenant\"}"
```

## 报表中心验收

data-insight 支持同步生成 Excel 报表任务，导出文件使用下载 token 校验；同一接口会返回行动建议和 BI 口径一致性校验结果。前端“报表中心”页面可创建标准汇总、集团总览、团队效率、渠道 ROI、招生漏斗、冲突治理和小时趋势报表，也可以直接调用：

```powershell
curl.exe http://127.0.0.1:8091/api/reports/advice
curl.exe http://127.0.0.1:8091/api/reports/consistency
curl.exe "http://127.0.0.1:8091/api/reports?tenant_id=demo-tenant"
curl.exe -X POST http://127.0.0.1:8091/api/reports -H "Content-Type: application/json" -d "{\"tenant_id\":\"demo-tenant\",\"report_type\":\"standard_summary\",\"format\":\"xlsx\",\"created_by\":\"demo-manager\"}"
```

创建任务返回的 `download_url` 可直接下载 `.xlsx` 文件；下载失败会返回 `invalid download token`。

## 虚拟账户验收

virtual-account 支持创建 1-7 天虚拟账户，默认 7 天；支持迁移授权、手动销毁、到期清理和清理审计。清理时会分别记录 token、文件、cache、queue 删除数量，并校验无业务数据残留。前端“虚拟账户”页面可直接操作，也可以直接调用：

```powershell
$env:UV_CACHE_DIR=".cache\uv"
$env:PYTHONPATH="services\virtual-account\src"
uv run python -m unittest services/virtual-account/src/virtual_account/service_test.py
uv run --package eduadcrm-virtual-account virtual-account
curl.exe -X POST http://127.0.0.1:8092/api/virtual-accounts -H "Content-Type: application/json" -d "{\"tenant_id\":\"demo-tenant\",\"owner_user_id\":\"demo-user\",\"display_name\":\"7 天虚拟投放账户\",\"purpose\":\"短期投放联调\"}"
curl.exe http://127.0.0.1:8092/api/virtual-accounts
curl.exe -X POST http://127.0.0.1:8092/api/virtual-accounts/{account_id}/authorize-migration -H "Content-Type: application/json" -d "{\"target_user_id\":\"real-user\",\"authorized_by\":\"demo-manager\",\"reason\":\"迁移虚拟账户配置\"}"
curl.exe -X POST http://127.0.0.1:8092/api/virtual-accounts/{account_id}/destroy -H "Content-Type: application/json" -d "{\"actor\":\"demo-manager\",\"reason\":\"manual_destroy\"}"
curl.exe -X POST http://127.0.0.1:8092/api/virtual-accounts/cleanup-expired -H "Content-Type: application/json" -d "{\"actor\":\"scheduler\"}"
curl.exe http://127.0.0.1:8092/api/virtual-accounts/cleanup-logs
```

## 当前已落地接口

- `GET /api/config`
- `PATCH /api/config/{key}`
- `POST /api/config/{key}/rollback`
- `GET /api/config/change-logs`
- `POST /api/auth/register`
- `POST /api/auth/login`
- `POST /api/auth/refresh`
- `POST /api/auth/logout`
- `GET /api/auth/profile`
- `GET /api/orgs/tree`
- `POST /api/orgs`
- `PATCH /api/orgs/{org_id}`
- `POST /api/orgs/{org_id}/move`
- `GET /api/orgs/{org_id}/summary`
- `GET /api/teams`
- `POST /api/teams`
- `PATCH /api/teams/{team_id}`
- `GET /api/channels`
- `POST /api/channels`
- `PATCH /api/channels/{channel_id}`
- `POST /api/oauth/{platform}/authorize`
- `GET /api/oauth/{platform}/callback`
- `GET /api/ad-accounts`
- `POST /api/ad-sync/run`
- `GET /api/ad-sync/jobs`
- `GET /api/ad-sync/entities`
- `GET /api/ad-sync/raw-reports`
- `GET /api/platform-adapters`
- `GET /api/platform-adapters/preview`
- `GET /api/leads/imports`
- `POST /api/leads/imports`
- `POST /api/leads/imports/{batch_id}/upload`
- `POST /api/leads/imports/{batch_id}/confirm`
- `GET /api/leads`
- `POST /api/leads`
- `POST /api/leads/batch`
- `GET /api/conflicts`
- `GET /api/conflicts/{group_id}`
- `POST /api/conflicts/{group_id}/resolve`
- `GET /api/attributions/rules`
- `GET /api/attributions`
- `POST /api/attributions/calculate`
- `POST /api/etl/run`
- `GET /api/etl/batches`
- `GET /api/etl/facts/ad-daily`
- `GET /api/etl/facts/lead-daily`
- `GET /api/etl/facts/hourly`
- `GET /api/analytics/group-overview`
- `GET /api/analytics/team-efficiency`
- `GET /api/analytics/channel-roi`
- `GET /api/analytics/funnel`
- `GET /api/analytics/conflicts`
- `GET /api/analytics/hourly-trend`
- `GET /api/bi/catalog/metrics`
- `GET /api/bi/catalog/dimensions`
- `GET /api/bi/catalog/datasets`
- `GET /api/bi/catalog/charts`
- `GET /api/bi/dashboards`
- `POST /api/bi/dashboards`
- `POST /api/bi/dashboards/{dashboard_id}`
- `GET /api/bi/extensions`
- `POST /api/bi/extensions/{extension_code}/enable`
- `POST /api/bi/extensions/{extension_code}/disable`
- `POST /api/bi/extensions/{extension_code}/rollback`
- `GET /api/reports`
- `POST /api/reports`
- `GET /api/reports/{report_id}`
- `GET /api/reports/{report_id}/download`
- `GET /api/reports/advice`
- `GET /api/reports/consistency`
- `GET /api/virtual-accounts`
- `POST /api/virtual-accounts`
- `GET /api/virtual-accounts/{account_id}`
- `POST /api/virtual-accounts/{account_id}/authorize-migration`
- `POST /api/virtual-accounts/{account_id}/destroy`
- `POST /api/virtual-accounts/cleanup-expired`
- `GET /api/virtual-accounts/cleanup-logs`
