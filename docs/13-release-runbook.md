# EduAdCRM 发布运行手册

## 发布门禁

每次发布前运行：

```powershell
powershell -ExecutionPolicy Bypass -File scripts/release-gate.ps1
```

如果本机或测试环境已启动全部服务，可追加 HTTP 探针：

```powershell
powershell -ExecutionPolicy Bypass -File scripts/release-gate.ps1 -Http
```

门禁覆盖：

- Go 控制面和广告集成服务单测。
- lead-lifecycle、data-insight、virtual-account 的 uv 单测。
- 前端生产构建。
- 可选 HTTP 健康检查、广告同步、线索导入、BI 报表和虚拟账户清理探针。

## 迁移顺序

1. 备份 PostgreSQL。
2. 执行 `migrations/001_core_config_auth.sql` 到 `migrations/008_virtual_accounts.sql`。
3. 启动或滚动更新后端服务。
4. 启动前端新版本。
5. 运行 `scripts/release-gate.ps1 -Http`。
6. 抽查系统配置、广告同步、线索导入、标准 BI、报表中心和虚拟账户页面。

## 备份与恢复

- PostgreSQL：发布前做一次全量备份；生产环境开启 WAL 归档。
- 对象存储：发布前确认报表和导入文件 bucket 生命周期策略。
- Redis：仅保存缓存、黑名单和短期 state，恢复时可从 PostgreSQL 和业务流程重建。
- 恢复演练：至少在预发布环境完成一次数据库恢复和服务重启验证。

## 回滚流程

1. 暂停 worker 和定时清理任务，避免旧版本无法理解新消息。
2. 回滚前端静态资源到上一版本。
3. 回滚后端服务镜像或二进制。
4. 对已执行的数据库迁移做兼容性检查；当前迁移均为新增表或新增能力，默认不需要删除表。
5. 运行 `scripts/release-gate.ps1`，再运行关键 HTTP 探针。
6. 恢复 worker 和定时任务。

## 安全检查

- `auth.access_token_ttl` 默认 24h，`auth.refresh_token_ttl` 默认 720h，发布前确认配置服务当前值。
- OAuth token 只以 `token_ref` 进入业务数据。
- 报表下载必须使用下载 token。
- 虚拟账户销毁必须产生 cleanup log，并满足 `no_business_data=true` 才算清理完成。
- 普通用户不能修改平台级配置、BI 扩展包和虚拟账户销毁结果。

## 性能基准

试点环境的初始门槛：

- 核心 API 健康检查 P95 小于 200ms。
- 标准 BI 首屏接口 P95 小于 1s。
- 常规报表生成在 30s 内完成。
- 5 万行线索导入应异步处理并提供进度。
- 虚拟账户到期清理任务不阻塞在线 API。

## 上线后观察

发布后 30 分钟重点观察：

- 服务错误率和重启次数。
- PostgreSQL 连接数、慢查询和锁等待。
- Redis 内存、命中率和过期 key。
- 广告同步任务失败率。
- ETL、报表和虚拟账户清理日志。
- 多租户查询是否出现异常数据交叉。
