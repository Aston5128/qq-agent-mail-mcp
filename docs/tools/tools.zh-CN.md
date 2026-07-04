# Tools

[English](tools.md)

server 暴露具名 MCP tools，不提供原始 `agently_cli(args)` 执行器。

## 工具列表

| MCP tool | 底层命令 | 用途 |
| --- | --- | --- |
| `agently_me` | `agently-cli +me` | 查看账号和别名信息。 |
| `agently_message_list` | `agently-cli message +list` | 列出邮件。 |
| `agently_message_read` | `agently-cli message +read` | 按 ID 读取单封邮件。 |
| `agently_message_search` | `agently-cli message +search` | 搜索邮件。 |
| `agently_message_send` | `agently-cli message +send` | 使用 confirmation-token 流程发送邮件。 |
| `agently_message_reply` | `agently-cli message +reply` | 回复邮件。 |
| `agently_message_forward` | `agently-cli message +forward` | 转发邮件。 |
| `agently_attachment_upload` | `agently-cli attachment +upload` | 上传本地文件作为附件。 |
| `agently_attachment_download` | `agently-cli attachment +download` | 下载附件。 |

## list/search 通用过滤参数

`agently_message_list` 支持：

- `limit`
- `cursor`
- `dir`
- `after`
- `before`
- `is_unread`
- `has_attachments`

`agently_message_search` 额外支持：

- `q`
- `search_in`
- `from`
- `to`

## 读取

`agently_message_read` 需要：

- `id`

## 发送 / 回复 / 转发

`agently_message_send` 支持：

- `to`
- `cc`
- `bcc`
- `subject`
- `body`
- `body_file`
- `body_format`
- `attachment`
- `confirmation_token`

`agently_message_reply` 支持：

- `id`
- `body`
- `body_file`
- `body_format`
- `cc`
- `bcc`
- `attachment`
- `confirmation_token`
- `reply_all`

`agently_message_forward` 支持：

- `id`
- `to`
- `cc`
- `bcc`
- `body`
- `body_file`
- `body_format`
- `attachment`
- `confirmation_token`
- `include_attachments`

## 附件

`agently_attachment_upload` 需要：

- `file`

`agently_attachment_download` 需要：

- `message_id`
- `attachment_id`

可选：

- `output`

## 发送确认

`agently-cli message +send`、`+reply`、`+forward` 使用 confirmation-token 流程：

1. 第一次调用返回确认载荷。
2. 第二次调用带上 `confirmation_token` 完成操作。

MCP server 保留这个行为，不会自动确认发送。

## 未暴露

当前版本刻意不暴露 `agently-cli message +trash`，即使 CLI 本身支持。未来如果增加删除/回收站能力，也应该在 CLI confirmation 之外再加 MCP 层二次确认。
