# 教育行业广告×CRM数据平台 — MVP 开发任务列表

> **项目代号**：EduAdCRM-MVP  
> **文档版本**：v2.0（整合技术架构 + UX架构修订）  
> **修订日期**：2026-04-28  
> **修订说明**：基于 `edu-adcrm-tech-architecture.md` 和 `edu-adcrm-ux-architecture.md` 全面重写，解决旧版技术栈冲突、UX断点缺失、任务粒度不合理等问题  
> **目标**：最短时间验证"广告投放 + CRM数据闭环"在教育行业的商业可行性

---

## 一、产品定位（来自研报 + UX架构增强）

### 原始需求引用
> "面向中小企业的'广告投放+CRM数据闭环+行业化落地方案'→ 低门槛 / 轻量接入 / 按需付费 / 自带行业模板"  
> "让客户30天内看到广告效果改善"  
> "无需开发，2小时内接入广告平台+CRM"

### UX架构核心原则（新增）
> **Value-First（价值优先）**：用户注册后5分钟内必须看到产品的核心价值——"ROI数据长什么样"。

### MVP 范围决策
| 维度 | MVP 选择 | 排除（后续迭代）|
|------|----------|----------------|
| 渠道 | 抖音（巨量引擎）+ 百度营销 | 腾讯、快手、小红书 |
| 行业 | 教育培训（K12/职业教育/兴趣班）| 电商、医美、本地生活 |
| CRM接入 | Excel/CSV 文件上传（含智能映射）| API直连企微、销售易 |
| 核心价值 | 广告花费→线索→成交ROI可视化 | 自动投放优化、AI素材生成 |
| 交付形式 | Web SaaS（响应式）+ 微信小程序（查看为主）| 移动App（Phase 2）|

**核心卖点（MVP验证话术）**：  
> "上传你的招生学员数据，连接抖音/百度广告账户，30分钟看清每条广告带来多少招生、花了多少钱。"

---

## 二、技术栈（与技术架构对齐）

```
后端：  Python FastAPI + SQLAlchemy 2.x + Alembic
ASGI：  Uvicorn + Gunicorn（多进程）
队列：  Celery 5.x + Redis 7.x（替代数据库队列）
缓存：  Redis 7.x（缓存+队列+Token黑名单三合一）
Excel：  openpyxl + pandas（替代 Laravel-Excel）
认证：  python-jose + passlib（JWT）
HTTP：  httpx（异步，对接广告平台API）
日志：  structlog（结构化）
测试：  pytest + httpx

前端 Web：  React 18 + Ant Design Pro 6 + Vite
状态管理：  Zustand
请求库：  Axios + React Query
样式：  Tailwind CSS + Ant Design
图表：  ECharts + echarts-for-react（替代 ApexCharts）

前端小程序：Taro 4（React语法）+ Vant Weapp
图表小程序：ec-canvas（ECharts小程序版）

数据库：  MySQL 8.0
文件存储：腾讯云 COS / 本地 MinIO
部署：  Docker + docker-compose + 宝塔面板 + Nginx
```

---

## 三、功能模块 & 开发任务

---

### 模块 A：用户认证与多租户基础

#### [ ] A1：后端认证体系（FastAPI JWT）
**描述**：实现邮箱注册、登录、密码重置、Token刷新、登出（Redis黑名单）、微信小程序登录（code换JWT）  
**验收标准**：
- 注册→邮件验证→登录流程无报错
- JWT access token 2h / refresh token 7d
- 登出后 Token 加入 Redis 黑名单
- 微信小程序 `wx.login()` code 换取 JWT
- 多账户数据完全隔离（SQL 查询均带 tenant_id 过滤）

**要创建/编辑的文件**：
- `backend/app/api/v1/auth.py` — 注册/登录/刷新/登出/微信登录
- `backend/app/core/security.py` — JWT 生成/验证
- `backend/app/core/tenant.py` — 多租户中间件
- `backend/app/models/user.py` — User ORM
- `backend/app/schemas/auth.py` — 请求/响应模型
- `backend/alembic/versions/xxx_create_users_tenants.py`

**工期估计**：6小时  

---

#### [ ] A2：前端登录/注册页面
**描述**：实现 Web 端登录、注册、忘记密码页面，与后端 JWT API 对接  
**验收标准**：
- 登录/注册表单验证完整
- 登录成功后跳转仪表盘（演示模式）
- 支持"记住我"（refresh token 持久化）
- 忘记密码→邮件重置流程

