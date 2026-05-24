# EduAdCRM 开发入口

本仓库从文档仓库开始新建 monorepo。当前实现优先落地配置服务、自有鉴权和广告 OAuth 回调骨架。

## 本地启动

```powershell
docker compose -f infra/docker-compose.yml up -d
go test ./...
go run ./services/control-plane/cmd/control-plane
go run ./services/ad-integration/cmd/ad-integration
npm.cmd --prefix apps/frontend-web run dev
```

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
