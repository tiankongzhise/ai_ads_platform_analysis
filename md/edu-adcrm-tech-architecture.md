# EduAdCRM-MVP 技术方案文档

> **项目代号**：EduAdCRM-MVP  
> **文档版本**：v2.0（基于MVP v2.0任务列表全面修订）  
> **生成日期**：2026-04-28  
> **适用范围**：教育行业广告×CRM数据平台，MVP 阶段  
> **修订说明**：整合UX架构v1.0断点修复需求，纳入演示数据系统、3步引导向导、分阶段空态、智能归因等核心功能

---

## 一、原始技术栈问题分析与调整说明

### 1.1 原方案不合理之处（v1.0）

| 原方案 | 问题 | 调整方向 |
|--------|------|----------|
| Laravel 10 + PHP 8.2 | 用户要求后端优先 Python | 替换为 Python FastAPI |
| Livewire 3 + FluxUI | 服务端渲染，无法支持 App / 小程序多端复用 | 替换为前后端分离，React/Vue + Taro 小程序 |
| Laravel Queue（数据库驱动）| 数据库队列在大文件处理场景不稳定，高并发时易阻塞 | 替换为 Celery + Redis |
| Playwright 截图测试 sh 脚本 | 硬编码路径，不可移植，不属于业务需求 | 移除该约束，改为 pytest + 标准化 CI |
| 交付形式"排除移动 App" | 与前端多端适配要求冲突 | 架构设计支持 Web / App / 小程序三端 |
| ApexCharts（Livewire绑定）| 依赖 Livewire 响应式，脱离后端渲染后无优势 | 改用 ECharts（生态更强，小程序有官方版本）|

### 1.2 v2.0 主要技术演进

| 演进点 | v1.0 | v2.0 | 原因 |
|--------|------|------|------|
| 演示数据系统 | 无 | B1 新增 | UX BP-1：注册→空仪表盘致命断点 |
| 引导向导 | 分散配置 | B2 3步向导 | UX BP-2：引导不连价值链 |
| 仪表盘空态 | 空态死胡同 | B3 分阶段空态 | UX BP-5：空态无引导 |
| 智能字段映射 | 纯手动 | C2 智能匹配 | UX BP-4：归因黑洞 |
| 归因确认 | 设置页手动 | E2 自动弹窗确认 | UX BP-4：隐性依赖 |
| 广告同步触发 | 仅凌晨定时 | D1 绑定后立即拉取 | UX BP-3：绑定后无数据 |
| 行动建议 | 无 | F1 规则型建议 | UX BP-7：报表无行动 |
| 导航架构 | 平铺菜单 | G1 分组+优先级 | UX架构重设计 |

---

## 二、系统架构总览

### 2.1 架构风格

采用 **前后端分离 + API 驱动** 的单体分层架构（Modular Monolith）。

- **MVP 阶段不拆微服务**：团队小、边界未稳定，过早拆分是架构宇航员陷阱
- **模块边界清晰**：Auth / CRM / AdAccount / Analytics / Report / Onboarding 六个模块内聚，接口契约稳定后可按需独立部署
- **API First**：后端只提供 RESTful JSON API，三端（Web / App / 小程序）共用同一套接口

### 2.2 整体分层图

```
┌─────────────────────────────────────────────────────────────┐
│                        客户端层                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐   │
│  │  Web (PC/H5) │  │  Mobile App  │  │  微信小程序       │   │
│  │  React + Ant │  │  React Native│  │  Taro (React)    │   │
│  │  Design Pro  │  │  (Phase 2)  │  │  + Vant Weapp   │   │
│  └──────┬───────┘  └──────┬───────┘  └────────┬─────────┘   │
└─────────┼─────────────────┼───────────────────┼─────────────┘
          │                 │                   │
          └─────────────────┼───────────────────┘
                            │  HTTPS / REST API + JWT
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                       API 网关层                              │
│              Nginx (反向代理 + SSL 终止 + 限流)               │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────┐
│                      后端应用层                               │
│                  Python FastAPI (Uvicorn)                    │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐   │
│  │   Auth   │ │   CRM    │ │ AdAccount│ │  Analytics   │   │
│  │  模块    │ │  模块    │ │   模块   │ │    模块      │   │
│  └──────────┘ └──────────┘ └──────────┘ └──────────────┘   │
│  ┌──────────┐ ┌──────────┐ ┌────────────────────────────────┐ │
│  │  Report  │ │Onboarding│ │         任务调度 (Celery)      │ │
│  │  模块    │ │  模块    │ │  每日广告拉取 / 月报生成        │ │
│  └──────────┘ └──────────┘ │  绑定后立即拉取 / 导入处理      │ │
│                             └────────────────────────────────┘ │
└──────────────────────────┬──────────────────────────────────┘
                           │
          ┌────────────────┼────────────────┐
          ▼                ▼                ▼
┌─────────────┐  ┌──────────────┐  ┌──────────────┐
│  MySQL 8.0  │  │    Redis     │  │  对象存储     │
│  (主数据库) │  │  (缓存+队列) │  │  COS/MinIO   │
│             │  │  Celery队列  │  │  Excel文件    │
└─────────────┘  └──────────────┘  └──────────────┘
                           │
          ┌────────────────┼────────────────┐
          ▼                                 ▼
┌──────────────────┐              ┌──────────────────┐
│  Celery Worker   │              │   外部广告 API    │
│  (异步任务处理)  │              │  巨量引擎 / 百度  │
│  - Excel解析      │              │  营销 API        │
│  - 广告数据同步   │              └──────────────────┘
│  - 报表生成       │
│  - 演示数据      │  ← v2.0新增
└──────────────────┘
```

---

## 三、技术栈选型

### 3.1 后端