**要创建/编辑的文件**：
- `frontend-web/src/pages/login/` — 登录/注册/忘记密码
- `frontend-web/src/api/auth.ts` — 认证 API 封装
- `frontend-web/src/store/authStore.ts` — Zustand 认证状态

**工期估计**：4小时  

---

### 模块 B：演示数据与引导体验（UX核心断点修复）

> **UX架构映射**：修复 BP-1（注册→空仪表盘）和 BP-2（引导不连价值链），这两个断点影响100%新用户

#### [ ] B1：演示数据系统（后端 + 前端）
**描述**：新用户注册后默认进入演示模式，展示"星海教育"虚构机构的完整ROI数据，让用户5分钟内看到产品核心价值  
**验收标准**：
- 后端提供 `/api/v1/demo/dashboard` 接口，返回教育行业典型演示数据
- 演示数据真实感强：CPE ¥30-50, ROI 1:3-6, 花费¥12K/日
- 前端 Header 右上角有"演示数据/真实数据"切换开关
- 演示数据区域加蓝色虚线边框 + "演示"角标
- 用户完成3步引导配置后，自动切换真实数据（如有）

**要创建/编辑的文件**：
- `backend/app/api/v1/demo.py` — 演示数据接口
- `backend/app/services/demo_data.py` — 演示数据生成（教育行业模板）
- `frontend-web/src/components/DemoDataToggle/` — 演示/真实切换组件
- `frontend-web/src/store/demoStore.ts` — 演示模式状态

**工期估计**：4小时  
**UX参考**：edu-adcrm-ux-architecture.md §2.3.1

---

#### [ ] B2：3步引导向导（OnboardingWizard）
**描述**：将"机构信息→广告绑定→CRM导入"合并为一个连贯的3步引导向导，替代原来的分散配置。注册后仪表盘底部固定"3步引导卡片"  
**验收标准**：
- 引导向导组件 `OnboardingWizard` 包含3步：
  - Step 1：机构信息（名称、行业子类、主要招生产品）
  - Step 2：绑定广告账户（嵌入OAuth流程，提供"暂不绑定"选项）
  - Step 3：导入学员数据（嵌入上传流程，提供"下载行业模板"和"暂不上传"选项）
- 每步有进度条指示（①●━━ ②━━ ③━━）
- 仪表盘底部固定 `SetupProgressCard`，显示配置进度，完成后自动消失
- 配置进度持久化（后端记录完成状态）

**要创建/编辑的文件**：
- `frontend-web/src/components/OnboardingWizard/` — 引导向导组件
- `frontend-web/src/components/SetupProgressCard/` — 配置进度卡片
- `frontend-web/src/pages/onboarding/` — 引导页
- `backend/app/api/v1/onboarding.py` — 引导状态 API
- `backend/app/models/onboarding_status.py` — 引导进度记录

**工期估计**：5小时  
**UX参考**：edu-adcrm-ux-architecture.md §2.4.1

---

#### [ ] B3：仪表盘分阶段空态设计
**描述**：替代原来的"暂无数据"死胡同，实现分阶段空态——每完成一步配置，仪表盘就"亮"一部分  
**验收标准**：
- 场景A（未绑定广告+未导入CRM）：显示行动引导卡片 + "继续配置" / "查看演示数据" 按钮
- 场景B（已绑定广告，数据同步中）：花费卡片显示"同步中🔄" + 其他卡片暗灰 + "数据同步后效果预览"入口
- 场景C（有广告数据，CRM未导入）：花费卡片有数据 + CRM相关卡片暗灰 + "导入学员数据"行动按钮
- 所有空态有"查看演示数据"逃生出口

**要创建/编辑的文件**：
- `frontend-web/src/components/EmptyState/` — 统一空态组件（含行动引导）
- `frontend-web/src/pages/dashboard/` — 仪表盘空态逻辑

**工期估计**：3小时  
**UX参考**：edu-adcrm-ux-architecture.md §2.5.1

---

### 模块 C：CRM数据导入

#### [ ] C1：后端 Excel 上传 + 异步导入
**描述**：实现 Excel/CSV 文件上传 API，触发 Celery 异步任务进行解析、去重、批量写入  
**验收标准**：
- 上传接口 `POST /api/v1/crm/upload`，文件存入 COS/MinIO
- Celery Worker 异步解析（openpyxl/pandas）
- 手机号 SHA-256 哈希去重
- 批量写入 DB（每500行一批）
- 进度查询接口 `GET /api/v1/crm/upload/{task_id}`（Redis存储进度）
- 字段映射确认接口 `POST /api/v1/crm/upload/{batch_id}/confirm`
- 支持5万行文件处理 ≤ 10秒

