# 腾讯广告 SDK 来源

| 项 | 内容 |
| --- | --- |
| 平台 | 腾讯广告 |
| SDK | 官方 Go SDK |
| 仓库 | <https://github.com/tencentad/marketing-api-go-sdk> |
| 许可证 | Apache-2.0 |
| 可信度 | 官方 SDK，可作为 Go 侧优先参考。 |
| 拉取方式 | `git clone --depth 1 https://github.com/tencentad/marketing-api-go-sdk.git sdk` |
| 快照分支 | `master` |
| 快照 commit | `f7a057e` |
| 快照时间 | 2026-05-23 |
| 本轮用途 | OAuth2、DailyReports、HourlyReports、Report 服务和模型字段参考。 |

## 风险

- 本轮未确认官方 Python SDK，Python 服务需 REST 封装或通过 Go 适配器调用。
- refresh token 刷新语义和接口频控需以当前官方文档和应用权限为准。
