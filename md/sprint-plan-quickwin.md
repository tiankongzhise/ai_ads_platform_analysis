# EduAdCRM 短期敏捷开发计划（Quick-Win Sprint）

> **计划版本**：v1.0  
> **制定日期**：2026-04-28  
> **计划周期**：约 5 个工作日（1 周冲刺）  
> **目标**：前后端联调全部打通，核心三大功能可演示，Mock 层保底兜底  
> **基于**：CODE_REVIEW.md 审查意见 + 现有代码盘点 + 产品需求优先级

---

## 一、现状诊断（快照）

### 1.1 开发完成度评估

| 模块 | 后端 | 前端 | 联调状态 |
|------|------|------|----------|
| 广告账户 OAuth（巨量/百度）| ✅ 已实现 | ✅ 已实现 | ⚠️ P1-1 OAuth 跳转方式错误 |
| 广告账户 CRUD + 同步 | ✅ 已实现 | ✅ 已实现 | ⚠️ P1-2 balance 类型不一致 |
| 用户认证（登录/注册/登出）| ❌ 路由未创建 | ✅ 页面已有 | ❌ 404 |
| 演示数据系统 | ❌ 路由未创建 | ✅ Store/组件已有 | ❌ 404 |
| 引导向导（Onboarding）| ❌ 路由未创建 | ✅ 组件已有 | ❌ 404 |
| CRM 导入 | ❌ 路由未创建 | ✅ 页面已有 | ❌ 404 |
| ROI 仪表盘分析 | ❌ 路由未创建 | ✅ 页面已有 | ❌ 404 |
| 报表中心 | ❌ 路由未创建 | ✅ 页面已有 | ❌ 404 |
| 系统设置 | ❌ 路由未创建 | ✅ 页面已有 | ❌ 404 |

**核心问题**：后端仅有 `ad_accounts.py` 一个路由，另外 7 个模块路由文件全部缺失，导致前端 90% 接口返回 404。

### 1.2 必须修复的 Code Review P0/P1 问题

| 编号 | 问题 | 修复方案 |
|------|------|----------|
| P0-1/P0-2 | 7 个路由模块缺失 + 未注册 | 本计划核心任务 |
| P0-3 | `TenantContext` 类变量并发安全漏洞 | 改用 `ContextVar` |
| P0-4 | 缺少 ORM 模型（crm_lead, onboarding 等）| Mock 层先用 SQLite 补全 |
| P0-5 | 缺少 Pydantic Schema 文件 | 随路由一起创建 |
| P1-1 | OAuth 用 `window.open` 而非 `window.location.href` | 前端一行修复 |
| P1-6 | OAuth 回调后未读取 URL 参数 | 前端 useEffect 补全 |

---

## 二、三大目标（本次冲刺）

```
目标 1：多租户注册 / 登录 / 注销  —— 前后端联调通过
目标 2：百度账户绑定 + OAuth 授权 + 拉取默认最近7天数据  —— 前后端联调通过
目标 3：SQLite Mock 层  —— 所有未开发接口从 Mock 返回假数据，前端不再出现 404
```

---

## 三、技术设计决策

### 3.1 SQLite Mock 层设计（ADR-001）

**状态**：已接受  
**背景**：MySQL 主数据库需要完整业务逻辑支撑，但当前大量路由未实现，联调必须先通过。需要一个低成本的 Mock 数据层，不影响已实现的真实逻辑，后期无缝切换真实实现。

**决策**：
- 在 `backend/app/` 下新建 `mock/` 目录，包含 `mock_db.py`（SQLite 连接）和 `seed.py`（Mock 数据生成）
- Mock 数据库文件路径：`backend/mock_data.sqlite3`（`.gitignore` 中排除）
- 每个 Mock 路由通过环境变量 `MOCK_MODE=true` 或依赖注入标志来决定走 Mock 还是真实逻辑
- **已实现的路由**（广告账户模块）**不受影响**，继续使用 MySQL

**权衡**：
- ✅ 快速打通联调，前端不再 404
- ✅ Mock 代码与真实代码同目录，替换成本低
- ⚠️ Mock 数据质量影响联调体验，需要生成符合业务场景的数据

### 3.2 多租户认证设计（ADR-002）

