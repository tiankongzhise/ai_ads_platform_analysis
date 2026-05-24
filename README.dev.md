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
- `GET /api/teams`
- `POST /api/teams`
- `GET /api/channels`
- `POST /api/channels`
- `POST /api/oauth/{platform}/authorize`
- `GET /api/oauth/{platform}/callback`
- `GET /api/ad-accounts`
- `GET /api/ad-sync/jobs`
