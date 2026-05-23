# OAuth2 授权与 Token 生命周期

## 1. 通用接入流程

广告平台授权接入在本项目中分为两层：

1. 已完成的广告平台 OAuth 回调能力负责生成授权链接、接收平台回调和完成平台侧授权。
2. 本项目的 `callback-adapter-service` 只消费可信回调结果，创建或更新 `ad_platform_accounts`，保存 `token_ref`，再投递 `ad.sync.requested`。

通用状态流：

```text
用户选择平台和组织
  -> 生成平台授权链接，写入 state
  -> 用户在平台侧授权
  -> 平台回调 redirect_uri，返回 code/auth_code/authorization_code 和 state
  -> 回调能力校验 state、防重放和租户上下文
  -> 使用授权码换取 access_token 与 refresh_token
  -> token 加密存储，业务表只保存 token_ref
  -> 创建或更新广告账户引用
  -> 投递 ad.sync.requested
```

## 2. State 与回调校验

`state` 必须由后端生成并保存到 Redis，建议内容包括：

- `tenant_id`
- `organization_id`
- `channel_id`
- `platform`
- `redirect_after_success`
- `nonce`
- `expires_at`

回调处理规则：

- `state` 只能使用一次，验证成功后立即删除。
- 回调中的平台账号必须落在当前租户和组织授权范围内。
- 授权码通常有效期很短，不应进入异步队列后再换 token。
- 回调失败时只记录错误摘要，不记录完整 token、secret 或用户敏感信息。

## 3. Token 存储模型

业务表 `ad_platform_accounts.token_ref` 只保存 token 存储引用，例如：

```json
{
  "token_ref": "kms://ad-platform-token/douyin/tenant_x/account_y"
}
```

加密 token 存储建议保存：

| 字段 | 说明 |
| --- | --- |
| `platform` | 平台代码：`douyin`、`tencent`、`baidu`、`xiaohongshu`。 |
| `tenant_id` | 租户 ID。 |
| `account_id` | 本系统广告账户 ID。 |
| `external_account_id` | 平台广告主或账户 ID。 |
| `access_token_ciphertext` | 加密后的 access token。 |
| `refresh_token_ciphertext` | 加密后的 refresh token；平台不返回时为空。 |
| `access_token_expires_at` | access token 预计过期时间。 |
| `refresh_token_expires_at` | refresh token 预计过期时间。 |
| `scope_list` | 授权权限列表。 |
| `last_rotated_at` | 最近一次换新或刷新时间。 |

## 4. 平台差异摘要

| 平台 | 授权码字段 | 换 token | 刷新 token | 有效期信息 | 本项目处理 |
| --- | --- | --- | --- | --- | --- |
| 抖音/巨量引擎 | `auth_code` | `POST https://ad.oceanengine.com/open_api/oauth2/access_token/` | `POST https://ad.oceanengine.com/open_api/oauth2/refresh_token/` | 官方入门指南说明 access token 24 小时，refresh token 30 天；刷新会返回新的二者。 | 提前 2 小时刷新；刷新成功后原 refresh token 作废，必须原子替换。 |
| 腾讯广告 | `authorization_code` | `GET https://api.e.qq.com/oauth/token`，`grant_type=authorization_code` | 同接口，`grant_type=refresh_token` | 官方 OAuth 文档返回 `access_token_expires_in`、`refresh_token_expires_in`；示例 access token 为 86400 秒，refresh token 为 2592000 秒。 | 刷新接口在 `grant_type=refresh_token` 时不返回新的 refresh token，应保留原 refresh token 和新 access token。 |
| 百度营销 | 依授权模式而定 | 官方 Python SDK 请求头使用 `userName + accessToken`；SDK 还包含 `OAuthAuthorizedToolAPI/getAuthCode` 授权工具。 | 临时 token 需续期或重置；永久授权需开发前在官方文档中确认完整流程。 | 百度资料显示临时授权有效期可为 3 个月；长期接入应走永久授权。 | Python 侧优先使用官方 SDK；token 模型要同时支持短期 access token 与长期授权工具。 |
| 小红书聚光 | `auth_code` | 官方公开页需登录查看；社区 Go SDK 暴露 `AccessToken` 方法。 | 社区 Go SDK 暴露 `RefreshToken` 方法。 | 社区 Go SDK 模型包含 `access_token_expires_in` 与 `refresh_token_expires_in`。 | 生产接入前必须用官方文档复核接口地址、字段和有效期。 |

## 5. Refresh 策略

`ad-sync-service` 调用平台 API 前必须先检查 token 状态：

- 若 access token 距离过期小于平台安全窗口，则先刷新。
- 同一账户同一时间只允许一个刷新任务，使用 Redis 分布式锁避免并发刷新覆盖。
- 刷新成功后原子更新密文、过期时间和 `last_rotated_at`。
- 刷新失败若属于网络或平台临时错误，进入指数退避重试。
- 刷新失败若属于 refresh token 过期、授权撤销或权限不足，将账户标记为 `reauthorization_required`。

建议安全窗口：

| 平台 | access token 刷新提前量 | refresh token 到期提醒 |
| --- | --- | --- |
| 抖音/巨量引擎 | 提前 2 小时 | 到期前 7 天提醒并主动刷新。 |
| 腾讯广告 | 提前 2 小时 | 到期前 7 天提醒；刷新接口未返回新 refresh token 时按原过期时间继续追踪。 |
| 百度营销 | 临时授权到期前 14 天 | 临时授权应提示人工续期；永久授权按官方文档实现自动获取。 |
| 小红书聚光 | 先按提前 2 小时默认值 | 以官方文档确认值覆盖默认策略。 |

## 6. 失败与重授权

以下错误应进入重授权流程：

- 平台返回 token 失效且刷新失败。
- refresh token 过期或被平台撤销。
- 广告主取消授权。
- scope 缺失导致报表接口不可用。
- 平台账户状态异常且无法通过重试恢复。

账户进入重授权状态后：

- 停止自动同步任务。
- 前端广告账户页面展示重新授权入口。
- 保留历史 `ad_raw_reports` 和 ETL 数据，不删除历史报表。
- 新授权成功后复用原 `ad_platform_accounts` 记录，更新 `token_ref` 和 `last_sync_status`。

## 7. 来源

- 巨量引擎开放平台入门指南：<https://open.oceanengine.com/labels/34>
- 巨量引擎开放平台首页：<https://open.oceanengine.com/>
- 腾讯广告 OAuth 授权文档：<https://developers.e.qq.com/docs/apilist/auth/oauth2>
- 百度官方 Python SDK README：<https://github.com/baidu/baiduads-sdk/blob/main/python/README.md>
- 百度 SDK `OAuthAuthorizedToolAPI` 文档路径：<https://github.com/baidu/baiduads-sdk/blob/main/python/baiduads-sdk-auto/docs/OAuthAuthorizedToolAPI.md>
- 小红书商业开放平台：<https://ad-market.xiaohongshu.com/>
- 小红书聚光社区 Go SDK OAuth 包：<https://pkg.go.dev/github.com/bububa/spotlight-mapi/api/oauth>