**状态**：已接受  
**背景**：现有 `User`、`Tenant` ORM 模型已存在，`security.py` JWT 逻辑已存在，缺的是 `auth.py` 路由。

**决策**：
- 注册时同时创建 `Tenant` 记录（一对一，租户即注册机构）
- JWT payload 携带 `tenant_id` + `user_id`
- 中间件从 JWT 中提取 `tenant_id` 写入 `ContextVar`（修复 P0-3 并发漏洞）
- Token 规格：access_token 2h，refresh_token 7d，登出写入 Redis 黑名单

### 3.3 百度 OAuth 设计（ADR-003）

**状态**：已接受  
**背景**：`baidu_service.py` 已实现 API 封装，`ad_accounts.py` 已有巨量引擎 OAuth 流程可参考。

**决策**：
- 百度 OAuth 回调统一走 `window.location.href` 整页跳转（修复 P1-1）
- 绑定成功后立即触发 Celery 任务拉取最近 7 天数据（同巨量引擎逻辑）
- 回调参数读取（`oauth_result`）在 `AdAccountsPage` `useEffect` 中统一处理（修复 P1-6）

---

## 四、Sprint 任务拆分（5 天）

### Day 1（今日）：基础设施修复

#### 后端任务

**T-B01：修复 `TenantContext` 并发安全漏洞（P0-3）**
- 文件：`backend/app/core/tenant.py`
- 将类变量改为 `ContextVar`，参考 CODE_REVIEW.md P0-3 提供的修复方案
- 预计：30 分钟

**T-B02：创建 SQLite Mock 层基础设施**
- 新建 `backend/app/mock/__init__.py`
- 新建 `backend/app/mock/mock_db.py`（SQLite aiosqlite 异步连接）
- 新建 `backend/app/mock/seed.py`（生成全套 Mock 数据）
- Mock 数据覆盖：演示机构"星海教育"完整数据集（30 天广告数据 + 500 条 CRM 线索 + 3 份报表）
- 预计：2 小时

**T-B03：创建 `auth.py` 路由（真实逻辑）**
- 文件：`backend/app/api/v1/auth.py`
- 接口：`POST /auth/register`、`POST /auth/login`、`POST /auth/logout`、`POST /auth/refresh`、`GET /auth/profile`
- 依赖已有：`User` ORM、`Tenant` ORM、`security.py`（JWT）、`redis_client.py`（Token 黑名单）
- 新建 `backend/app/schemas/auth.py`（LoginRequest、LoginResponse、RegisterRequest、UserResponse）
- 预计：3 小时

**T-B04：注册所有路由到 `main.py`**
- 按需创建占位路由文件（后续 Day 2-3 填充 Mock 实现）
- 确保服务启动后所有接口都有响应（哪怕是 Mock 响应）
- 预计：30 分钟

#### 前端任务

**T-F01：修复 OAuth `window.open` → `window.location.href`（P1-1）**
- 文件：`frontend-web/src/pages/ad-accounts/index.tsx`
- 预计：10 分钟

**T-F02：补全 OAuth 回调结果处理（P1-6）**
- 文件：`frontend-web/src/pages/ad-accounts/index.tsx`
- 在 `useEffect` 中读取 `oauth_result` 参数并 Toast 提示，清除 URL 参数
- 预计：20 分钟

---

### Day 2：认证联调 + 百度 OAuth 路由

#### 后端任务

**T-B05：完善 `auth.py` 并联调**
- 补全注册时自动创建 Tenant 逻辑
- 补全登出 Redis Token 黑名单逻辑
- 补全 Token 刷新端点
- 预计：2 小时

**T-B06：百度 OAuth 路由补全（真实逻辑）**
- 文件：`backend/app/api/v1/ad_accounts.py`（百度回调已有框架，需要补全）
- 确认 `baidu_service.py` 的 Token 换取逻辑完整
- 绑定后触发 Celery `sync_baidu_ad_data` 任务（7 天）
- 预计：2 小时

**T-B07：`onboarding.py` 路由（SQLite Mock）**
- 文件：`backend/app/api/v1/onboarding.py`
- 接口：`GET /onboarding/status`、`POST /onboarding/status`、`POST /onboarding/complete`
- 使用 SQLite Mock 存储引导进度
- 新建 `backend/app/schemas/onboarding.py`
- 预计：1.5 小时