**要创建/编辑的文件**：
- `backend/app/api/v1/crm.py` — CRM API
- `backend/app/services/crm_importer.py` — Excel解析+字段映射
- `backend/app/tasks/import_tasks.py` — Celery异步导入任务
- `backend/app/models/crm_lead.py` — CRM Lead ORM
- `backend/app/schemas/crm.py` — 请求/响应模型
- `backend/alembic/versions/xxx_create_crm_tables.py`

**工期估计**：6小时  

---

#### [ ] C2：智能字段映射（前端 SmartFieldMapper）
**描述**：替代纯手动映射，实现自动模糊匹配列名 + 置信度展示 + 数据预览确认  
**验收标准**：
- 自动匹配常见列名变体（"姓名"/"name"/"名字"/"学员姓名"→标准字段"姓名"）
- 匹配结果标注置信度：✅自动匹配 / ⚠️建议匹配（置信度80%）/ ❌未匹配
- 数据预览展示前5行（手机号脱敏显示 138****1234）
- 必填字段（姓名、手机号）标 * 强调
- 来源渠道字段用💡强调其ROI分析关键性
- 非必填字段默认"跳过此列"

**要创建/编辑的文件**：
- `frontend-web/src/components/SmartFieldMapper/` — 智能映射组件
- `frontend-web/src/utils/fieldMatcher.ts` — 模糊匹配算法

**工期估计**：5小时  
**UX参考**：edu-adcrm-ux-architecture.md §2.4.2

---

#### [ ] C3：CRM导入进度 + 即时反馈
**描述**：实现导入进度条（可最小化），完成后自动跳转仪表盘  
**验收标准**：
- 进度条 + 预计剩余时间
- 实时显示：已处理/成功/跳过（重复）/失败 数量
- 可最小化，不阻塞用户浏览其他页面
- 导入完成后自动跳转仪表盘
- 跳过/失败数据提供"下载错误报告"

**要创建/编辑的文件**：
- `frontend-web/src/components/ProgressOverlay/` — 导入进度组件
- `frontend-web/src/hooks/useImportProgress.ts` — 轮询进度 hook

**工期估计**：3小时  
**UX参考**：edu-adcrm-ux-architecture.md §2.4.3

---

#### [ ] C4：CRM线索管理页
**描述**：展示已导入线索列表，支持搜索、筛选、状态更新  
**验收标准**：
- 分页显示，每页20条
- 按来源渠道、成交状态、时间范围筛选
- 手动标注线索成交状态（已成交/未成交/跟进中）
- 删除整批导入记录
- 行动：标记"来源渠道"缺失的线索，提示用户补全

**要创建/编辑的文件**：
- `frontend-web/src/pages/crm/` — 线索列表页
- `backend/app/api/v1/crm.py` — 线索 CRUD 接口

**工期估计**：4小时  

---

### 模块 D：广告账户接入

#### [ ] D1：巨量引擎（抖音）广告账户 OAuth 授权
**描述**：通过 OAuth 2.0 引导用户授权抖音广告账户，**绑定后立即拉取7天历史数据**  
**验收标准**：
- OAuth授权流程完整（获取授权URL→跳转→回调→存储token）
- 授权成功后显示账户名称、余额、账户状态
- Token过期自动刷新
- **绑定后立即触发 Celery 任务拉取最近7天历史数据**（UX关键改进）
- 支持绑定多个广告账户
- OAuth跳转前明确告知"会跳转到巨量引擎授权页面"

**要创建/编辑的文件**：
- `backend/app/services/juliang_service.py` — 巨量引擎API封装
- `backend/app/api/v1/ad_accounts.py` — 广告账户API
- `backend/app/tasks/sync_ad_tasks.py` — 数据同步任务
- `backend/app/models/ad_account.py` — AdAccount ORM
- `backend/alembic/versions/xxx_create_ad_tables.py`

**外部依赖**：申请巨量引擎营销API开发者账号（open.oceanengine.com），**第一天提交申请**  
**工期估计**：8小时  
**UX参考**：edu-adcrm-ux-architecture.md §2.4.1 步骤2/3

---

#### [ ] D2：百度营销广告账户 OAuth 授权
**描述**：通过百度营销API OAuth授权接入，拉取广告账户信息  
**验收标准**：
- OAuth授权流程完整
- 账户信息正常显示
- 与巨量引擎账户在同一列表页管理
- 绑定后立即拉取7天历史数据
- 支持账户断开重连

