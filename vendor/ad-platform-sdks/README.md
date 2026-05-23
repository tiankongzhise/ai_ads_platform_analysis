# 广告平台 SDK 源码镜像

## 1. 目录说明

本目录保存广告平台 API 调研阶段收集的 SDK 源码快照。快照用于离线审阅、接口字段对照和后续适配器设计，不等同于已经选定为生产依赖。

| 目录 | 平台 | 计划镜像 |
| --- | --- | --- |
| `douyin/` | 抖音/巨量引擎 | 官方 Go SDK `oceanengine/ad_open_sdk_go`。 |
| `tencent/` | 腾讯广告 | 官方 Go SDK `tencentad/marketing-api-go-sdk`。 |
| `baidu/` | 百度营销 | 官方 Python/Java SDK `baidu/baiduads-sdk`。 |
| `xiaohongshu/` | 小红书聚光 | 社区 Go SDK `bububa/spotlight-mapi`。 |

## 2. 使用原则

- 官方 SDK 可作为优先参考，但仍需在开发前确认版本、许可证和 API 权限。
- 社区 SDK 只作为接口形态和字段模型参考，生产使用前必须做许可证、维护频率、安全和错误处理审计。
- 不使用 Git submodule，便于离线交付和审阅。
- 每个平台子目录保留 `SOURCE.md`，记录来源、拉取命令、快照时间、commit/tag 和风险。

