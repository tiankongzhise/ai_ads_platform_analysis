# 腾讯广告 Marketing API 调研

## 1. 平台定位

腾讯广告 Marketing API 用于接入腾讯广告账户、推广计划、广告组、创意、报表和资产管理等能力。本项目平台代码使用 `tencent`。

开放平台入口：

- 开发者中心：<https://developers.e.qq.com/>
- OAuth 授权文档：<https://developers.e.qq.com/docs/apilist/auth/oauth2>
- 官方 Go SDK：<https://github.com/tencentad/marketing-api-go-sdk>
- Go SDK 包文档：<https://pkg.go.dev/github.com/tencentad/marketing-api-go-sdk/pkg/ads>

## 2. SDK 结论

| 项 | 结论 |
| --- | --- |
| 官方 Go SDK | 可用，仓库为 `tencentad/marketing-api-go-sdk`，Go 包显示 Apache-2.0 许可证。 |
| 官方 Python SDK | 本轮未确认公开官方 Python SDK。 |
| Java SDK | Maven 中存在 `com.tencent.ads:marketing-api-java-sdk`，可作为补充参考。 |
| 本项目建议 | Go 侧优先官方 SDK；Python 侧先 REST 封装，或通过内部 Go 适配器代理请求。 |

Go SDK 的 `SDKClient` 暴露 `Oauth()`、`DailyReports()`、`HourlyReports()`、`Report()` 等服务，适合快速验证授权和日报接口。

## 3. OAuth2 认证

### 3.1 授权入口

腾讯广告要求广告账号对开发者应用授权。用户授权后，回调地址获得 `authorization_code`，后端再换取 token。

### 3.2 换取 token

接口：

```text
GET https://api.e.qq.com/oauth/token
```

授权码模式关键参数：

```text
grant_type=authorization_code
client_id=<app_id>
client_secret=<app_secret>
authorization_code=<authorization_code>
redirect_uri=<redirect_uri>
```

成功响应包含 `access_token`、`refresh_token`、`access_token_expires_in`、`refresh_token_expires_in` 和授权账号信息。官方示例中 access token 有效期为 86400 秒，refresh token 有效期为 2592000 秒。

### 3.3 刷新 token

同一 token 接口使用刷新模式：

```text
grant_type=refresh_token
refresh_token=<refresh_token>
```

公开文档说明刷新时返回新的 `access_token` 和新的过期时间，但不返回新的 `refresh_token`。因此本系统刷新成功后必须：

- 替换 access token 密文。
- 更新 `access_token_expires_at`。
- 保留原 refresh token 密文和过期时间。

## 4. 报表能力

腾讯广告 Go SDK 中可见：

- `DailyReports()`：日报。
- `HourlyReports()`：小时报。
- `Report()`：其他报表服务。
- `RealtimeCost()`：实时消耗。
- `DailyBalanceReport()`：日资金报表。

日报接口支持多个层级，公开资料中可见账户、推广计划、项目、广告组、广告、素材、关键词等层级。分页参数通常包含 `page` 和 `page_size`，公开资料显示 `page_size` 最大可到 1000，日期查询范围可覆盖近 365 天。

## 5. 本项目报表类型建议

| 本项目 `report_type` | 腾讯广告层级 | 粒度 | 用途 |
| --- | --- | --- | --- |
| `account_daily` | advertiser/account | 日 | 广告主总消耗和整体 ROI。 |
| `campaign_daily` | campaign | 日 | 推广计划效果。 |
| `adgroup_daily` | adgroup | 日 | 广告组效果。 |
| `ad_daily` | ad | 日 | 广告/创意维度分析。 |
| `account_hourly` | advertiser/account | 小时 | 小时趋势看板。 |

首批可以只接 `DailyReports` 和 `HourlyReports`，资金、资产和人群接口不进入本轮范围。

## 6. 限制和冲突点

- 每个广告账户都需要授权，access token 对特定账号生效。
- 请求必须同时满足 token 有效、接口配额未用完、接口调用频次未超限。
- refresh token 过期后需要重新走 OAuth2 授权。
- 不同报表层级字段不同，不能将账户、计划、广告组和素材混成同一 `report_type`。
- 当天数据可能延迟，建议近 7 天滚动重拉。
- 大账户分页量大时应按日期和层级拆分任务。

## 7. 适配器建议

```text
TencentAdapter
  - platform_code(): tencent
  - auth: OAuth token 接口，refresh 时保留原 refresh_token
  - fetch: 官方 Go SDK 或 REST
  - report: DailyReports/HourlyReports
  - paging: page + page_size
```

原始记录入库建议额外保存：

- `account_id` 或 `advertiser_id`
- `campaign_id`
- `adgroup_id`
- `ad_id`
- `date`
- `page`
- `request_id`

## 8. 开发前待确认

- 当前应用是否已开通 Marketing API 正式权限。
- 是否需要代理商多账户管理能力。
- 日报接口在当前账号下支持的层级和字段清单。
- 接口频控是否按 app、账号、接口或 token 维度计算。
- 小时级数据是否满足招生看板实时性要求。

## 9. 来源

- 腾讯广告开发者中心：<https://developers.e.qq.com/>
- OAuth 授权文档：<https://developers.e.qq.com/docs/apilist/auth/oauth2>
- 官方 Go SDK：<https://github.com/tencentad/marketing-api-go-sdk>
- Go SDK 包文档：<https://pkg.go.dev/github.com/tencentad/marketing-api-go-sdk/pkg/ads>
- Go SDK v3 包文档：<https://pkg.go.dev/github.com/tencentad/marketing-api-go-sdk/pkg/ads/v3>
- 腾讯广告 Java SDK Maven：<https://repo1.maven.org/maven2/com/tencent/ads/marketing-api-java-sdk/>
- 日报接口 Apifox 镜像：<https://7j0rt64lfd.apifox.cn/api-122863231>

