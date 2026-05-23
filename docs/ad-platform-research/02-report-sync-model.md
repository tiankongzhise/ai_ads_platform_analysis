# 广告报表同步与异步获取模型

## 1. 本项目统一模型

广告报表获取统一由 `ad-sync-service` 执行，所有平台适配器都必须输出原始记录并写入 `ad_raw_reports.raw_payload`。标准字段映射只在 ETL 阶段执行。

```text
ad.sync.requested
  -> 创建 ad_sync_jobs
  -> 读取 token_ref 并检查 token 状态
  -> 按平台构造报表请求
  -> 同步分页拉取或创建异步任务
  -> 将平台原始记录写入 ad_raw_reports
  -> 更新 ad_sync_jobs.result_meta
  -> 投递 etl.ad_raw.ready
```

## 2. 报表获取类型

| 类型 | 适用场景 | 处理方式 |
| --- | --- | --- |
| 同步分页报表 | 日报、小时报、账户/计划/单元/创意等常规粒度，数据量可控。 | 请求时带日期、层级、筛选、分页参数；逐页写入原始表。 |
| 异步报表任务 | 大范围历史回溯、大字段集合、代理商返点、素材明细或平台明确要求异步的报表。 | 先创建任务，再轮询状态，成功后下载文件或结果列表。 |
| 订阅/RDS/SPI 数据 | 平台提供推送、RDS 同步或事件订阅时。 | 作为补充通道，仍需归入原始层并记录来源。 |

## 3. 通用任务字段

`ad_sync_jobs.request_meta` 建议保存：

```json
{
  "platform": "douyin",
  "external_account_id": "123",
  "report_type": "campaign_daily",
  "date_from": "2026-05-01",
  "date_to": "2026-05-23",
  "level": "campaign",
  "time_granularity": "daily",
  "fields": ["spend", "impressions", "clicks", "conversions"],
  "filters": [],
  "page_size": 1000,
  "source": "api"
}
```

`ad_sync_jobs.result_meta` 建议保存：

```json
{
  "total_records": 1200,
  "pages": 2,
  "external_task_id": "optional",
  "download_url_ref": "optional",
  "platform_request_id": "optional",
  "source_file_ref": "optional"
}
```

## 4. 原始入库与去重键

每条平台返回记录建议生成稳定去重键：

```text
tenant_id
+ platform
+ account_id
+ report_type
+ report_date
+ external_entity_type
+ external_entity_id
+ time_granularity
+ payload_hash
```

`payload_hash` 使用平台原始记录规范化 JSON 后计算。若平台返回天然主键和更新时间，则优先将其写入 `normalized_key`，并保留完整 `raw_payload`。

## 5. 同步报表分页规则

同步分页适配器必须处理：

- 页码分页：`page`、`page_size`、`total_page`。
- 游标分页：`cursor`、`next_cursor`、`has_more`。
- 时间切片：当平台限制单次日期跨度时，将日期范围拆成多天或多周。
- 字段切片：当字段数量过多或指标互斥时，将字段拆分为多个任务。
- 层级切片：账户、计划、单元、创意、关键词等层级分别拉取，不能混成单一报表类型。

## 6. 异步报表任务规则

异步任务适配器建议使用以下状态映射：

| 平台状态 | 系统状态 | 处理 |
| --- | --- | --- |
| created / pending | `running` | 继续轮询。 |
| running / processing | `running` | 按退避策略轮询。 |
| success / completed | `success` | 下载或读取结果并入库。 |
| failed | `failed` | 记录平台错误并按可重试性处理。 |
| expired | `failed` | 重新创建任务。 |

轮询策略：

- 首次等待 5 到 10 秒。
- 后续指数退避，最高不超过 5 分钟。
- 任务超时后标记失败并记录 `external_task_id`。
- 下载地址必须视为敏感临时 URL，不写入长期明文日志。

## 7. 冲突与限制清单