#### 前端任务

**T-F03：登录/注册页面联调**
- 文件：`frontend-web/src/pages/login/index.tsx`
- 确认注册、登录、登出流程与后端 API 完整对接
- 检查 `authStore.ts` 的 `login/register/logout` action 与 API 签名一致
- 预计：1.5 小时

**T-F04：修复 `timeRange` parseInt 问题（P1-3）**
- 文件：`frontend-web/src/pages/dashboard/index.tsx`
- `options` 的 value 直接使用数字
- 预计：10 分钟

---

### Day 3：Mock 层填充 + 仪表盘/演示数据联调

#### 后端任务

**T-B08：`demo.py` 路由（SQLite Mock）**
- 文件：`backend/app/api/v1/demo.py`
- 接口：`GET /demo/dashboard`、`GET /demo/metrics`
- 从 SQLite Mock 中读取"星海教育"演示数据集返回
- 新建 `backend/app/schemas/demo.py`（DemoDashboardResponse）
- 预计：1.5 小时

**T-B09：`analytics.py` 路由（SQLite Mock）**
- 文件：`backend/app/api/v1/analytics.py`
- 接口：`GET /analytics/dashboard`、`GET /analytics/dashboard/status`、`GET /analytics/trend`、`GET /analytics/channel-compare`、`GET /analytics/cross-table`、`GET /analytics/cross-table/export`
- 从 SQLite Mock 读取数据，`dashboard/status` 根据当前租户真实数据状态返回（有无广告账户、有无 CRM）
- 新建 `backend/app/schemas/analytics.py`
- 预计：3 小时

**T-B10：`settings.py` 路由（SQLite Mock）**
- 文件：`backend/app/api/v1/settings.py`
- 接口：租户信息 CRUD、归因规则 CRUD
- 预计：1.5 小时

#### 前端任务

**T-F05：演示数据 + 仪表盘联调**
- 确认 `demoStore.ts` 正确调用 `/demo/dashboard`
- 确认 `DashboardPage` 正确调用 `/analytics/dashboard` 和 `/analytics/dashboard/status`
- 修复 P2-2（EmptyState scenario 类型对齐）
- 修复 P2-7（`enableDemoMode` 解构方式）
- 预计：2 小时

---

### Day 4：CRM 路由 + 报表路由 Mock

#### 后端任务

**T-B11：`crm.py` 路由（真实 + Mock 混合）**
- 文件：`backend/app/api/v1/crm.py`
- 文件上传 `POST /crm/upload` → 使用真实 Celery 任务（`import_tasks.py` 已有框架）
- 进度查询 `GET /crm/upload/{task_id}` → Redis 真实状态
- 字段预览 `POST /crm/upload/{batch_id}/preview` → Mock 返回字段分析结果
- 字段确认 `POST /crm/upload/{batch_id}/confirm` → 写入 SQLite Mock
- 归因建议 `GET /crm/upload/{batch_id}/attribution-suggest` → Mock 返回
- 线索列表/更新/删除 → SQLite Mock
- 新建 `backend/app/schemas/crm.py`、`backend/app/models/crm_lead.py`（基础 ORM）
- 预计：4 小时

**T-B12：`reports.py` 路由（SQLite Mock）**
- 文件：`backend/app/api/v1/reports.py`
- 接口：生成报表（Mock 立即返回）、报表列表、下载（返回示例 Excel）、行动建议
- 新建 `backend/app/schemas/report.py`
- 预计：2 小时

#### 前端任务

**T-F06：修复 CRM `beforeUpload` + `previewMapping` Method（P1-4, P1-5）**
- 文件：`frontend-web/src/pages/crm/index.tsx`
- 删除 `beforeUpload={() => false}`
- `previewMapping` 改为 POST 请求
- 修复弹窗显示逻辑顺序（P2-6）
- 预计：1.5 小时

**T-F07：CRM 页面联调验收**
- 上传 Excel → 进度轮询 → 字段映射弹窗 → 归因确认弹窗 → 线索列表刷新
- 预计：1 小时

---

### Day 5：集成联调 + BUG 修复 + 冒烟测试

#### 后端任务

**T-B13：注册所有路由到 `main.py` + 异常处理完善（P3-2）**
- 确认所有 8 个路由已注册
- 添加 `EduAdCRMException` 自定义异常 Handler
- 统一 UTC 时间（P3-3）
- 预计：1 小时

