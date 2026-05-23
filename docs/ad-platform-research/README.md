# 广告平台 API 调研索引

## 1. 本轮目标

本目录用于沉淀教育广告 CRM 数据分析平台首批广告渠道的前期接入调研。首批平台包括抖音/巨量引擎、腾讯广告、百度营销和小红书聚光，调研重点是开放平台入口、SDK 可用性、OAuth2 授权、token 生命周期、报表获取方式、限流约束和后续适配建议。

本轮调研只形成接入资料和 SDK 快照，不实现业务代码。后续开发应在 `ad-sync-service` 的平台适配器中消费这些资料。

## 2. 文档结构

| 文档 | 内容 |
| --- | --- |
| `00-overview.md` | 平台清单、SDK 可信度、接入优先级和调研结论。 |
| `01-oauth2-token-lifecycle.md` | OAuth2 授权、回调、换 token、刷新 token 和 token 存储建议。 |
| `02-report-sync-model.md` | 同步/异步报表、分页、轮询、限流、去重和原始入库模型。 |
| `platform-douyin-oceanengine.md` | 抖音/巨量引擎开放平台专项调研。 |
| `platform-tencent-ads.md` | 腾讯广告 Marketing API 专项调研。 |
| `platform-baidu-marketing.md` | 百度营销 API 专项调研。 |
| `platform-xiaohongshu-spotlight.md` | 小红书聚光 Marketing API 专项调研。 |
| `99-risks-and-next-steps.md` | 接入风险、权限申请、验收清单和后续建议。 |

SDK 源码镜像放在 `vendor/ad-platform-sdks/`，每个平台子目录必须包含 `SOURCE.md`，记录来源、语言、版本、许可证和生产使用风险。

## 3. 调研原则

- 优先采用官方开放平台文档和官方 SDK。
- 官方 SDK 不完整时，可记录社区 SDK，但必须标注为参考实现，生产使用前需要二次审计。
- 不在原始入库阶段强行拉平平台差异，平台返回记录先保存到 `ad_raw_reports.raw_payload`。
- OAuth token 不写入业务表明文字段，只保存加密 token 存储或密钥服务引用。
- 报表同步要按平台限流、数据可回溯和幂等重试设计。