**要创建/编辑的文件**：
- `backend/app/services/baidu_service.py` — 百度营销API封装
- `.env` 配置百度 API Key/Secret

**外部依赖**：申请百度营销API开发者账号（mssp.baidu.com），**第一天提交申请**  
**工期估计**：6小时  

---

#### [ ] D3：广告数据每日定时拉取
**描述**：Celery Beat 每日凌晨自动拉取昨日广告数据：花费/展示/点击/表单提交量  
**验收标准**：
- Celery Beat 定时任务每日02:00（巨量）/ 02:30（百度）执行
- 拉取字段：日期/广告组/广告计划/花费/点击数/展示数/表单提交数
- 失败时发邮件通知用户 + Celery任务失败回调
- 支持手动触发"立即同步" `POST /api/v1/ad/accounts/{id}/sync`
- 同步状态在仪表盘可见（通知铃铛红点）

**要创建/编辑的文件**：
- `backend/app/tasks/sync_ad_tasks.py` — 定时同步任务
- `backend/app/tasks/celery_app.py` — Celery Beat 配置
- `backend/app/services/email.py` — 失败告警邮件

**工期估计**：8小时  

---

### 模块 E：ROI 分析核心（MVP最高优先级）

#### [ ] E1：ROI 仪表盘（含演示模式 + 分阶段空态）
**描述**：将广告花费数据与CRM成交数据按"渠道来源"关联，生成核心ROI仪表盘。整合演示数据模式和分阶段空态设计  
**核心指标（教育行业定制）**：
- 今日花费 / 累计线索 / 线索成本(CPE) / ROI比值 / 转化率
- 近N天花费 vs 新增线索趋势折线图
- 抖音 vs 百度 渠道对比柱状图
- 广告计划排行表（可点击跳转交叉分析）

**验收标准**：
- 5个核心指标卡片，支持演示/真实数据切换
- ECharts 折线图展示近30天花费 vs 线索趋势
- ECharts 柱状图展示渠道对比
- 分阶段空态（场景A/B/C）正确展示
- 日期选择默认"近7天"（替代"今日"，因广告数据T+1）
- 指标卡片可点击→跳转交叉分析（携带上下文）
- Redis 缓存聚合结果（5分钟TTL），首屏响应 ≤ 1秒

**要创建/编辑的文件**：
- `frontend-web/src/pages/dashboard/` — ROI 仪表盘
- `frontend-web/src/components/RoiCard/` — 指标卡片
- `frontend-web/src/components/TrendChart/` — 趋势折线图
- `frontend-web/src/components/ChannelBar/` — 渠道对比柱状图
- `backend/app/api/v1/analytics.py` — 仪表盘数据接口
- `backend/app/services/roi_calculator.py` — ROI计算逻辑

**工期估计**：10小时  
**UX参考**：edu-adcrm-ux-architecture.md §2.3.1 + §2.5.1

---

#### [ ] E2：归因规则自动确认（替代设置页手动配置）
**描述**：CRM导入完成后自动分析来源字段值的分布，生成归因匹配建议，用户确认即可（替代隐藏在设置页的手动配置）  
**验收标准**：
- 导入完成后自动弹出归因确认弹窗 `AttributionConfirm`
- 自动分析来源渠道值分布（"抖音"2340条、"百度"1890条等）
- 自动匹配到广告平台（巨量引擎/百度营销）
- 未匹配数据明确提示"将归入其他渠道，不计入ROI"
- 保留手动调整入口（设置页）
- 关联结果实时更新仪表盘

**要创建/编辑的文件**：
- `backend/app/services/attribution.py` — 归因引擎（自动分析+匹配）
- `frontend-web/src/components/AttributionConfirm/` — 归因确认弹窗
- `backend/app/models/attribution_rules.py` — 归因规则表

**工期估计**：5小时  
**UX参考**：edu-adcrm-ux-architecture.md §2.5.2

---

#### [ ] E3：广告计划×课程 交叉分析（含上下文跳转）
**描述**：展示广告计划/广告组与招生课程的交叉分析，**支持从仪表盘直接跳转（携带筛选上下文）**  
**验收标准**：
- 可选择时间范围（7天/30天/自定义）
- 表格：广告计划名/花费/表单提交/CRM匹配线索数/成交数/成交金额/ROI
- 支持按ROI排序
- 导出CSV
- **从仪表盘跳转时自动携带筛选上下文**（时间范围+平台+课程）
- 顶部面包屑显示来源：「仪表盘 > 抖音渠道 > 近30天」
- ECharts 图表元素可点击（click 事件→跳转交叉分析）