| 限制类型 | 说明 | 处理建议 |
| --- | --- | --- |
| 日期范围限制 | 平台通常限制单次查询跨度，历史数据也可能只保留固定窗口。 | 任务创建时按平台配置自动拆片。 |
| 维度和指标冲突 | 某些指标只支持特定层级、时间口径或过滤条件。 | 平台适配器维护 `report_capabilities`，请求前校验。 |
| 层级冲突 | 账户、计划、单元、创意、关键词等报表层级字段不同。 | `report_type` 必须包含层级，不做跨层级合并。 |
| 时间口径冲突 | 请求时间、归因时间、转化发生时间等口径可能不同。 | 原始层保留平台口径字段，ETL 输出时标注指标版本。 |
| 金额单位冲突 | 平台可能以元、分或厘返回金额。 | 原始层不改动，ETL 根据平台和字段配置换算。 |
| 分页限制 | 单页大小、页码最大值、总行数上限不同。 | 对大账户按日期、层级或实体 ID 拆分。 |
| 并发和频控限制 | 平台会按 app、广告主、接口或 token 限制 QPS。 | Redis 令牌桶 + 平台错误码退避。 |
| 数据延迟 | 当天数据可能延迟或多次回补。 | 近 7 天滚动重拉，历史数据按需补偿。 |
| 异步任务保留期 | 异步结果文件或下载 URL 可能短期有效。 | 成功后立即下载到对象存储并写入 `source_file_ref`。 |

## 8. 平台差异摘要

| 平台 | 已确认报表能力 | 同步/异步建议 |
| --- | --- | --- |
| 抖音/巨量引擎 | 官方开放平台包含数据报表、自定义报表、代理商消耗报表、异步下载任务、RDS/SPI 订阅能力。 | 常规日报走同步 API；代理返点、素材或大文件类走异步任务；有 RDS/SPI 权限时作为增量补充。 |
| 腾讯广告 | 日报接口支持账户、计划、项目、广告组、广告、素材、关键词等层级；公开资料显示日期可查近 365 天，分页最大 `page_size` 可到 1000。 | 首批按 `daily_reports/get` 同步分页接入；大账户按日期和层级拆分。 |
| 百度营销 | `ReportService` 支持实时报告和异步报告；异步文件下载 URL 可能短期有效。 | 小范围走实时报告；大范围走异步报告并立即转存文件。 |
| 小红书聚光 | 官方开放平台说明支持推广投放多维报表和多账号批量获取；社区 Go SDK 暴露实时数据报表包。 | 生产前复核官方报表接口；首批可按账户/计划/单元/创意实时报表 REST 封装。 |

## 9. 适配器能力配置

每个平台应配置能力表，不把限制写死在业务流程中：

```json
{
  "platform": "tencent",
  "report_type": "daily_report_adgroup",
  "sync_mode": "paged",
  "max_date_span_days": 31,
  "history_window_days": 365,
  "max_page_size": 1000,
  "supported_granularity": ["daily"],
  "amount_unit": "cent",
  "supports_async": false,
  "rate_limit_policy": "platform_default"
}
```

## 10. 来源

- 巨量引擎开放平台入门指南与接口清单：<https://open.oceanengine.com/labels/34>
- 巨量引擎开放平台首页：<https://open.oceanengine.com/>
- 腾讯广告日报接口资料：<https://developers.e.qq.com/docs/api/insights/ad_insights/daily_reports_get>
- 腾讯广告日报接口 Apifox 镜像：<https://7j0rt64lfd.apifox.cn/api-122863231>
- 百度 SEM API PDF 资料：<https://bce-cdn.bj.bcebos.com/doc/pdf/SEM.zh.pdf>
- 小红书商业开放平台：<https://ad-market.xiaohongshu.com/>
- 小红书聚光社区 Go SDK 实时报表包：<https://pkg.go.dev/github.com/bububa/spotlight-mapi/api/report/realtime>