**T-B14：P2-1 OAuth state 验证加固**
- 生产环境禁止无 state 的 OAuth 回调降级
- 预计：30 分钟

#### 前端任务

**T-F08：全流程冒烟测试**
1. 注册新租户 → 登录 → 进入引导向导
2. 引导步骤 1：填写机构信息 → 调用 `/onboarding/status`
3. 引导步骤 2：百度 OAuth 绑定 → 回调接收 → 触发 7 天数据同步
4. 仪表盘：演示模式切换、真实数据展示
5. CRM：上传 Excel → 字段映射 → 归因确认 → 线索列表
6. 登出 → Token 失效验证

**T-F09：Token 刷新并发锁（P3-1）**
- 文件：`frontend-web/src/api/request.ts`
- 添加 `refreshPromise` 锁避免并发刷新
- 预计：30 分钟

---

## 五、任务优先级矩阵

```
高优先级（阻断联调，必做）
├── T-B01  TenantContext 并发修复
├── T-B02  SQLite Mock 基础设施
├── T-B03  auth.py 认证路由（真实）
├── T-B04  路由注册到 main.py
├── T-B05  认证路由完善联调
├── T-B06  百度 OAuth 完整流程
├── T-F01  OAuth window.location.href 修复
├── T-F02  OAuth 回调结果处理
└── T-F03  登录注册页面联调

中优先级（功能可演示）
├── T-B07  onboarding.py Mock
├── T-B08  demo.py Mock
├── T-B09  analytics.py Mock
├── T-B10  settings.py Mock
├── T-B11  crm.py 混合实现
├── T-F05  仪表盘演示数据联调
└── T-F06  CRM 页面 Bug 修复

低优先级（完整体验）
├── T-B12  reports.py Mock
├── T-B13  异常处理完善
├── T-B14  OAuth state 验证
├── T-F07  CRM 联调验收
├── T-F08  全流程冒烟测试
└── T-F09  Token 刷新并发锁
```

---

## 六、文件创建清单

### 后端新增文件

```
backend/
├── mock_data.sqlite3              # SQLite Mock 数据库（gitignore）
├── app/
│   ├── mock/
│   │   ├── __init__.py
│   │   ├── mock_db.py            # SQLite 异步连接（aiosqlite）
│   │   └── seed.py               # Mock 数据生成脚本
│   ├── api/v1/
│   │   ├── auth.py               # 认证路由（真实逻辑）
│   │   ├── demo.py               # 演示数据（Mock）
│   │   ├── onboarding.py         # 引导状态（Mock）
│   │   ├── crm.py                # CRM 导入（真实+Mock混合）
│   │   ├── analytics.py          # 数据分析（Mock）
│   │   ├── reports.py            # 报表（Mock）
│   │   └── settings.py           # 设置（Mock）
│   ├── schemas/
│   │   ├── auth.py               # 认证 Schema
│   │   ├── demo.py               # 演示数据 Schema
│   │   ├── onboarding.py         # 引导状态 Schema
│   │   ├── crm.py                # CRM Schema
│   │   ├── analytics.py          # 分析 Schema
│   │   └── report.py             # 报表 Schema
│   └── models/
│       └── crm_lead.py           # CRM Lead ORM（基础版）
```

### 后端修改文件

```
backend/app/
├── main.py                       # 注册所有路由
└── core/
    └── tenant.py                 # ContextVar 修复
```

### 前端修改文件

```
frontend-web/src/
├── pages/
│   ├── ad-accounts/index.tsx     # OAuth window.href + 回调处理
│   ├── dashboard/index.tsx       # timeRange parseInt 修复 + EmptyState 对齐
│   └── crm/index.tsx             # beforeUpload + previewMapping + 弹窗顺序
└── api/
    └── request.ts                # Token 刷新并发锁
```

---

## 七、Mock 数据规格（SQLite）

### 7.1 Mock 表结构

