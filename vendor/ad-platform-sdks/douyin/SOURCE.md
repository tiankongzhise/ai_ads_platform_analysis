# 抖音/巨量引擎 SDK 来源

| 项 | 内容 |
| --- | --- |
| 平台 | 抖音/巨量引擎 |
| SDK | 官方 Go SDK |
| 仓库 | <https://github.com/oceanengine/ad_open_sdk_go> |
| 许可证 | Apache-2.0 |
| 可信度 | 官方 SDK，可作为 Go 侧优先参考。 |
| 拉取方式 | `git clone --depth 1 https://github.com/oceanengine/ad_open_sdk_go.git sdk` |
| 快照分支 | `master` |
| 快照 commit | `6c1395d` |
| 快照时间 | 2026-05-23 |
| 本轮用途 | OAuth2、报表接口、模型字段和 CommonApi 调用参考。 |

## 风险

- SDK 中存在大量生成代码，生产依赖前需要锁定 tag 或 commit。
- SDK 接口存在不代表当前应用已获得对应 API 权限。
- Python 服务如不直接依赖 Go SDK，需要按官方文档封装 REST 客户端。
