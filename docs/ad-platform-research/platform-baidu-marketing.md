# 百度营销 API 调研

## 1. 平台定位

百度营销 API 面向百度搜索推广、信息流和相关商业产品，适合拉取搜索广告消耗、展现、点击、转化和报告数据。本项目平台代码使用 `baidu`。

开放平台和 SDK 入口：

- 百度商业 API 开发者中心：<https://dev2.baidu.com/>
- 官方 SDK 仓库：<https://github.com/baidu/baiduads-sdk>
- PyPI 包：<https://pypi.org/project/baiduads-sdk/>
- 百度 SEM API PDF 资料：<https://bce-cdn.bj.bcebos.com/doc/pdf/SEM.zh.pdf>

## 2. SDK 结论

| 项 | 结论 |
| --- | --- |
| 官方 Python SDK | 可用，仓库 `baidu/baiduads-sdk` 包含 `python` 实现，PyPI 包名为 `baiduads-sdk`。 |
| 官方 Java SDK | 可用，同一仓库包含 `java` 实现。 |
| 官方 Go SDK | 本轮未确认官方 Go SDK。 |
| 社区 Go SDK | 可找到 `github.com/564104865/baidu-marketing` 等社区实现，只能作为参考。 |
| 本项目建议 | Python `ad-sync-service` 优先使用官方 Python SDK；Go 服务只保存账户引用和任务状态，不直接调用百度报表。 |

## 3. 鉴权和请求头

百度官方 Python SDK 的请求模型以 `ApiRequestHeader` 为核心，常见字段包括：

```text
userName
accessToken
```

也就是说，业务调用通常需要在请求 header 中带百度账户名和 access token，而不是只传 Bearer token。

示例逻辑：

```python
header = ApiRequestHeader(
    user_name="baidu_user",
    access_token="encrypted_token_from_token_ref"
)
```

### 3.1 授权工具

官方 SDK 包含 `OAuthAuthorizedToolAPI`，代表性接口：

```text
OAuthAuthorizedToolAPI.get_auth_code
OAuthAuthorizedToolAPI.refresh_access_token
```

该授权工具需要 `app_id`、`secret_key` 和 `state` 等参数。开发前必须用当前开发者账户确认：

- 是否使用永久授权。
- 是否需要人工授权页跳转。
- refresh access token 的有效期和返回字段。
- 是否存在多个账户名共用一个授权 app 的限制。

### 3.2 本项目 token_ref 映射

百度 token 存储建议额外保存：

- `user_name`：百度推广账户名。
- `access_token`：加密存储。
- `refresh_token`：如授权模式返回则加密存储。
- `app_id`：应用 ID。
- `scope_list`：当前授权权限。
- `auth_mode`：`temporary` 或 `permanent`。

`ad_platform_accounts.external_account_id` 可优先使用百度账户 ID；如果开发阶段只能取得账户名，则写入账户名并在 `meta` 中补充真实账户 ID。

## 4. 报表能力

官方 SDK 和 SEM 文档中包含 `ReportService`。核心能力包括：

| 能力 | SDK/文档接口 | 说明 |
| --- | --- | --- |
| 实时报表 | `ReportService.get_real_time_query_data` | 适合小范围、低延迟查询。 |
| 异步报表创建 | `ReportService.get_professional_report_id` 等 | 创建大范围报告任务并返回 `reportId`。 |
| 异步状态查询 | `ReportService.get_report_state` | 获取报告生成状态。 |
| 异步下载地址 | `ReportService.get_report_file_url` | 获取报告下载 URL，需在报告生成后调用。 |

异步报告适合大量历史数据，下载 URL 有短期有效期风险，获取后应立即转存到对象存储。

## 5. 本项目报表类型建议

| 本项目 `report_type` | 百度层级 | 粒度 | 用途 |
| --- | --- | --- | --- |
| `account_daily` | 账户 | 日 | 账户总消耗和 ROI。 |
| `campaign_daily` | 计划 | 日 | 推广计划效果分析。 |
| `adgroup_daily` | 单元 | 日 | 推广单元效果分析。 |
| `keyword_daily` | 关键词 | 日 | 搜索广告关键词效果，后续可选。 |
| `creative_daily` | 创意 | 日 | 创意效果，后续行动建议。 |

首批建议只实现账户、计划、单元三个层级；关键词和创意可能数据量较大，适合异步报告或后续阶段。

## 6. 限制和冲突点

- 百度 SDK 请求头需要 `userName`，token 存储必须保存账户名和 token 的绑定关系。
- 报表接口存在实时和异步两套模式，不能用单一同步分页模型覆盖所有场景。
- 异步报告需要先查状态，再取文件 URL；状态未完成时直接取 URL 可能返回错误。
- 百度账户体系、代理商账户和客户账户关系需要在权限申请阶段确认。
- 社区 Go SDK 不可直接作为生产依赖，必须审计许可证、错误处理和接口版本。
- SEM PDF 与 SDK 生成文档可能存在版本差异，开发前以当前开发者中心接口为准。

## 7. 适配器建议

```text
BaiduAdapter
  - platform_code(): baidu
  - auth: token_ref 解密为 userName + accessToken
  - fetch: 官方 Python SDK
  - sync_report: get_real_time_query_data
  - async_report: create report -> get_report_state -> get_report_file_url
```

原始记录入库建议额外保存：

- `userName`
- `accountId`
- `campaignId`
- `adgroupId`
- `keywordId`
- `creativeId`
- `date`
- `reportId`
- `file_url_fetched_at`

## 8. 开发前待确认

- 是否已经拿到百度商业 API 应用、`app_id` 和 `secret_key`。
- 授权模式使用临时授权还是永久授权。
- `refresh_access_token` 的真实返回字段和有效期。
- 当前账户可使用的 `ReportService` 报告类型。
- 实时报表最大时间跨度、异步报告保留期和文件格式。
- 搜索推广、信息流推广是否需要分不同产品线建模。

## 9. 来源

- 百度商业 API 开发者中心：<https://dev2.baidu.com/>
- 官方 SDK 仓库：<https://github.com/baidu/baiduads-sdk>
- 官方 Python SDK README：<https://github.com/baidu/baiduads-sdk/blob/main/python/README.md>
- `OAuthAuthorizedToolAPI` 文档：<https://github.com/baidu/baiduads-sdk/blob/main/python/baiduads-sdk-auto/docs/OAuthAuthorizedToolAPI.md>
- `ReportService` 文档：<https://github.com/baidu/baiduads-sdk/blob/main/python/baiduads-sdk-auto/docs/ReportService.md>
- PyPI 包：<https://pypi.org/project/baiduads-sdk/>
- 百度 SEM API PDF 资料：<https://bce-cdn.bj.bcebos.com/doc/pdf/SEM.zh.pdf>
- 社区 Go SDK 包文档：<https://pkg.go.dev/github.com/564104865/baidu-marketing/api/search/report>