```sql
-- 演示仪表盘数据（30天）
CREATE TABLE mock_dashboard_trend (
    date TEXT,
    spend REAL,
    leads INTEGER,
    platform TEXT  -- 'juliang' | 'baidu'
);

-- 演示线索列表
CREATE TABLE mock_crm_leads (
    id TEXT PRIMARY KEY,
    name TEXT,
    phone_masked TEXT,   -- 脱敏后手机号
    source_channel TEXT, -- '抖音' | '百度' | '其他'
    course_name TEXT,
    deal_status TEXT,    -- 'deal' | 'no_deal' | 'following'
    deal_amount REAL,
    created_at TEXT
);

-- 演示报表
CREATE TABLE mock_reports (
    id TEXT PRIMARY KEY,
    report_type TEXT,
    date_range_start TEXT,
    date_range_end TEXT,
    status TEXT,
    created_at TEXT
);

-- 引导状态（按 tenant_id 存储）
CREATE TABLE mock_onboarding (
    tenant_id TEXT PRIMARY KEY,
    step1_done INTEGER DEFAULT 0,
    step1_org_name TEXT,
    step2_done INTEGER DEFAULT 0,
    step3_done INTEGER DEFAULT 0,
    completed_at TEXT
);

-- 归因规则
CREATE TABLE mock_attribution_rules (
    id TEXT PRIMARY KEY,
    tenant_id TEXT,
    platform TEXT,
    keywords TEXT  -- JSON 数组
);
```

### 7.2 演示数据参数（"星海教育"）

```python
DEMO_CONFIG = {
    "org_name": "星海教育",
    "daily_spend_range": (8000, 15000),  # 元
    "cpe_range": (30, 50),               # 元/线索
    "roi_range": (3.0, 6.0),
    "juliang_spend_ratio": 0.6,          # 抖音占 60% 预算
    "baidu_spend_ratio": 0.4,
    "total_leads_30d": 1200,
    "deal_rate": 0.18,                   # 成交率 18%
}
```

---

## 八、接口实现策略一览

> **约定**：`[REAL]` = 真实业务逻辑 | `[MOCK]` = SQLite Mock | `[MIXED]` = 真实+Mock 混合

| 接口 | 实现策略 | 说明 |
|------|----------|------|
| `POST /auth/register` | **[REAL]** | 写入 MySQL，创建 Tenant + User |
| `POST /auth/login` | **[REAL]** | 验证密码，签发 JWT |
| `POST /auth/logout` | **[REAL]** | Token 写入 Redis 黑名单 |
| `POST /auth/refresh` | **[REAL]** | 验证 refresh_token，签发新 access_token |
| `GET /auth/profile` | **[REAL]** | 从 JWT 读取，查询 MySQL User |
| `GET /ad/juliang/oauth-url` | **[REAL]** | 已实现 |
| `GET /ad/baidu/oauth-url` | **[REAL]** | 已实现 |
| `GET /ad/baidu/callback` | **[REAL]** | 补全 Token 换取 + 7 天同步触发 |
| `GET /ad/accounts` | **[REAL]** | 已实现 |
| `POST /ad/accounts/{id}/sync` | **[REAL]** | 已实现 |
| `GET /demo/dashboard` | **[MOCK]** | 返回"星海教育"固定演示数据 |
| `GET /demo/metrics` | **[MOCK]** | 返回核心指标子集 |
| `GET /onboarding/status` | **[MIXED]** | 真实 tenant_id + SQLite 引导状态 |
| `POST /onboarding/status` | **[MIXED]** | 写入 SQLite |
| `POST /onboarding/complete` | **[MIXED]** | 标记完成，写入 SQLite |
| `POST /crm/upload` | **[REAL]** | 文件写入临时目录，触发 Celery |
| `GET /crm/upload/{task_id}` | **[REAL]** | 从 Redis 读取真实进度 |
| `POST /crm/upload/{batch_id}/preview` | **[MOCK]** | 返回模拟字段分析结果 |
| `POST /crm/upload/{batch_id}/confirm` | **[MOCK]** | 写入 SQLite，返回成功 |
| `GET /crm/upload/{batch_id}/attribution-suggest` | **[MOCK]** | 返回模拟归因建议 |
| `GET /crm/leads` | **[MOCK]** | 从 SQLite 返回模拟线索 |
| `PATCH /crm/leads/{id}` | **[MOCK]** | 更新 SQLite |
| `GET /crm/batches` | **[MOCK]** | 从 SQLite 返回批次列表 |
| `DELETE /crm/batches/{id}` | **[MOCK]** | 删除 SQLite 记录 |
| `GET /analytics/dashboard` | **[MOCK]** | 从 SQLite 聚合计算 |
| `GET /analytics/dashboard/status` | **[MIXED]** | 查询 MySQL 广告账户 + SQLite CRM 状态 |
| `GET /analytics/trend` | **[MOCK]** | SQLite 30 天趋势数据 |
| `GET /analytics/channel-compare` | **[MOCK]** | SQLite 渠道对比 |
| `GET /analytics/cross-table` | **[MOCK]** | SQLite 交叉分析 |
| `GET /analytics/cross-table/export` | **[MOCK]** | 生成简单 CSV |
| `POST /reports/generate` | **[MOCK]** | 返回 Mock task_id，秒完成 |
| `GET /reports` | **[MOCK]** | SQLite 报表列表 |
| `GET /reports/{id}/download` | **[MOCK]** | 返回示例 Excel bytes |
| `GET /reports/{id}/insights` | **[MOCK]** | 固定 3 条行动建议 |
| `GET /settings/tenant` | **[REAL]** | 查询 MySQL Tenant 表 |
| `PATCH /settings/tenant` | **[REAL]** | 更新 MySQL Tenant 表 |
| `GET /settings/attribution-rules` | **[MOCK]** | SQLite |
| `POST /settings/attribution-rules` | **[MOCK]** | SQLite |
| `PATCH /settings/attribution-rules/{id}` | **[MOCK]** | SQLite |
| `DELETE /settings/attribution-rules/{id}` | **[MOCK]** | SQLite |