| 组件 | 选型 | 版本 | 选型理由 |
|------|------|------|----------|
| Web 框架 | **FastAPI** | 0.111+ | 原生异步、自动生成 OpenAPI 文档、性能优、Python 生态最佳 REST 选择 |
| ASGI 服务器 | **Uvicorn** | 0.29+ | FastAPI 官方推荐，生产环境配合 Gunicorn 多进程 |
| ORM | **SQLAlchemy 2.x** | 2.0+ | 成熟稳定，支持异步，迁移工具 Alembic 完善 |
| 数据库迁移 | **Alembic** | - | 与 SQLAlchemy 原生集成 |
| 异步任务队列 | **Celery + Redis** | Celery 5.x | 替代数据库队列驱动，稳定可靠，支持定时任务（Celery Beat） |
| 缓存 | **Redis** | 7.x | 接口缓存、Token 黑名单、Celery Broker、导入进度 五用合一 |
| Excel 解析 | **openpyxl + pandas** | - | Python 原生方案，5万行 CSV/Excel 处理稳定 |
| 认证 | **python-jose + passlib** | - | JWT 生成与校验，bcrypt 密码哈希 |
| HTTP 客户端 | **httpx** | - | 异步 HTTP，对接巨量引擎/百度营销 API |
| 邮件 | **FastAPI-Mail** | - | 异步邮件发送，支持 SMTP |
| 报表生成 | **openpyxl** | - | Excel 报表；PDF 用 **WeasyPrint** |
| 定时任务 | **Celery Beat** | - | 广告数据每日自动拉取、月报生成 |
| 日志 | **structlog** | - | 结构化日志，便于后续接入 ELK |
| 测试 | **pytest + httpx** | - | API 集成测试 |

### 3.2 前端

| 场景 | 选型 | 选型理由 |
|------|------|----------|
| **Web PC / H5 响应式** | **React 18 + Ant Design Pro 6** | 成熟中后台方案，响应式布局，图表生态完整 |
| **微信小程序** | **Taro 4（React 语法）** | 一套 React 代码编译至小程序，与 Web 端共享业务逻辑和组件 |
| **移动 App（Phase 2）** | **React Native（Expo）** | 与 Web/小程序共享部分逻辑，MVP 阶段可暂缓 |
| **图表** | **ECharts（Apache）+ echarts-for-react** | Web 端用 echarts-for-react；小程序用官方 ec-canvas；功能全面 |
| **状态管理** | **Zustand** | 轻量，比 Redux 简单，适合 MVP 快速开发 |
| **请求库** | **Axios + React Query** | Axios 统一封装 API 请求，React Query 处理缓存和加载态 |
| **样式** | **Tailwind CSS + Ant Design** | Ant Design 提供组件，Tailwind 处理定制布局 |
| **构建工具** | **Vite** | 开发体验快，打包效率高 |

### 3.3 数据库

| 组件 | 选型 | 说明 |
|------|------|------|
| 主数据库 | **MySQL 8.0** | 与现有宝塔环境一致，保留原方案 |
| 缓存/队列 | **Redis 7.x** | Celery broker + API 缓存 + Session + 导入进度 |
| 文件存储 | **腾讯云 COS / 本地 MinIO** | Excel 上传文件持久化，不占数据库 |

### 3.4 基础设施

| 组件 | 选型 | 说明 |
|------|------|------|
| 反向代理 | **Nginx** | 与现有宝塔环境一致 |
| 容器化 | **Docker + docker-compose** | 开发/生产环境一致性，宝塔支持 Docker |
| 进程管理 | **Supervisor** | 管理 Uvicorn / Celery Worker / Celery Beat 进程 |
| SSL | **Let's Encrypt（宝塔一键）** | HTTPS 必须，微信小程序强制要求 |

---

## 四、核心功能模块设计（v2.0 新增）

### 4.1 演示数据系统（B1）

**目的**：解决 UX BP-1（注册→空仪表盘）致命断点，让用户注册后5分钟内看到产品价值。

#### 后端实现

```python
# backend/app/api/v1/demo.py
@router.get("/demo/dashboard")
async def get_demo_dashboard():
    """
    返回教育行业演示数据（星海教育虚构机构）
    数据特征：CPE ¥30-50, ROI 1:3-6, 日花费 ¥8K-15K
    """
    return {
        "mode": "demo",
        "org_name": "星海教育",
        "metrics": {
            "today_spend": 12580,
            "total_leads": 328,
            "cost_per_lead": 38.4,
            "roi": 4.2,
            "conversion_rate": 2.8
        },
        "trends": [...],  # 近30天花费vs线索趋势
        "channel_compare": [...],  # 抖音vs百度对比
        "campaign_ranking": [...]  # 广告计划排行
    }
```

#### 前端实现

- **DemoDataToggle 组件**：Header 右上角切换开关
- **演示数据标识**：蓝色虚线边框 + "演示"角标
- **存储**：Zustand `demoStore.ts`，记录当前模式

### 4.2 3步引导向导（B2）

**目的**：解决 UX BP-2（引导不连价值链）致命断点，将分散配置串联为连贯路径。

#### 引导状态模型

```python
# backend/app/models/onboarding_status.py
class OnboardingStatus(Base):
    __tablename__ = "onboarding_status"
    
    tenant_id = Column(String, primary_key=True)
    step1_org_info = Column(JSON)  # {name, industry_sub_type, main_product}
    step2_ad_bound = Column(JSON)   # {platforms: ["juliang", "baidu"]}
    step3_crm_imported = Column(Boolean, default=False)
    completed_at = Column(DateTime)
    
# API
POST /api/v1/onboarding/status    # 更新引导状态
GET  /api/v1/onboarding/status    # 获取当前进度
```

#### 前端组件

| 组件 | 文件 | 说明 |
|------|------|------|
| `OnboardingWizard` | `components/OnboardingWizard/` | 3步弹窗向导 |
| `SetupProgressCard` | `components/SetupProgressCard/` | 仪表盘底部固定进度卡片 |

### 4.3 仪表盘分阶段空态（B3）

**目的**：解决 UX BP-5（仪表盘空态）致命断点，将"暂无数据"死胡同变为行动引导。

#### 空态场景判定逻辑

```python
# backend/app/api/v1/analytics.py
def get_dashboard_status(tenant_id: str) -> dict:
    """
    返回仪表盘数据状态，决定展示哪种空态
    """
    has_ad_account = check_ad_accounts(tenant_id)
    has_ad_data = check_ad_data_sync_status(tenant_id)
    has_crm_data = check_crm_leads(tenant_id)
    sync_in_progress = check_sync_task_status(tenant_id)
    
    if not has_ad_account and not has_crm_data:
        return "scenario_a"  # 未绑定广告+未导入CRM
    elif sync_in_progress:
        return "scenario_b"  # 广告同步中
    elif has_ad_data and not has_crm_data:
        return "scenario_c"  # 有广告数据，CRM未导入
    else:
        return "full_data"   # 完整数据
```

