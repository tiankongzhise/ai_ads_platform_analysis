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
- frontend-web: `http://localhost:5173`

## Redis 与 PGMQ

ad-integration 设置 `REDIS_ADDR` 后，OAuth state 会写入 Redis，并在回调时通过 `GETDEL` 原子消费，避免服务重启丢失 state 或并发重放。

设置 `DATABASE_URL` 后，OAuth 回调创建的 `ad.sync.requested` 会写入 `ad_sync.pgmq_messages`：

```powershell
$env:REDIS_ADDR="127.0.0.1:6379"
$env:DATABASE_URL="postgres://eduadcrm:eduadcrm@127.0.0.1:5432/eduadcrm?sslmode=disable"
go run ./services/ad-integration/cmd/ad-integration
```

## 抖音与腾讯同步验收

开发环境没有真实平台密钥时，ad-integration 会使用本地 SDK 字段形态客户端生成确定性演示数据。可在前端“广告授权”页面选择抖音或腾讯并点击“运行同步”，也可以直接调用：

```powershell
curl.exe -X POST http://127.0.0.1:8081/api/ad-sync/run -H "Content-Type: application/json" -d "{\"platform\":\"douyin\",\"date_from\":\"2026-05-20\",\"date_to\":\"2026-05-20\"}"
curl.exe http://127.0.0.1:8081/api/ad-sync/entities?platform=douyin
curl.exe http://127.0.0.1:8081/api/ad-sync/raw-reports?platform=douyin
```

设置 `DATABASE_URL` 后，账户引用、同步任务、账户/计划/单元快照和原始日报/小时报会写入 `ad_sync` schema。

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