---

## 九、验收标准（5 天结束时）

### 功能验收

| 测试项 | 预期结果 |
|--------|----------|
| 用户注册 | 200，自动创建 Tenant，返回 JWT |
| 用户登录 | 200，返回 access_token + refresh_token |
| Token 刷新 | 200，返回新 access_token |
| 登出 | 200，Token 加入黑名单，再次请求 401 |
| 未登录访问受保护接口 | 401 |
| 百度 OAuth 跳转 | 整页跳转，不弹新窗口 |
| 百度 OAuth 回调 | 跳回 `/ad-accounts?oauth_result=success` |
| 绑定后 7 天数据同步 | Celery 任务被触发（可在日志中确认） |
| 演示数据切换 | 仪表盘显示"星海教育"数据，有蓝色演示角标 |
| 引导向导 3 步 | 每步保存到 SQLite，刷新后状态保持 |
| CRM 上传 Excel | 文件接收成功，返回 batch_id + task_id |
| 字段映射弹窗 | 正确显示（先于归因弹窗） |
| 线索列表 | 返回 Mock 数据，分页正常 |
| 报表列表 | 返回 Mock 报表，行动建议显示 |
| 设置页租户信息 | 显示真实注册的机构名称 |

### 技术验收

- [ ] `uvicorn app.main:app --reload` 启动无报错
- [ ] `/docs` OpenAPI 文档中显示所有 8 个模块的接口
- [ ] 并发 10 个请求不出现 tenant_id 串行
- [ ] 前端 `npm run dev` 启动后，所有页面无 404 接口错误

---

## 十、后续迭代计划（Week 2+）

本次冲刺结束后，按以下顺序逐步将 Mock 接口替换为真实业务逻辑：

| 优先级 | 模块 | 说明 |
|--------|------|------|
| P0 | CRM 导入完整流程 | 补全 `crm_lead.py` ORM + openpyxl 解析 + SHA-256 去重 |
| P0 | Analytics 真实计算 | ROI 计算服务，广告花费 × CRM 成交关联查询 |
| P1 | Onboarding 持久化 | 迁移到 MySQL `onboarding_status` 表 |
| P1 | 归因规则真实存储 | 迁移到 MySQL `attribution_rules` 表 |
| P2 | 报表真实生成 | openpyxl 生成报表 + WeasyPrint PDF |
| P2 | P2-5 UUID 替换 | 所有模型 ID 改为 `uuid.uuid4()` |
| P3 | 小程序端 | Taro + 微信登录 |
| P3 | 邮件告警 | 数据同步失败邮件通知 |

---

*计划制定：2026-04-28 | 基于 CODE_REVIEW.md + 代码盘点 + 产品需求分析*