#### 前端空态组件

```tsx
// frontend-web/src/components/EmptyState/
interface EmptyStateProps {
  scenario: 'a' | 'b' | 'c';
  onAction: () => void;
  onViewDemo: () => void;
}
```

### 4.4 智能字段映射（C2）

**目的**：解决 UX BP-4（归因黑洞）严重断点，降低Excel导入门槛。

#### 模糊匹配算法

```python
# frontend-web/src/utils/fieldMatcher.ts
function matchField(userColumns: string[], standardFields: Field[]): MatchResult[] {
    // 常见别名映射
    const aliases = {
        "姓名": ["姓名", "name", "名字", "学员姓名", "学生姓名", "客户姓名"],
        "手机号": ["手机号", "手机", "电话", "tel", "phone", "联系方式", "mobile"],
        "来源渠道": ["渠道", "来源", "source", "channel", "广告渠道", "报名渠道"],
        // ...
    }
    
    // 使用编辑距离 + 关键词匹配计算置信度
    // 返回 MatchResult[]: { field, matchedColumn, confidence: 'auto' | 'suggested' | 'none' }
}
```

### 4.5 归因规则自动确认（E2）

**目的**：解决 UX BP-4（归因黑洞）严重断点，在导入时自动分析并确认归因规则。

#### 自动归因流程

```python
# backend/app/services/attribution.py
class AttributionEngine:
    def analyze_and_suggest(self, tenant_id: str, batch_id: str) -> AttributionSuggestion:
        """
        1. 分析CRM数据中来源字段的值分布
        2. 匹配到已知广告平台关键词
        3. 返回未匹配数据量及处理建议
        """
        pass
    
    def confirm_rules(self, tenant_id: str, rules: list[AttributionRule]):
        """
        确认归因规则，开始ROI计算
        """
        pass

# API
POST /api/v1/crm/upload/{batch_id}/attribution-suggest  # 获取归因建议
POST /api/v1/crm/upload/{batch_id}/confirm              # 确认归因+开始导入
```

### 4.6 广告绑定后立即拉取（D1）

**目的**：解决 UX BP-3（绑定后无数据）严重断点，绑定后立即给用户看到数据的希望。

#### 触发逻辑

```python
# backend/app/api/v1/ad_accounts.py
@router.get("/ad/juliang/callback")
async def juliang_oauth_callback(code: str, ...):
    # 1. 存储access_token
    # 2. 创建ad_account记录
    
    # 3. 立即触发Celery任务拉取7天历史数据（关键改进）
    from app.tasks.sync_ad_tasks import sync_ad_data_juliang
    sync_ad_data_juliang.delay(account_id, days=7)  # 立即执行，非等凌晨
    
    # 4. 返回引导向导下一步（CRM导入）
    return {"status": "success", "next_step": "crm_import"}
```

---

## 五、数据库设计（核心表）

### 5.1 多租户设计原则

MVP 阶段采用 **共享数据库 + 行级隔离**（shared database, separate rows）：
- 所有核心表含 `tenant_id` 字段
- 所有查询通过 SQLAlchemy 中间件自动注入 `WHERE tenant_id = ?` 过滤
- 简单、低成本，满足 MVP 验证阶段需求

### 5.2 核心数据表（v2.0 新增引导状态表）

```sql
-- 租户表
tenants (id, name, industry_sub_type, main_product, created_at)

-- 用户表
users (id, tenant_id, email, password_hash, role, is_active, created_at)

-- 引导状态表（v2.0新增）
onboarding_status (
  tenant_id PK, 
  step1_org_info JSON,
  step2_ad_bound JSON,
  step3_crm_imported BOOLEAN,
  completed_at TIMESTAMP
)

-- 广告账户表
ad_accounts (
  id, tenant_id, platform ENUM('juliang','baidu'),
  account_id, account_name, access_token, refresh_token,
  token_expires_at, balance, status, created_at
)

-- 广告每日数据
ad_daily_stats (
  id, tenant_id, ad_account_id, platform,
  date, campaign_id, campaign_name, adgroup_id, adgroup_name,
  spend, impressions, clicks, form_submissions,
  cost_per_click, cost_per_form, created_at
)

-- CRM 线索批次
crm_import_batches (
  id, tenant_id, filename, file_path, total_rows,
  success_rows, skip_rows, fail_rows, status, created_at
)

-- CRM 线索明细
crm_leads (
  id, tenant_id, batch_id,
  name, phone_hash(sha256脱敏存储), source_channel,
  course_name, deal_status ENUM('deal','no_deal','following'),
  deal_amount, deal_date, raw_source_value, created_at
)

-- 渠道归因规则
attribution_rules (
  id, tenant_id, platform ENUM('juliang','baidu','other'),
  keywords JSON,  -- ["抖音","douyin","tiktok"]
  created_at
)

-- 操作日志
audit_logs (
  id, tenant_id, user_id, action, resource_type, resource_id,
  detail JSON, ip_address, created_at
)
```

---

## 六、后端模块设计（FastAPI）

### 6.1 项目结构（v2.0 更新）

