# 抖音/巨量引擎开放平台调研

## 1. 平台定位

抖音广告数据接入建议按巨量引擎开放平台 Marketing API 建模。本项目平台代码统一使用 `douyin`，对外展示可使用“抖音/巨量引擎”。

开放平台入口：

- 开放平台首页：<https://open.oceanengine.com/>
- 入门指南：<https://open.oceanengine.com/labels/34>
- 接口列表：<https://open.oceanengine.com/labels/7>
- 官方 Go SDK：<https://github.com/oceanengine/ad_open_sdk_go>

## 2. SDK 结论

| 项 | 结论 |
| --- | --- |
| 官方 Go SDK | 可用，仓库为 `oceanengine/ad_open_sdk_go`，许可证为 Apache-2.0。 |
| 官方 Python SDK | 本轮未确认公开官方 Python SDK。 |
| Java SDK | 官方生态存在 Java SDK，可作为字段和 OAuth 参考。 |
| 本项目建议 | Go 控制面和后续 Go 适配器可优先参考官方 Go SDK；Python `ad-sync-service` 可先用 REST 客户端封装。 |

官方 Go SDK 提供：

- token 获取接口封装。
- 请求封装和响应解释。
- 通用 `CommonApi`，可调用未生成代码的接口。
- 大量 OpenAPI 生成接口和模型。

## 3. OAuth2 认证

### 3.1 授权入口

应用需要先注册成为巨量引擎开发者，并申请对应 API 权限组。用户在授权页面完成广告主授权后，平台回调返回 `auth_code`。

### 3.2 换取 token

授权码换 token 接口：

```text
POST https://ad.oceanengine.com/open_api/oauth2/access_token/
```

官方 Go SDK 中对应接口：

```text
Oauth2AccessTokenApi -> OpenApiOauth2AccessTokenPost
```

换取成功后获得 `access_token`、`refresh_token`、授权广告主信息和有效期。入门指南说明 access token 有效期为 24 小时，refresh token 有效期为 30 天。

### 3.3 刷新 token

刷新接口：

```text
POST https://ad.oceanengine.com/open_api/oauth2/refresh_token/
```

官方 Go SDK 中对应接口：

```text
Oauth2RefreshTokenApi -> OpenApiOauth2RefreshTokenPost
```

刷新成功会返回新的 access token 和新的 refresh token。业务实现必须原子替换 token 密文和过期时间，避免旧 refresh token 覆盖新值。

### 3.4 广告主列表

授权完成后需要拉取授权广告主：

```text
GET /open_api/oauth2/advertiser/get/
```

官方 Go SDK 中对应接口：

```text
Oauth2AdvertiserGetApi -> OpenApiOauth2AdvertiserGetGet
```

本系统应使用返回广告主 ID 创建或更新 `ad_platform_accounts.external_account_id`。

## 4. 报表能力

巨量引擎开放平台的报表能力包括常规数据报表、自定义报表、商品报表异步任务、代理商相关报表和 RDS/SPI 等补充数据通道。官方 Go SDK 的接口列表中可见以下代表性接口：

| 能力 | SDK 接口示例 | 说明 |
| --- | --- | --- |
| 广告报表 | `ReportAdGetV2Api` | 常规广告层级报表。 |
| 商品日报异步创建 | `ReportProductDailyAsyncTaskCreateV30Api` | 创建商品日报异步任务。 |
| 商品小时报异步创建 | `ReportProductHourlyAsyncTaskCreateV30Api` | 创建商品小时报异步任务。 |
| 异步任务查询 | `ReportProductAsyncTaskGetV30Api` | 查询异步任务状态。 |
| 异步结果下载 | `ReportProductAsyncTaskDownloadV30Api` | 下载异步任务结果。 |

## 5. 本项目报表类型建议

首批只接与教育招生 ROI 相关的核心报表：

| 本项目 `report_type` | 平台层级 | 粒度 | 用途 |
| --- | --- | --- | --- |
| `account_daily` | 广告主/账户 | 日 | 总消耗、曝光、点击、转化基线。 |
| `campaign_daily` | 广告计划/项目 | 日 | 渠道 ROI、计划对比。 |
| `adgroup_daily` | 广告组/单元 | 日 | 投放单元效果分析。 |
| `creative_daily` | 创意/素材 | 日 | 后续素材效果和行动建议。 |
| `account_hourly` | 广告主/账户 | 小时 | 小时趋势看板。 |

## 6. 限制和冲突点

- 权限组决定接口可用性，SDK 存在接口不代表当前应用可调用。
- access token 过期快，必须在同步前自动刷新。
- refresh token 刷新后会换新，不能并发刷新。
- 报表接口的维度、指标和过滤条件存在组合限制，需在 `report_capabilities` 中维护。
- 大范围历史回溯应拆分日期或走异步任务。
- RDS/SPI 属于增强数据通道，权限申请和运维复杂度高，不作为首批必需能力。

## 7. 适配器建议

```text
DouyinAdapter
  - platform_code(): douyin
  - auth: 使用 token_ref 读取 access_token，过期前 refresh
  - fetch: 优先 CommonApi 或生成接口
  - report: account_daily/campaign_daily/adgroup_daily/creative_daily/account_hourly
  - async: 商品类或大文件类报表走 async_task
```

原始记录入库时建议额外保存：

- `advertiser_id`
- `campaign_id`
- `adgroup_id`
- `creative_id`
- `stat_time` 或平台返回日期字段
- 平台 `request_id`

## 8. 开发前待确认

- 当前开发者应用是否已开通 Marketing API 和报表权限。
- 授权回调域名、应用 ID、secret 和权限组。
- 教育行业账户实际可用的报表接口、字段和时间窗口。
- 是否需要千川、巨量本地推、线索表单等垂直能力。
- RDS/SPI 是否有商业必要。

## 9. 来源

- 巨量引擎开放平台首页：<https://open.oceanengine.com/>
- 巨量引擎开放平台入门指南：<https://open.oceanengine.com/labels/34>
- 巨量引擎开放平台接口列表：<https://open.oceanengine.com/labels/7>
- 官方 Go SDK：<https://github.com/oceanengine/ad_open_sdk_go>
- Go SDK 包文档：<https://pkg.go.dev/github.com/oceanengine/ad_open_sdk_go/api>

