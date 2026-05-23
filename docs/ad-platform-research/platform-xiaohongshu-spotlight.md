# 小红书聚光 Marketing API 调研

## 1. 平台定位

小红书聚光是小红书广告投放平台，Marketing API 面向品牌和代理商提供账户、投放、报表和批量数据获取能力。本项目平台代码使用 `xiaohongshu`。

开放平台入口：

- 小红书商业开放平台：<https://ad-market.xiaohongshu.com/>
- 小红书聚光平台：<https://ad.xiaohongshu.com/>
- 小红书商业市场：<https://www.xiaohongshu.com/business>

## 2. SDK 结论

| 项 | 结论 |
| --- | --- |
| 官方 Go SDK | 本轮未确认公开官方 Go SDK。 |
| 官方 Python SDK | 本轮未确认公开官方 Python SDK。 |
| 社区 Go SDK | `bububa/spotlight-mapi` 可用作接口形态参考，包文档显示 Apache-2.0 许可证。 |
| 社区 PHP SDK | Packagist 上存在聚光 Marketing API SDK，但需确认是否官方维护。 |
| 本项目建议 | 首批不把社区 SDK 作为生产默认依赖；先以官方文档和 REST 封装为准，社区 SDK 只用于识别接口分组和字段模型。 |

公开开放平台页面可确认：

- 支持 Marketing API。
- 支持推广投放多维报表查询。
- 支持多账号数据批量获取。
- 支持客户自有平台进行数据洞察分析。

## 3. OAuth2 认证

小红书官方完整 OAuth 文档可能需要登录开放平台查看。本轮从社区 Go SDK 可确认常见 OAuth 方法形态：

```text
AccessToken
RefreshToken
```

社区 SDK 的 OAuth 包包含 `AccessTokenRequest`、`AccessTokenResponse`、`RefreshTokenRequest` 等模型，响应模型中可见 access token 和 refresh token 过期时间字段。

生产接入前必须通过官方文档确认：

- 授权链接和回调域名配置方式。
- 回调授权码字段名。
- `access_token` 接口地址、签名方式和必填参数。
- `refresh_token` 接口地址、刷新后是否返回新的 refresh token。
- token 有效期和提前刷新窗口。
- 广告主、代理商和多账号授权关系。

## 4. 报表能力

开放平台官网说明支持推广投放多维度/全方位报表数据查询。社区 Go SDK 公开包中可见报表模块：

- 实时报表：`api/report/realtime`
- 离线报表：`api/report/offline`

离线报表包中可见账户层级、计划层级、单元层级、创意层级相关方法。实时数据包可作为首批 REST 封装的接口分组参考，但字段和路径必须以官方文档为准。

## 5. 本项目报表类型建议

| 本项目 `report_type` | 聚光层级 | 粒度 | 用途 |
| --- | --- | --- | --- |
| `account_daily` | 账户 | 日 | 账户整体消耗和 ROI。 |
| `campaign_daily` | 计划 | 日 | 推广计划效果。 |
| `unit_daily` | 单元 | 日 | 投放单元效果。 |
| `creative_daily` | 创意 | 日 | 创意效果和素材分析。 |
| `account_realtime` | 账户 | 实时 | 演示账户和近实时看板。 |

首批建议先接账户、计划、单元、创意四个常规层级；实时数据只作为演示账户增强能力，不影响小时聚合主链路。

## 6. 限制和冲突点

- 官方 API 文档可能需要登录和应用审批后才能完整访问。
- 社区 SDK 的接口路径、字段和错误码可能落后于官方版本。
- 多账号报表涉及代理商/客户账号绑定关系，需明确数据权限边界。
- 聚光平台和小红书其他商业产品可能存在账户体系差异，不能默认共用 token。
- 报表层级、实时/离线报表字段和归因口径需要开发前用测试账号复核。

## 7. 适配器建议

```text
XiaohongshuAdapter
  - platform_code(): xiaohongshu
  - auth: 官方 OAuth2 文档确认后实现
  - fetch: REST 客户端优先
  - report: account_daily/campaign_daily/unit_daily/creative_daily
  - sdk_reference: bububa/spotlight-mapi 只作参考
```

原始记录入库建议额外保存：

- `advertiser_id`
- `campaign_id`
- `unit_id`
- `creative_id`
- `date`
- `report_scene`
- `request_id`

## 8. 开发前待确认

- 是否具备小红书商业开放平台应用资质。
- 是否已经申请 Marketing API 和报表权限。
- 授权链路、签名算法和 token 生命周期。
- 代理商多账号授权与客户账号授权的差异。
- 实时/离线报表接口字段、分页、日期跨度和频控。
- 社区 SDK 是否允许用于生产，许可证、维护频率和接口版本是否满足要求。

## 9. 来源

- 小红书商业开放平台：<https://ad-market.xiaohongshu.com/>
- 小红书聚光平台：<https://ad.xiaohongshu.com/>
- 小红书商业市场：<https://www.xiaohongshu.com/business>
- 社区 Go SDK：<https://github.com/bububa/spotlight-mapi>
- 社区 Go SDK OAuth 包：<https://pkg.go.dev/github.com/bububa/spotlight-mapi/api/oauth>
- 社区 Go SDK 实时报表包：<https://pkg.go.dev/github.com/bububa/spotlight-mapi/api/report/realtime>
- 社区 Go SDK 离线报表包：<https://pkg.go.dev/github.com/bububa/spotlight-mapi/api/report/offline>