**要创建/编辑的文件**：
- `frontend-web/src/pages/analytics/` — 交叉分析页
- `frontend-web/src/components/ContextLink/` — 上下文跳转组件
- `backend/app/api/v1/analytics.py` — 交叉分析接口

**工期估计**：8小时  
**UX参考**：edu-adcrm-ux-architecture.md §2.6.1

---

### 模块 F：报表与导出

#### [ ] F1：招生ROI月报 + 行动建议
**描述**：月报自动生成，支持Excel/PDF下载；**增加规则型行动建议**，让报表从"存档文件"变成"决策驱动器"  
**验收标准**：
- 报表包含：月度总结数据 + 渠道对比 + 课程维度分析
- 支持下载Excel版本
- Celery Beat 每月1日08:00自动生成 + 邮件发送
- **行动建议卡片 `InsightCard`**（基于规则，非AI）：
  - ROI对比→建议预算倾斜（如"抖音ROI高于百度62%，建议将百度20%预算转至抖音"）
  - 线索成本异常→建议优化落地页
  - 花费上升但线索持平→建议暂停低效广告组
- 3-5条模板化建议，覆盖最常见优化场景

**要创建/编辑的文件**：
- `backend/app/tasks/report_tasks.py` — 月报生成Celery任务
- `backend/app/services/report_generator.py` — 报表生成（openpyxl/WeasyPrint）
- `backend/app/services/insight_engine.py` — 规则型行动建议引擎
- `backend/app/api/v1/reports.py` — 报表API
- `frontend-web/src/components/InsightCard/` — 行动建议卡片
- `frontend-web/src/pages/reports/` — 报表页

**工期估计**：8小时  
**UX参考**：edu-adcrm-ux-architecture.md §2.6.2

---

### 模块 G：导航与系统设置

#### [ ] G1：新导航架构
**描述**：实现分组+优先级的侧边栏导航，替代原来的平铺菜单  
**验收标准**：
- 导航分组：概览 / 分析 / 数据 / 输出 / 配置
- 用户语言替换："交叉分析"→"课程×广告分析"，"CRM线索管理"→"学员线索"
- 仪表盘标注⭐核心
- 底部常驻"配置进度"条（3步引导未完成时）
- 响应式适配（移动端折叠为汉堡菜单）

**要创建/编辑的文件**：
- `frontend-web/src/layouts/MainLayout/` — 主布局+侧边栏
- `frontend-web/src/components/SidebarNav/` — 导航组件

**工期估计**：3小时  
**UX参考**：edu-adcrm-ux-architecture.md §3.2

---

#### [ ] G2：广告账户管理页
**描述**：统一展示已绑定的抖音/百度广告账户，支持添加/删除/刷新授权  
**验收标准**：
- 按平台分组展示
- 显示账户名、余额、最后同步时间
- 同步时间解释文案："数据每日凌晨自动同步"
- 手动触发同步按钮
- 断开连接需二次确认

**要创建/编辑的文件**：
- `frontend-web/src/pages/ad-accounts/` — 广告账户管理页

**工期估计**：3小时  

---

#### [ ] G3：归因规则设置页
**描述**：允许用户查看和调整渠道来源字段的标准化映射规则（归因规则主要在导入时自动确认，此页面作为高级调整入口）  
**验收标准**：
- 展示当前归因规则列表
- 支持添加/编辑/删除规则
- 每条规则显示匹配到的数据量

**要创建/编辑的文件**：
- `frontend-web/src/pages/settings/` — 设置页
- `backend/app/api/v1/settings.py` — 设置API

**工期估计**：3小时  

---

#### [ ] G4：订阅/计费页（MVP验证用）
**描述**：展示当前套餐，收集升级意向  
**验收标准**：
- 明确展示功能限制（免费版最多导入5000条CRM数据，绑定1个广告账户）
- 付费升级按钮跳转微信/支付宝二维码收款页（MVP手动处理）

**工期估计**：2小时  

---

### 模块 H：微信小程序端

#### [ ] H1：小程序仪表盘 + 线索查看
**描述**：Taro 小程序端核心功能，聚焦查看+告警场景  
**验收标准**：
- TabBar：📊仪表盘 / 👥学员 / 📋报表 / 👤我的
- 仪表盘：4个指标卡片（2列布局）+ 简化趋势图（近7天）+ 今日洞察
- 学员列表：只读线索列表 + 搜索筛选
- 微信登录：`wx.login()` → 后端换取JWT
- ec-canvas 图表适配
- 广告数据同步失败→微信服务通知