```
backend/
├── app/
│   ├── main.py                 # FastAPI 入口，注册路由
│   ├── core/
│   │   ├── config.py           # 配置（pydantic-settings，env 优先）
│   │   ├── database.py         # SQLAlchemy engine & session
│   │   ├── security.py         # JWT 生成/验证
│   │   ├── tenant.py           # 多租户中间件（自动注入 tenant_id）
│   │   └── exceptions.py       # 统一错误处理
│   ├── models/                 # SQLAlchemy ORM 模型
│   │   ├── user.py
│   │   ├── onboarding_status.py  # v2.0新增：引导状态
│   │   ├── ad_account.py
│   │   ├── crm_lead.py
│   │   └── ...
│   ├── schemas/                # Pydantic 请求/响应模型
│   │   ├── auth.py
│   │   ├── onboarding.py       # v2.0新增：引导相关
│   │   ├── crm.py
│   │   └── analytics.py
│   ├── api/
│   │   └── v1/
│   │       ├── auth.py         # 注册/登录/刷新Token
│   │       ├── demo.py         # v2.0新增：演示数据API
│   │       ├── onboarding.py   # v2.0新增：引导状态API
│   │       ├── crm.py          # Excel上传/线索管理
│   │       ├── ad_accounts.py  # 广告账户OAuth/管理（含立即拉取）
│   │       ├── analytics.py    # ROI仪表盘/交叉分析
│   │       ├── reports.py      # 报表生成/下载
│   │       └── settings.py     # 归因规则/账户设置
│   ├── services/
│   │   ├── demo_data.py        # v2.0新增：演示数据生成（教育行业模板）
│   │   ├── juliang_service.py  # 巨量引擎 API 封装
│   │   ├── baidu_service.py    # 百度营销 API 封装
│   │   ├── crm_importer.py     # Excel解析+字段映射
│   │   ├── field_matcher.py    # v2.0新增：智能字段匹配
│   │   ├── roi_calculator.py   # ROI 计算逻辑
│   │   ├── attribution.py       # 线索归因引擎（自动分析）
│   │   ├── insight_engine.py   # v2.0新增：规则型行动建议
│   │   └── report_generator.py # Excel/PDF 报表生成
│   ├── tasks/                  # Celery 异步任务
│   │   ├── celery_app.py       # Celery 初始化
│   │   ├── import_tasks.py     # Excel 异步导入
│   │   ├── sync_ad_tasks.py    # 广告数据同步（含立即拉取模式）
│   │   └── report_tasks.py     # 月报生成任务
│   └── utils/
│       ├── email.py
│       └── file_storage.py     # COS/MinIO 封装
├── alembic/                    # 数据库迁移
├── tests/                      # pytest 测试
├── requirements.txt
├── pyproject.toml
├── Dockerfile
└── docker-compose.yml
```

### 6.2 核心 API 接口清单（v2.0 新增接口）

#### Auth 模块
```
POST /api/v1/auth/register          # 邮箱注册
POST /api/v1/auth/login             # 登录，返回 access_token + refresh_token
POST /api/v1/auth/refresh           # 刷新 Token
POST /api/v1/auth/logout            # 登出（Redis 黑名单）
POST /api/v1/auth/forgot-password   # 忘记密码
POST /api/v1/auth/reset-password    # 重置密码
POST /api/v1/auth/wx-login          # 微信小程序登录（code换JWT）
GET  /api/v1/auth/profile           # 当前用户信息
```

#### Demo 模块（v2.0 新增）
```
GET  /api/v1/demo/dashboard         # 获取演示仪表盘数据（星海教育）
GET  /api/v1/demo/metrics           # 获取演示核心指标
```

#### Onboarding 模块（v2.0 新增）
```
GET  /api/v1/onboarding/status      # 获取引导进度
POST /api/v1/onboarding/status      # 更新引导状态（Step1/Step2/Step3）
POST /api/v1/onboarding/complete    # 完成引导
```

#### CRM 模块（v2.0 增强）
```
POST /api/v1/crm/upload             # 上传 Excel/CSV 文件（触发异步任务）
GET  /api/v1/crm/upload/{task_id}   # 查询上传任务进度
POST /api/v1/crm/upload/{batch_id}/preview  # 预览+智能字段匹配
POST /api/v1/crm/upload/{batch_id}/confirm  # 确认字段映射+归因规则+开始导入
GET  /api/v1/crm/upload/{batch_id}/attribution-suggest  # v2.0新增：获取归因建议
GET  /api/v1/crm/leads              # 线索列表（分页+筛选）
PATCH /api/v1/crm/leads/{id}        # 更新线索状态
DELETE /api/v1/crm/batches/{id}     # 删除整批导入记录
GET  /api/v1/crm/batches            # 导入批次列表
```

#### 广告账户模块（v2.0 增强）
```
GET  /api/v1/ad/accounts            # 广告账户列表
GET  /api/v1/ad/juliang/oauth-url   # 获取巨量引擎授权 URL
GET  /api/v1/ad/juliang/callback    # OAuth 回调处理（触发立即拉取7天历史）
GET  /api/v1/ad/baidu/oauth-url     # 获取百度营销授权 URL
GET  /api/v1/ad/baidu/callback      # OAuth 回调处理（触发立即拉取7天历史）
DELETE /api/v1/ad/accounts/{id}     # 断开广告账户
POST /api/v1/ad/accounts/{id}/sync  # 手动触发数据同步
GET  /api/v1/ad/sync-status         # v2.0新增：同步状态（用于仪表盘空态判定）
```

#### Analytics 模块
```
GET  /api/v1/analytics/dashboard    # ROI 仪表盘核心指标（含数据状态）
GET  /api/v1/analytics/dashboard/status  # v2.0新增：仪表盘数据状态（决定空态场景）
GET  /api/v1/analytics/trend        # 近N天花费 vs 线索趋势
GET  /api/v1/analytics/channel-compare   # 渠道对比（抖音 vs 百度）
GET  /api/v1/analytics/cross-table  # 广告计划×课程交叉分析
GET  /api/v1/analytics/cross-table/export  # 导出 CSV
```

#### Reports 模块（v2.0 增强）
```
POST /api/v1/reports/generate       # 手动生成报表（指定时间范围）
GET  /api/v1/reports                # 报表列表
GET  /api/v1/reports/{id}/download  # 下载 Excel/PDF
GET  /api/v1/reports/{id}/insights  # v2.0新增：获取报表行动建议
```

---

## 六A、广告平台 OAuth 回调传值机制（新增）

> **本章解答**：后端如何获取广告平台回调链接传递的 `code` 和 `state` 参数，并以此完成 Token 换取和用户身份还原。

### 6A.1 OAuth 回调的本质

广告平台（巨量引擎 / 百度营销）采用标准 **OAuth 2.0 Authorization Code Flow**。用户在广告平台完成授权后，平台会**将值以 Query String 形式拼在 `redirect_uri` 上**，向我们的服务器发起一次 **HTTP GET** 请求：

```
GET /api/v1/ad/juliang/callback?code=AUTH_CODE_XXX&state=RANDOM_STATE_XXX
                                ↑                  ↑
                        一次性授权码            防 CSRF 随机串
```

FastAPI 用 `Query(...)` 注解自动从 URL 解析这两个参数：

```python
@router.get("/juliang/callback")
async def juliang_oauth_callback(
    code: str = Query(..., description="巨量引擎回传的一次性授权码"),
    state: Optional[str] = Query(None, description="防 CSRF 状态参数"),
    error: Optional[str] = Query(None, description="用户拒绝时平台传入的错误码"),
    db: AsyncSession = Depends(get_db),
):
    ...
```

