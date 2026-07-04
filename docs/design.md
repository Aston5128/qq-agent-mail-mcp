# 设计说明

本文档描述 QQ Agent Mail MCP 的通用设计。它不包含任何具体应用的工作流。

## 目标

通过一个轻量、可检查的 `agently-cli` 封装层，把 QQ Agent Mail 能力暴露给支持 MCP 的 Agent。

Server 应该让 Agent 更方便地使用 CLI，而不是自己变成一个邮件自动化平台。

## 非目标

- 不暴露原始 shell 执行。
- 不提供通用的 `agently_cli(args)` 工具。
- 第一版不包含具体应用的邮件处理逻辑。
- 不隐藏自动发送确认。
- 不长期存储邮件正文或附件。
- 当前版本不暴露删除/回收站能力（见下文「删除（回收站）」）。

## 工具映射

Server 应该暴露命名的 MCP 工具，这些工具直接对应稳定的 `agently-cli` 命令。

| 工具 | CLI 命令 | 说明 |
| --- | --- | --- |
| `agently_me` | `agently-cli +me` | 返回账号和别名信息。 |
| `agently_message_list` | `agently-cli message +list` | 支持安全的过滤参数：limit、cursor、dir、before、after、unread、attachments。 |
| `agently_message_read` | `agently-cli message +read` | 按消息 ID 读取单条消息。 |
| `agently_message_search` | `agently-cli message +search` | 按关键词搜索消息。 |
| `agently_message_send` | `agently-cli message +send` | 保留 CLI 的 confirmation-token 流程。 |
| `agently_message_reply` | `agently-cli message +reply` | 回复已有消息。 |
| `agently_message_forward` | `agently-cli message +forward` | 转发已有消息。 |
| `agently_attachment_upload` | `agently-cli attachment +upload` | 上传本地文件作为附件。 |
| `agently_attachment_download` | `agently-cli attachment +download` | 把普通附件下载到受控目录。 |

> 删除/回收站（`agently-cli message +trash`）CLI 已支持，但当前版本刻意不暴露，见下文「删除（回收站）」。

## 结构化透传

封装层应该把类型化的 MCP 参数翻译成 CLI 标志。

不应该接受任意 CLI 参数数组，因为那样会：

- 让工具更难被 Agent 理解；
- 绕过校验；
- 扩大安全攻击面；
- 让兼容性更难维护。

输出应该尽量贴近 `agently-cli` 的 JSON 输出。当 CLI 返回 JSON 时，返回解析后的 JSON。当 CLI 返回非 JSON 错误时，返回结构化的 MCP 错误，并对 stderr 做脱敏处理。

当 `agently-cli` 非零退出但 stdout 是 JSON 时，Server 应该保留这份 JSON，并把它放进 `isError=true` 的 MCP tool result 中。未授权、未找到、参数错误这类 CLI 级错误不应该变成 MCP 协议级错误，否则上层 Agent 很难根据错误类型自我修正。

## 发送确认

`agently-cli message +send` 刻意设计为两阶段流程。

MCP Server 不应该对调用方隐藏这个行为：

1. 第一次请求（不带 `confirmation_token`）返回 CLI 的确认载荷。
2. 第二次请求（带 `confirmation_token`）完成发送。

如果将来确实需要专门的自动确认行为，应该作为独立命名的工具实现，而不是混进通用透传工具里。

## 删除（回收站）

`agently-cli` 提供 `message +trash`：软删除（移入回收站，保留 30 天后彻底清除），与 `+send` 一样是两阶段确认流程。已实测可用（0.0.3 阶段：收件箱 3 封邮件经 `+trash` 全部成功移入回收站）。

当前版本**刻意不向 Agent 暴露删除能力**——上表工具映射不含 trash。账号 `+me` 的 scopes 虽包含 `mail:delete`，但 Server 不转发该操作。

未来计划暴露删除，但会**在 MCP 层再加一道二次确认**（独立于 CLI 自带的两阶段 confirmation-token），避免 Agent 误删。具体形态待定。

## 运行时模型

概念上支持两种部署形态：

1. 远程 MCP sidecar：适合保持 Agent 容器较小。
2. 本地 stdio 命令：适合只支持在 Agent 容器内启动 MCP 工具的运行时。

凭据持久化是决定因素。第一个实现任务应该先验证 `agently-cli` 在目标 Linux 环境中如何存储凭据。

## 语言选择

Go 是实现语言，因为这个项目是一个轻量的进程封装层，应该有较小的内存占用。

TypeScript 仍然可以作为 MCP 示例的参考点，但不是这个项目的目标运行时。