**要创建/编辑的文件**：
- `frontend-weapp/src/pages/` — 4个TabBar页面
- `frontend-weapp/src/components/EchartsWX/` — ec-canvas封装
- `frontend-weapp/src/components/MetricCard/` — 指标卡片
- `frontend-weapp/src/api/` — 网络请求封装（wx.request）

**工期估计**：8小时  
**UX参考**：edu-adcrm-ux-architecture.md §6.1 + §6.2

---

## 四、开发优先级与冲刺计划

### Sprint 1（第1-2周）：跑通核心闭环 + 首次体验

> **UX架构核心**：Sprint 1 必须解决 BP-1/BP-2/BP-5 三个致命断点，让新用户注册后5分钟内看到价值

| 任务 | 模块 | 负责方向 | 工期 | UX断点修复 |
|------|------|----------|------|-----------|
| A1 后端认证体系 | Auth | 后端 | 6h | — |
| A2 前端登录/注册 | Auth | 前端 | 4h | — |
| B1 演示数据系统 | 引导 | 全栈 | 4h | **BP-1** 注册→空仪表盘 |
| B2 3步引导向导 | 引导 | 前端 | 5h | **BP-2** 引导不连价值链 |
| B3 仪表盘分阶段空态 | 引导 | 前端 | 3h | **BP-5** 仪表盘空态 |
| C1 后端Excel上传+异步 | CRM | 后端 | 6h | — |
| C2 智能字段映射 | CRM | 前端 | 5h | **BP-4** 归因黑洞（部分） |
| D1 巨量引擎OAuth | 广告 | 后端 | 8h | — |
| E1 ROI仪表盘 | 分析 | 全栈 | 10h | — |
| G1 新导航架构 | 导航 | 前端 | 3h | — |

**Sprint 1 目标**：Web 端可注册→看到演示ROI→通过3步引导绑定抖音→导入Excel→看到真实ROI图表  
**Sprint 1 总工期**：54小时（约7人天）

---

### Sprint 2（第3-4周）：完整MVP + 体验修复

| 任务 | 模块 | 负责方向 | 工期 | UX断点修复 |
|------|------|----------|------|-----------|
| D2 百度广告OAuth | 广告 | 后端 | 6h | — |
| D3 广告数据每日拉取 | 广告 | 后端 | 8h | — |
| C3 导入进度+即时反馈 | CRM | 前端 | 3h | — |
| C4 CRM线索管理页 | CRM | 前端 | 4h | — |
| E2 归因规则自动确认 | 分析 | 全栈 | 5h | **BP-4** 归因黑洞 |
| E3 交叉分析+上下文跳转 | 分析 | 全栈 | 8h | **BP-6** 交叉分析可达性 |
| G2 广告账户管理页 | 设置 | 前端 | 3h | — |
| G3 归因规则设置页 | 设置 | 前端 | 3h | — |
| H1 小程序仪表盘 | 小程序 | 全栈 | 8h | — |

**Sprint 2 目标**：产品可邀请3-5家教育机构试用，双渠道广告数据+CRM+ROI完整闭环  
**Sprint 2 总工期**：48小时（约6人天）

---

### Sprint 3（第5-6周）：报表+行动建议+打磨+验证

| 任务 | 模块 | 负责方向 | 工期 | UX断点修复 |
|------|------|----------|------|-----------|
| F1 月报+行动建议 | 报表 | 全栈 | 8h | **BP-7** 报表无行动 |
| G4 订阅/计费页 | 设置 | 前端 | 2h | — |
| 产品打磨 & Bug修复 | — | 全员 | — | — |
| 邀请10家教育机构内测 | — | 产品 | — | — |

**Sprint 3 目标**：至少3家愿意付费，或收到明确付费意向  
**Sprint 3 总工期**：10小时 + 产品验证

---

## 五、UX断点修复追踪

> 完整映射 UX架构识别的7个断点及其修复方案和对应任务