### 6A.2 完整回调链路

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        OAuth 回调完整链路                                  │
│                                                                          │
│  1. 前端点击"绑定巨量引擎"                                                │
│       │                                                                  │
│       ▼                                                                  │
│  2. GET /api/v1/ad/juliang/oauth-url  （需 JWT 认证）                    │
│       │ 后端：生成 state = secrets.token_urlsafe(32)                     │
│       │        Redis.setex("oauth_state:{state}", 600,                  │
│       │                    "{tenant_id}:{user_id}")                      │
│       │ 返回：{"oauth_url": "https://open.oceanengine.com/authorize?     │
│       │               app_id=xxx&redirect_uri=xxx&state=STATE"}         │
│       │                                                                  │
│       ▼                                                                  │
│  3. 前端 window.location.href = oauth_url                               │
│       │（浏览器跳转到巨量引擎授权页）                                      │
│       │                                                                  │
│       ▼                                                                  │
│  4. 用户在广告平台完成授权                                                 │
│       │                                                                  │
│       ▼                                                                  │
│  5. 广告平台 → 302 跳转 →                                                │
│       GET /api/v1/ad/juliang/callback?code=AUTH_CODE&state=STATE        │
│       │ ← 这里就是"回调链接传值"的发生点 →                               │
│       │                                                                  │
│       ▼                                                                  │
│  6. 后端 juliang_oauth_callback() 处理：                                 │
│       a. 从 URL Query 读取 code、state                                   │
│       b. Redis.getdel("oauth_state:{state}") → 还原 tenant_id            │
│       c. 用 code 调用巨量引擎 Token 接口，换取 access_token              │
│       d. 获取广告主列表，写入 AdAccount 表                               │
│       e. 触发 Celery: sync_ad_data_juliang.delay(account_id, days=7)    │
│       f. Redis.setex("ad_sync_status:{account_id}", ...) 记录进度        │
│       │                                                                  │
│       ▼                                                                  │
│  7. 302 重定向 →                                                         │
│       {FRONTEND_URL}/ad-accounts?oauth_result=success&account_id=xxx    │
│       │（浏览器跳转回前端，前端读取 Query 参数展示结果）                   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### 6A.3 关键参数说明

| 参数 | 来源 | 说明 |
|------|------|------|
| `code` | 广告平台拼入回调 URL | 一次性授权码，有效期约 5 分钟，调用 Token 接口兑换后失效 |
| `state` | 广告平台原样回传 | 我们在发起授权时生成的随机串，用于**防 CSRF 攻击**和**还原 tenant_id** |
| `error` | 广告平台拼入回调 URL | 用户拒绝授权时传入（如 `access_denied`），此时没有 `code` |

### 6A.4 state 防 CSRF + tenant_id 还原机制

**问题**：OAuth 回调是广告平台直接 GET 我们的接口，不携带 JWT，后端无法知道是哪个用户（租户）发起了授权。

**解决方案**：发起授权时将 `state → tenant_id` 的映射**提前存入 Redis**，回调时通过 `state` 反查：

```python
# 发起授权（app/api/v1/ad_accounts.py）
state = secrets.token_urlsafe(32)
await save_oauth_state(
    state=state,
    tenant_id=current_user["tenant_id"],
    user_id=current_user["user_id"],
)
# Redis key: oauth_state:{state}
# Redis value: {"tenant_id": "xxx", "user_id": "yyy"}
# TTL: 600 秒（10 分钟）

# 回调处理
user_ctx = await consume_oauth_state(state)   # getdel，读取后立即删除（防重放）
tenant_id = user_ctx["tenant_id"]
```

**`save_oauth_state` / `consume_oauth_state` 实现位置**：`app/core/redis_client.py`

### 6A.5 回调后的状态传递（前端感知）

回调处理完成后，后端通过 **302 重定向**携带结果给前端：

```
# 成功
{FRONTEND_URL}/ad-accounts?oauth_result=success&platform=juliang&account_id=xxx

# 用户取消
{FRONTEND_URL}/ad-accounts?oauth_result=cancelled&platform=juliang

# 失败
{FRONTEND_URL}/ad-accounts?oauth_result=error&platform=juliang&reason=xxx
```

前端在 `/ad-accounts` 页面加载时读取 `location.search` 中的 `oauth_result` 展示对应 Toast。

### 6A.6 同步状态跟踪（Redis → 前端轮询）

OAuth 回调触发 Celery 任务后，通过 Redis 跟踪进度，前端轮询 `GET /api/v1/ad/sync-status`：

```
Redis key: ad_sync_status:{account_id}
Redis value: {
  "status":      "running" | "success" | "error",
  "progress":    0-100,
  "synced_count": int,
  "error_msg":   null | string,
  "started_at":  ISO datetime,
  "finished_at": ISO datetime | null
}
TTL: 86400 秒（24小时）
```

**写入时机**：
- 触发 Celery 时（ad_accounts.py）：写入 `running` 状态
- 任务成功完成（sync_ad_tasks.py）：更新为 `success` + `synced_count`
- 任务失败超重试（sync_ad_tasks.py）：更新为 `error` + `error_msg`

### 6A.7 新增/修改文件清单

| 文件 | 变更类型 | 说明 |
|------|---------|------|
| `app/core/redis_client.py` | **新增** | Redis 连接池、OAuth state 存取、同步状态读写、JWT 黑名单 |
| `app/services/baidu_service.py` | **新增** | 百度营销 API 封装（OAuth + 数据拉取） |
| `app/api/v1/ad_accounts.py` | **完善** | 补全 Redis state 存取、百度回调、302重定向、幂等写入 |
| `app/tasks/sync_ad_tasks.py` | **完善** | 任务成功/失败后写回 Redis 同步状态 |
| `app/core/config.py` | **完善** | 新增 `FRONTEND_URL` 配置项 |
| `app/main.py` | **完善** | 生命周期中加入 Redis 预热与关闭 |

---

## 七、异步任务设计（Celery）

### 7.1 任务清单（v2.0 更新）

| 任务名 | 触发方式 | 说明 |
|--------|----------|------|
| `process_crm_import` | 用户上传后异步触发 | 解析 Excel/CSV，去重，写入 DB |
| `sync_ad_data_juliang` | Celery Beat 每日 02:00 **或** 绑定后立即 | 拉取巨量引擎昨日/最近7天广告数据 |
| `sync_ad_data_baidu` | Celery Beat 每日 02:30 **或** 绑定后立即 | 拉取百度昨日/最近7天广告数据 |
| `generate_monthly_report` | Celery Beat 每月1日 08:00 | 生成上月 ROI 月报，发送邮件 |
| `send_sync_failure_alert` | 任务失败后触发 | 发送失败通知邮件给用户 |

### 7.2 Excel 导入流程（v2.0 增强）

```
用户上传文件
    │
    ▼
FastAPI 接收 → 存入 COS/MinIO → 返回 batch_id
    │
    ▼ (异步)
Celery Worker
    ├── 读取文件（openpyxl / pandas）
    ├── 解析字段映射（智能匹配建议）
    ├── 去重检查（按 phone_hash）
    ├── 批量写入 DB（每500行一批）
    └── 更新任务状态（Redis 存储进度）
    │
    ▼
前端轮询 /api/v1/crm/upload/{task_id}
    └── 返回进度百分比 + 状态 + 汇总结果
    │
    ▼
归因规则确认弹窗（AttributionConfirm）
    │
    ▼
用户确认归因规则 → 开始正式导入 → 完成后跳转仪表盘
```

### 7.3 演示数据生成任务（v2.0 新增）

```python
# backend/app/tasks/demo_tasks.py
@celery_app.task
def generate_demo_data(tenant_id: str):
    """
    为新注册用户生成演示数据
    存储在Redis或专用表，不污染真实数据
    """
    pass
```

---

## 八、前端架构设计

### 8.1 Web 端（React + Ant Design Pro）

```
frontend-web/
├── src/
│   ├── api/             # Axios 封装 + API 接口定义
│   ├── components/      # 共用组件
│   │   ├── RoiCard/     # 指标卡片
│   │   ├── TrendChart/  # ECharts 折线图封装
│   │   ├── ChannelBar/  # 渠道对比柱状图
│   │   ├── EmptyState/  # v2.0新增：分阶段空态组件
│   │   ├── DemoDataToggle/  # v2.0新增：演示/真实切换
│   │   ├── OnboardingWizard/  # v2.0新增：3步引导向导
│   │   ├── SetupProgressCard/ # v2.0新增：配置进度卡片
│   │   ├── SmartFieldMapper/  # v2.0新增：智能字段映射
│   │   ├── AttributionConfirm/ # v2.0新增：归因确认弹窗
│   │   ├── ProgressOverlay/   # v2.0新增：导入进度（可最小化）
│   │   ├── InsightCard/       # v2.0新增：行动建议卡片
│   │   └── ContextLink/       # v2.0新增：上下文跳转链接
│   ├── pages/
│   │   ├── login/
│   │   ├── onboarding/  # 引导配置（v2.0增强）
│   │   ├── dashboard/   # ROI 仪表盘（v2.0增强：含分阶段空态）
│   │   ├── crm/         # CRM 导入+管理（v2.0增强：智能映射）
│   │   ├── ad-accounts/ # 广告账户管理
│   │   ├── analytics/   # 交叉分析（v2.0增强：上下文跳转）
│   │   ├── reports/     # 报表（v2.0增强：行动建议）
│   │   └── settings/   # 归因规则设置
│   ├── store/           # Zustand 状态
│   │   ├── authStore.ts
│   │   ├── demoStore.ts  # v2.0新增：演示模式状态
│   │   └── onboardingStore.ts  # v2.0新增：引导进度状态
│   ├── hooks/           # 自定义 hooks
│   │   ├── useImportProgress.ts  # v2.0新增：导入进度轮询
│   │   └── useDashboardStatus.ts # v2.0新增：仪表盘状态
│   └── utils/
│       ├── fieldMatcher.ts  # v2.0新增：智能字段匹配算法
│       └── ...
├── package.json
└── vite.config.ts
```

### 8.2 微信小程序端（Taro）

```
frontend-weapp/
├── src/
│   ├── api/             # 与 Web 端共享的 API 层（条件编译适配wx.request）
│   ├── pages/
│   │   ├── index/       # 首页（TabBar）
│   │   ├── dashboard/   # ROI 仪表盘（主要功能）
│   │   ├── crm/         # 线索列表（只读为主）
│   │   ├── reports/     # v2.0新增：报表查看Tab
│   │   └── profile/    # 我的/账户
│   ├── components/
│   │   ├── EchartsWX/  # ec-canvas 封装
│   │   └── MetricCard/ # 指标卡片（2列布局）
│   └── app.config.ts   # 小程序配置（4个TabBar）
├── project.config.json
└── package.json
```

---

## 九、部署架构

### 9.1 生产环境部署（宝塔 + Docker）

```
服务器（云主机，推荐 4C8G 起步）
├── Nginx（宝塔管理）
│   ├── 反向代理 → FastAPI (8000端口)
│   ├── 静态文件 → Web 前端 dist/
│   └── SSL 证书（Let's Encrypt，宝塔一键申请）
│
├── Docker Compose 管理以下服务：
│   ├── app（FastAPI + Uvicorn，Gunicorn多进程）
│   ├── celery-worker（Celery 任务处理）
│   ├── celery-beat（定时任务调度）
│   ├── redis（7.x）
│   └── flower（Celery 监控，可选）
│
├── MySQL 8.0（宝塔直接安装管理）
└── MinIO（对象存储，或直接用腾讯云COS）
```

### 9.2 docker-compose.yml（关键配置）

```yaml
version: '3.9'
services:
  app:
    build: .
    command: gunicorn app.main:app -w 4 -k uvicorn.workers.UvicornWorker -b 0.0.0.0:8000
    environment:
      - DATABASE_URL=mysql+aiomysql://user:pass@host:3306/eduadcrm
      - REDIS_URL=redis://redis:6379/0
    depends_on:
      - redis
    restart: always

  celery-worker:
    build: .
    command: celery -A app.tasks.celery_app worker --loglevel=info -c 4
    depends_on:
      - redis
    restart: always

  celery-beat:
    build: .
    command: celery -A app.tasks.celery_app beat --loglevel=info
    depends_on:
      - redis
    restart: always

  redis:
    image: redis:7-alpine
    restart: always
    volumes:
      - redis_data:/data

volumes:
  redis_data:
```

---

## 十、安全设计

| 安全点 | 方案 |
|--------|------|
| 认证 | JWT（access token 2h，refresh token 7d），Redis 维护黑名单 |
| 多租户隔离 | 中间件层强制注入 tenant_id，禁止跨租户查询 |
| 敏感数据 | 手机号 SHA-256 哈希存储（去重用），不存明文 |
| OAuth Token | 广告平台 access_token 加密存储 |
| API 限流 | Nginx 层限制每 IP 每分钟请求数 |
| 文件上传 | 限制类型（xlsx/csv）、大小（最大 10MB） |
| HTTPS | 全站强制 HTTPS，小程序必须 |
| 操作日志 | 所有数据变更记录 audit_logs |
| SQL 注入 | SQLAlchemy ORM 参数化查询，禁止裸 SQL |

---

## 十一、Sprint 计划（v2.0 全面更新）

### Sprint 1（第1-2周）：跑通核心闭环 + 首次体验

> **UX架构核心**：Sprint 1 必须解决 BP-1/BP-2/BP-5 三个致命断点，让新用户注册后5分钟内看到价值

| 任务 | 模块 | 技术实现 | 工期 | UX断点修复 |
|------|------|----------|------|-----------|
| A1 后端认证体系 | Auth | FastAPI JWT，邮件验证 | 6h | — |
| A2 前端登录/注册 | Auth | React + Ant Design 登录页 | 4h | — |
| **B1 演示数据系统** | 引导 | Demo API + 前端切换组件 | 4h | **BP-1** 注册→空仪表盘 |
| **B2 3步引导向导** | 引导 | OnboardingWizard + 状态持久化 | 5h | **BP-2** 引导不连价值链 |
| **B3 仪表盘分阶段空态** | 引导 | EmptyState + 场景判定API | 3h | **BP-5** 仪表盘空态 |
| C1 后端Excel上传+异步 | CRM | 文件上传 + Celery + pandas解析 | 6h | — |
| **C2 智能字段映射** | CRM | 模糊匹配算法 + 前端确认 | 5h | **BP-4** 归因黑洞（部分） |
| D1 巨量引擎OAuth | 广告 | httpx + OAuth + 立即拉取 | 8h | — |
| E1 ROI仪表盘 | 分析 | React + ECharts + 聚合API | 10h | — |
| G1 新导航架构 | 导航 | 分组侧边栏 + 进度常驻 | 3h | — |

**Sprint 1 目标**：Web 端可注册→看到演示ROI→通过3步引导绑定抖音→导入Excel→看到真实ROI图表  
**Sprint 1 总工期**：54小时（约7人天）

### Sprint 2（第3-4周）：完整MVP + 体验修复

| 任务 | 模块 | 技术实现 | 工期 | UX断点修复 |
|------|------|----------|------|-----------|
| D2 百度广告OAuth | 广告 | httpx + 百度OAuth | 6h | — |
| D3 广告数据每日拉取 | 广告 | Celery Beat + 定时任务 | 8h | — |
| C3 导入进度+即时反馈 | CRM | ProgressOverlay + 轮询hook | 3h | — |
| C4 CRM线索管理页 | CRM | 列表+筛选+状态更新 | 4h | — |
| **E2 归因规则自动确认** | 分析 | 自动分析 + AttributionConfirm弹窗 | 5h | **BP-4** 归因黑洞 |
| **E3 交叉分析+上下文跳转** | 分析 | 图表点击事件 + 上下文携带 | 8h | **BP-6** 交叉分析可达性 |
| G2 广告账户管理页 | 设置 | 账户列表+同步状态 | 3h | — |
| G3 归因规则设置页 | 设置 | 规则CRUD | 3h | — |
| H1 小程序仪表盘 | 小程序 | Taro + ec-canvas | 8h | — |

**Sprint 2 目标**：产品可邀请3-5家教育机构试用，双渠道广告数据+CRM+ROI完整闭环  
**Sprint 2 总工期**：48小时（约6人天）

### Sprint 3（第5-6周）：报表+行动建议+打磨+验证

| 任务 | 模块 | 技术实现 | 工期 | UX断点修复 |
|------|------|----------|------|-----------|
| **F1 月报+行动建议** | 报表 | InsightEngine + 规则建议卡片 | 8h | **BP-7** 报表无行动 |
| G4 订阅/计费页 | 设置 | 套餐展示 + 升级意向 | 2h | — |
| 产品打磨 & Bug修复 | — | — | — | — |
| 邀请10家教育机构内测 | — | — | — | — |

**Sprint 3 目标**：至少3家愿意付费，或收到明确付费意向  
**Sprint 3 总工期**：10小时 + 产品验证

---

## 十二、新增前端组件清单（v2.0 完整版）

| 组件 | 文件路径 | 用途 | 复杂度 | 对应UX |
|------|---------|------|--------|--------|
| `OnboardingWizard` | `components/OnboardingWizard/` | 3步引导向导弹窗 | 中 | BP-2 |
| `DemoDataToggle` | `components/DemoDataToggle/` | 演示/真实数据切换开关 | 低 | BP-1 |
| `SetupProgressCard` | `components/SetupProgressCard/` | 仪表盘底部固定进度卡片 | 低 | BP-2 |
| `SmartFieldMapper` | `components/SmartFieldMapper/` | 智能字段映射（模糊匹配+确认） | 高 | BP-4 |
| `AttributionConfirm` | `components/AttributionConfirm/` | 归因规则确认弹窗 | 中 | BP-4 |
| `ProgressOverlay` | `components/ProgressOverlay/` | 导入进度（可最小化） | 中 | — |
| `InsightCard` | `components/InsightCard/` | 行动建议卡片 | 中 | BP-7 |
| `ContextLink` | `components/ContextLink/` | 上下文跳转链接组件 | 低 | BP-6 |
| `EmptyState` | `components/EmptyState/` | 统一空态组件（含行动引导） | 低 | BP-5 |

---

## 十三、质量保障

### 13.1 测试策略

```
tests/
├── unit/
│   ├── test_roi_calculator.py    # ROI 计算逻辑单元测试
│   ├── test_attribution.py       # 归因规则匹配测试
│   ├── test_crm_importer.py     # Excel 解析测试
│   └── test_field_matcher.py     # v2.0新增：字段匹配测试
└── integration/
    ├── test_auth_api.py          # 认证 API 集成测试
    ├── test_demo_api.py          # v2.0新增：演示数据API测试
    ├── test_onboarding_api.py    # v2.0新增：引导状态API测试
    ├── test_crm_api.py           # CRM 上传/查询 API 测试
    └── test_analytics_api.py     # ROI 仪表盘 API 测试
```

### 13.2 非功能要求

| 指标 | 目标 | 实现方式 |
|------|------|----------|
| Excel 导入（5万行）| ≤ 10秒 | Celery 异步 + pandas 批量写入 |
| 仪表盘首屏响应 | ≤ 1秒 | Redis 缓存聚合结果（5分钟 TTL） |
| 广告数据同步失败 | 必须通知 | Celery 任务失败回调 → 邮件告警 + 仪表盘告警条 |
| 数据操作审计 | 全覆盖 | FastAPI 中间件自动记录 audit_logs |
| 演示数据响应 | < 100ms | Redis 缓存演示数据，秒级切换 |
| 移动响应式 | Web + 小程序 | Ant Design 响应式栅格 + Taro 适配 |

---

## 十四、UX断点修复技术映射（v2.0 新增）

| 断点 | 严重性 | 技术实现 | 对应任务 |
|------|--------|----------|----------|
| BP-1 注册→空仪表盘 | 🔴致命 | Demo API + DemoDataToggle组件 + 演示数据标识 | B1 |
| BP-2 引导不连价值链 | 🔴致命 | OnboardingWizard + SetupProgressCard + 引导状态持久化 | B2 |
| BP-3 绑定后无数据 | 🔴严重 | OAuth回调时立即触发sync_ad_data.delay(account_id, days=7) | D1 |
| BP-4 归因黑洞 | 🔴严重 | SmartFieldMapper + AttributionEngine自动分析 + AttributionConfirm弹窗 | C2 + E2 |
| BP-5 仪表盘空态 | 🔴致命 | EmptyState组件 + dashboard/status API场景判定 | B3 + E1 |
| BP-6 交叉分析可达性 | 🟡中等 | 图表click事件 + URL参数携带上下文 | E3 |
| BP-7 报表无行动 | 🟡中等 | InsightEngine规则引擎 + InsightCard组件 | F1 |

---

## 十五、技术债务与演进路径

### MVP 阶段有意接受的技术债

| 债务 | 说明 | 解决时机 |
|------|------|----------|
| 共享数据库多租户 | 行级隔离，未物理分库 | 用户 > 100 家时考虑 Schema 隔离 |
| 无点击归因 | 基于渠道字段软匹配，非精准归因 | Phase 2 加 UTM 参数追踪 |
| 小程序功能阉割 | 主要只做查看，上传在 Web | Phase 2 补全小程序上传能力 |
| 无监控告警 | 仅日志 + 邮件通知 | 接入 Sentry / Prometheus |
| React Native App | MVP 阶段只有 Web + 小程序 | Phase 2 按需开发 |
| 演示数据通用版 | 不根据用户行业个性化 | Phase 2 按行业子类定制 |
| 行动建议规则型 | 非AI驱动，3-5条模板 | Phase 2 接入AI分析 |

### Phase 2 演进方向

```
MVP（Modular Monolith）
    │
    ▼ 当租户 > 100，流量出现瓶颈
按需拆分高频模块：
    ├── Analytics Service（独立服务，读密集）
    ├── Ad Sync Service（独立 Worker）
    └── 主应用保留 Auth + CRM + 设置
```

---

## 十六、关键风险与应对（v2.0 更新）

| 风险 | 概率 | 影响 | 应对 |
|------|------|------|------|
| 巨量引擎/百度 API 审批延迟 | 高 | 阻塞广告模块 | **第一天提交申请**；审批期间先用 Mock 数据开发仪表盘 + 演示数据 |
| Excel 格式千变万化 | 中 | 字段映射复杂 | 智能匹配算法 + 置信度展示 + 用户确认 |
| API未审批导致OAuth失败 | 高 | 绑定流程断点 | 清晰的错误提示 + "稍后重试"选项 + 演示数据逃生出口 |
| 小程序微信审核 | 低 | Sprint 2 延期 | 提前准备资质材料，避免功能违规 |
| 多租户数据泄露 | 低 | 严重 | 中间件强制隔离 + 集成测试验证 |

---

## 附录：依赖清单

### Python 核心依赖（requirements.txt）

```txt
fastapi==0.111.0
uvicorn[standard]==0.29.0
gunicorn==22.0.0
sqlalchemy==2.0.30
alembic==1.13.1
aiomysql==0.2.0
pydantic-settings==2.2.1
python-jose[cryptography]==3.3.0
passlib[bcrypt]==1.7.4
celery[redis]==5.4.0
redis==5.0.4
httpx==0.27.0
pandas==2.2.2
openpyxl==3.1.2
fastapi-mail==1.4.1
structlog==24.1.0
weasyprint==62.3
pytest==8.2.0
pytest-asyncio==0.23.7
```

### 前端核心依赖（package.json 摘要）

```json
{
  "dependencies": {
    "react": "^18.3.0",
    "antd": "^5.17.0",
    "@ant-design/pro-components": "^2.7.0",
    "echarts": "^5.5.0",
    "echarts-for-react": "^3.0.2",
    "zustand": "^4.5.2",
    "axios": "^1.7.2",
    "@tanstack/react-query": "^5.40.0",
    "tailwindcss": "^3.4.4"
  },
  "devDependencies": {
    "vite": "^5.2.0",
    "@vitejs/plugin-react": "^4.3.0",
    "typescript": "^5.4.5"
  }
}
```

---

## 十七、文档变更记录

| 版本 | 日期 | 变更内容 |
|------|------|----------|
| v1.0 | 2026-04-27 | 初始版本，基于产品需求文档 |
| v2.0 | 2026-04-28 | 整合MVP v2.0任务列表、UX架构v1.0断点修复需求 |
| v2.1 | 2026-04-28 | 新增第六A章：广告平台 OAuth 回调传值机制完整实现；新增 Redis 工具层、百度营销服务封装；修复 state→tenant_id 还原缺失、百度回调未实现、同步状态无跟踪等问题 |

---

*技术方案文档版本：v2.0 | 生成日期：2026-04-28 | 修订日期：2026-04-28 | 基于 MVP任务列表v2.0 + UX架构v1.0*
