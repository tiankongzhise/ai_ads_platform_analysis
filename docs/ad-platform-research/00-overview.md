# 广告平台 API 调研总览

## 1. 平台清单

| 平台 | 开放平台入口 | 本项目平台代码 | 初步接入优先级 | 说明 |
| --- | --- | --- | --- | --- |
| 抖音/巨量引擎 | <https://open.oceanengine.com/> | `douyin` | P0 | 官方开放平台和官方 Go SDK 可用，适合优先接入。 |
| 腾讯广告 | <https://developers.e.qq.com/> | `tencent` | P0 | 官方 Marketing API 和官方 Go SDK 可用，适合作为首批接入。 |
| 百度营销 | <https://dev2.baidu.com/> | `baidu` | P1 | 官方 SDK 以 Java/Python 为主，Python 侧适配成本较低。 |
| 小红书聚光 | <https://ad-market.xiaohongshu.com/> | `xiaohongshu` | P1 | 官方开放平台入口可确认，公开 SDK 信息较少，需重点跟进权限和文档完整性。 |

## 2. SDK 可用性

| 平台 | Go SDK | Python SDK | 其他 SDK | 本轮建议 |
| --- | --- | --- | --- | --- |
| 抖音/巨量引擎 | 官方 `oceanengine/ad_open_sdk_go` | 未确认官方 Python SDK | 官方文档提及 Java/Go SDK | Go 服务可直接参考官方 SDK；Python 侧优先 REST 封装。 |
| 腾讯广告 | 官方 `tencentad/marketing-api-go-sdk` | 未确认官方 Python SDK | 官方 Java SDK 可作为补充 | Go 侧优先官方 SDK；Python 侧先 REST 封装或经内部 Go 适配器转发。 |
| 百度营销 | 未确认官方 Go SDK | 官方 `baidu/baiduads-sdk` Python | 官方 Java SDK | Python `ad-sync-service` 可优先使用官方 Python SDK；Go 侧只保留社区 SDK 参考。 |
| 小红书聚光 | 未确认官方 Go SDK；社区 `bububa/spotlight-mapi` 可参考 | 未确认官方 Python SDK | 社区 PHP SDK 可参考 | 不把社区 SDK 作为默认生产依赖，先以官方文档和 REST 封装为准。 |

## 3. 与现有架构的关系

- `channel-service` 保存渠道配置、平台账户引用和同步策略。
- `callback-adapter-service` 只消费已完成广告回调能力提供的可信授权结果。
- `ad-sync-service` 负责平台 SDK/REST 调用、分页、限流、错误分类、原始报表入库和同步状态更新。
- `etl-service` 在原始报表入库后再做标准字段映射，避免平台字段差异污染通用同步框架。

## 4. 首批落地顺序建议

1. 先接抖音/巨量引擎和腾讯广告，原因是 Go 官方 SDK 可用，授权与报表能力更容易纳入通用适配器。
2. 再接百度营销，优先在 Python 服务侧使用官方 SDK，补齐 Go 控制面只需要的账户状态和任务触发能力。
3. 最后接小红书聚光，先完成权限申请、官方文档确认和 REST 客户端封装，再评估社区 SDK 是否可生产使用。

## 5. 调研输出判断

本轮文档不承诺所有平台 API 字段已经可直接编码。后续进入开发前，每个平台还需要补齐：

- 开发者应用审批状态和可用权限包。
- 生产 app id、secret、回调域名和白名单。
- 可调用的账户、广告主、报表类型和字段清单。
- 平台限流、数据延迟、历史回溯窗口和异步任务保留时长。
- 沙箱或授权测试账户。