| 断点 | 严重性 | 修复方案 | 对应任务 | Sprint |
|------|--------|----------|----------|--------|
| BP-1 注册→空仪表盘 | 🔴致命 | 演示数据模式 + 引导卡片 | B1 + B3 | Sprint 1 |
| BP-2 引导不连价值链 | 🔴致命 | 3步引导向导 | B2 | Sprint 1 |
| BP-3 绑定后无数据 | 🔴严重 | 绑定后立即拉7天历史 | D1（内含立即拉取） | Sprint 1 |
| BP-4 归因黑洞 | 🔴严重 | 智能映射 + 自动归因确认 | C2 + E2 | Sprint 1-2 |
| BP-5 仪表盘空态 | 🔴致命 | 分阶段空态设计 | B3 + E1 | Sprint 1 |
| BP-6 交叉分析可达性 | 🟡中等 | 上下文跳转 | E3 | Sprint 2 |
| BP-7 报表无行动 | 🟡中等 | 规则型行动建议 | F1 | Sprint 3 |

---

## 六、新增前端组件清单

> 来自 UX架构 §7.2，与技术架构对齐

| 组件 | 用途 | 复杂度 | 对应任务 |
|------|------|--------|----------|
| `OnboardingWizard` | 3步引导向导 | 中 | B2 |
| `DemoDataToggle` | 演示/真实数据切换 | 低 | B1 |
| `SetupProgressCard` | 配置进度卡片（固定在仪表盘） | 低 | B2 |
| `SmartFieldMapper` | 智能字段映射（模糊匹配） | 高 | C2 |
| `AttributionConfirm` | 归因规则确认弹窗 | 中 | E2 |
| `ProgressOverlay` | 导入进度（可最小化） | 中 | C3 |
| `InsightCard` | 行动建议卡片 | 中 | F1 |
| `ContextLink` | 上下文跳转链接组件 | 低 | E3 |
| `EmptyState` | 统一空态组件（含行动引导） | 低 | B3 |

---

## 七、质量要求

- [ ] 所有 ECharts 组件使用 echarts-for-react 封装，图表配置统一管理
- [ ] Ant Design 组件严格按官方文档使用，不使用未公开属性
- [ ] 任何命令中无后台进程 — 永远不要附加 `&`
- [ ] 无服务器启动命令 — 假设开发服务器正在运行
- [ ] 移动响应式设计（Ant Design 响应式栅格 + Tailwind）
- [ ] Excel 上传处理 ≤ 10秒（5万行以内，Celery异步+pandas批量写入）
- [ ] 仪表盘首屏响应 ≤ 1秒（Redis缓存聚合结果，5分钟TTL）
- [ ] 广告数据同步失败有明确错误提示（邮件告警 + 仪表盘告警条）
- [ ] 所有涉及用户数据的操作记录 audit_logs
- [ ] 图片来自 Unsplash 或 https://picsum.photos/（无 Pexels，403错误）
- [ ] 运行 `pytest tests/ -v --cov=app` 确保核心测试通过
- [ ] 微信小程序端适配 ec-canvas 图表，2列指标卡片布局

---

## 八、关键交互模式规范

> 来自 UX架构 §4，开发时必须遵循

### 8.1 数据状态流转
所有数据展示组件必须遵循：加载态(骨架屏) → 空态(引导行动) → 部分数据态(渐进提示) → 完整数据态

### 8.2 操作反馈模式
| 操作类型 | 反馈方式 | 持续时间 |
|----------|----------|----------|
| 即时操作（点击、切换） | 按钮状态变化 | 150ms |
| 短操作（表单提交） | Toast 成功/失败 | 3秒自动消失 |
| 长操作（Excel导入） | 进度条+可最小化 | 操作完成后消失 |
| 异步结果（广告同步） | 通知铃铛+红点 | 用户查看后消失 |
| 危险操作（删除、断开） | 二次确认弹窗 | 用户确认后 |

---

## 九、技术说明

### 关键外部API申请（预计等待时间）
| API | 申请入口 | 预计审批时间 | 优先级 |
|-----|---------|------------|--------|
| 巨量引擎营销API | open.oceanengine.com | 3-7个工作日 | P0（第一天提交） |
| 百度营销API | mssp.baidu.com | 5-10个工作日 | P0（第一天提交） |

> ⚠️ API审批是关键路径！第一天就要提交申请。审批期间用演示数据 + Mock API 推进前端开发。

### 数据关联设计（MVP简化版）
MVP 不做复杂的点击归因，而是基于：
1. CRM中来源渠道字段（自动归因确认，非手动设置页配置）
2. 日期范围聚合

后续版本再加点击追踪参数（UTM）归因。

### 演示数据策略
- 演示机构："星海教育"（虚构）
- 数据真实感：CPE ¥30-50, ROI 1:3-6, 日花费 ¥8K-15K
- 首次注册默认演示模式，完成3步引导后自动切换真实数据
- Phase 2 再考虑个性化演示数据（根据用户填写的行业子类）

