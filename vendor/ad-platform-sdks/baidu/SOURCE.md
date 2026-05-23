# 百度营销 SDK 来源

| 项 | 内容 |
| --- | --- |
| 平台 | 百度营销 |
| SDK | 官方 Java/Python SDK |
| 仓库 | <https://github.com/baidu/baiduads-sdk> |
| PyPI | <https://pypi.org/project/baiduads-sdk/> |
| 许可证 | Apache-2.0 |
| 可信度 | 官方 SDK，Python 侧可优先参考。 |
| 拉取方式 | `git clone --depth 1 https://github.com/baidu/baiduads-sdk.git sdk` |
| 快照分支 | `main` |
| 快照 commit | `ea0d574` |
| 快照时间 | 2026-05-23 |
| 本轮用途 | Python SDK 调用、`ApiRequestHeader`、`OAuthAuthorizedToolAPI` 和 `ReportService` 参考。 |

## 风险

- 百度请求头模型依赖 `userName + accessToken`，与其他 OAuth2 平台不同。
- Go SDK 未确认官方实现，社区 Go SDK 不纳入默认生产依赖。
- 授权模式和 token 有效期需要用当前开发者中心文档复核。
