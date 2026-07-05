# 邮件移入回收站工具设计

## 背景

QQ Agent Mail MCP 当前已经暴露了读取、搜索、发送、回复、转发和附件相关的类型化 MCP tools。项目刻意不提供通用的 `agently_cli(args)` 执行器，而是把每个 MCP tool 映射到一个明确的 `agently-cli` 命令。

项目文档此前没有把 `agently-cli message +trash` 暴露到 MCP tool surface。新的需求是通过 MCP 暴露同一个移入回收站操作。

## 目标

新增一个具名 MCP tool：`agently_message_trash`，用于把单封邮件移入 QQ Agent Mail 回收站。底层转发到：

```text
agently-cli message +trash
```

Tool 名称使用 `trash`，不使用 `delete`，因为底层 CLI 命令就是 `trash`，而且行为是可恢复的移入回收站，不是永久删除。QQ Agent Mail 的回收站邮件会保留 30 天。

## 非目标

- 不增加永久删除能力。
- 不增加通用 CLI 透传。
- 不增加 MCP 层额外二次确认。
- 不自动确认 CLI 的 confirmation-token 流程。

## Tool 形状

在 `internal/agently/tools.go` 中新增 `MessageTrashInput`：

- `id`：必填，邮件 ID。
- `confirmation_token`：可选，第一次 CLI trash 调用返回的确认 token。

它的 `Args()` 方法应该生成：

```text
message +trash --id <id>
```

当提供 `confirmation_token` 时，应该追加：

```text
--confirmation-token <token>
```

字段名沿用现有发送、回复、转发工具的命名：`confirmation_token`。

## Server 注册

在 `internal/server/server.go` 中把该输入类型注册为 MCP tool：`agently_message_trash`。

描述需要明确这是可恢复的移入回收站行为，例如：

```text
Move a message to trash using agently-cli confirmation flow.
```

通用转发路径保持不变：

1. MCP client 调用 `agently_message_trash`。
2. 类型化输入构造 CLI 参数。
3. `Runner.Run` 执行 `agently-cli`。
4. 解析后的 JSON stdout 作为 MCP structured content 返回。
5. CLI 非零退出沿用现有 structured MCP tool error 返回。

## 确认行为

MCP server 不增加自己的额外确认字段。这个操作是可恢复的，因为邮件会进入回收站，并由 QQ Agent Mail 保留 30 天。

Server 仍然保留 CLI 自己的 confirmation-token 行为。如果 CLI 第一次调用返回确认载荷，client 可以再次调用同一个 tool，并带上 `confirmation_token`。

## 文档更新

更新英文和中文文档，不再描述为“delete/trash operations 未暴露”。文档应该改为说明：

- `agently_message_trash` 可用。
- 它映射到 `agently-cli message +trash`。
- 它把邮件移入回收站，而不是永久删除。
- QQ Agent Mail 的回收站邮件会保留 30 天。
- Server 保留 CLI confirmation-token 流程，并且不会自动确认该操作。

受影响文档：

- `README.md`
- `docs/readme/README.zh-CN.md`
- `docs/tools/tools.md`
- `docs/tools/tools.zh-CN.md`
- `docs/security/security.md`
- `docs/security/security.zh-CN.md`
- `docs/design/design.md`
- `docs/design/design.zh-CN.md`
- `CHANGELOG.md`
- `docs/changelogs/CHANGELOG.zh-CN.md`

## 测试

增加聚焦覆盖，不做大范围重构：

- `internal/agently/tools_test.go`：验证只带 `id` 的 `MessageTrashInput.Args()`。
- `internal/agently/tools_test.go`：验证带 `confirmation_token` 的 `MessageTrashInput.Args()`。
- `internal/server/streamable_http_test.go`：沿用现有 fake CLI 模式，验证 `agently_message_trash` 已注册并且可以通过 StreamableHTTP 调用。

运行：

```bash
go test ./...
go vet ./...
```