---

## 十、MVP 验证指标（商业可行性判定）

### 功能验证指标
| 指标 | 最低通过线 | 理想目标 |
|------|-----------|----------|
| 注册→完成配置率 | ≥40% | ≥60% |
| 注册→首次看到ROI数据 | ≤30分钟 | ≤10分钟 |
| CRM导入完成率 | ≥60% | ≥80% |
| 仪表盘空态停留时长 | ≤60秒 | ≤30秒 |

### 商业验证指标
| 指标 | 最低通过线 | 理想目标 |
|------|-----------|----------|
| 注册并完成配置的机构数 | 5家 | 15家 |
| 导入CRM数据并完成分析的机构数 | 3家 | 10家 |
| NPS（净推荐值）| ≥ 30 | ≥ 50 |
| 付费意愿确认数 | 1家 | 3家 |
| 免费试用到付费转化率 | 20% | 40% |

---

## 十一、与旧版任务列表的主要变更

| 变更项 | 旧版（v1.0） | 新版（v2.0） | 变更原因 |
|--------|-------------|-------------|----------|
| 后端技术栈 | Laravel + PHP | FastAPI + Python | 技术架构已调整 |
| 前端技术栈 | Livewire + FluxUI | React + Ant Design Pro | 技术架构已调整 |
| 图表库 | ApexCharts | ECharts | 技术架构已调整+小程序有官方版 |
| 队列 | Laravel Queue(数据库) | Celery + Redis | 技术架构已调整 |
| 演示数据系统 | 无 | B1 新增 | UX BP-1修复 |
| 3步引导向导 | A2仅基础信息 | B2 完整3步(含广告绑定+CRM) | UX BP-2修复 |
| 仪表盘空态 | "--"死胡同 | B3 分阶段空态+行动引导 | UX BP-5修复 |
| 字段映射 | 纯手动 | C2 智能映射(模糊匹配) | UX BP-4修复 |
| 归因规则 | 设置页手动配置 | E2 导入后自动确认 | UX BP-4修复 |
| 广告数据同步 | 等凌晨定时 | D1 绑定后立即拉7天历史 | UX BP-3修复 |
| 交叉分析 | 独立页面无关联 | E3 上下文跳转 | UX BP-6修复 |
| 报表 | 纯数据报表 | F1 报表+行动建议 | UX BP-7修复 |
| 导航 | 平铺菜单 | G1 分组+优先级 | UX架构重设计 |
| 小程序 | 排除 | H1 Sprint 2纳入 | 技术架构已支持 |
| 认证 | 仅邮箱 | A1 邮箱+微信小程序登录 | 技术架构已设计 |
| Sprint安排 | A2在Sprint2 | 引导向导拆分到Sprint1 | 致命断点必须首 sprint 修复 |

---

## 十二、技术债务与已知限制

| 债务 | 说明 | 解决时机 |
|------|------|----------|
| 共享数据库多租户 | 行级隔离，未物理分库 | 用户>100家时考虑Schema隔离 |
| 无点击归因 | 基于渠道字段软匹配 | Phase 2 加UTM参数追踪 |
| 小程序功能阉割 | 主要只做查看 | Phase 2 补全上传能力 |
| 行动建议为规则型 | 非AI驱动，3-5条模板 | Phase 2 接入AI分析 |
| 演示数据为通用版 | 不根据用户行业个性化 | Phase 2 按行业子类定制 |
| 无监控告警 | 仅日志+邮件 | 接入Sentry/Prometheus |
| React Native App | MVP只有Web+小程序 | Phase 2 按需开发 |

---

## 十三、竞争差异化定位（教育行业话术）

给教育机构负责人/校长的一句话：  
> "把你们的招生台账上传进来，绑定抖音和百度账户，系统自动告诉你：每招一个学员，抖音花了多少钱，百度花了多少钱，哪个广告计划最值得加大预算。"

竞品无法提供的差异点：
- 神策/GrowingIO：太贵、太复杂，需要专业数据团队
- 通用CRM：没有广告费用ROI分析
- 广告平台自带后台：只有广告数据，没有成交/招生数据

我们的差异：教育行业专属模板 + Excel智能导入 + 30分钟看到结果 + 行动建议驱动决策

---

*任务列表版本：v2.0 | 修订日期：2026-04-28 | 基于技术架构v1.0 + UX架构v1.0 全面修订*
