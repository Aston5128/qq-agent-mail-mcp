# Tools

[简体中文](tools.zh-CN.md)

The server exposes named MCP tools. It does not expose a raw
`agently_cli(args)` executor.

## Tool list

| MCP tool | Backing command | Purpose |
| --- | --- | --- |
| `agently_me` | `agently-cli +me` | Show account and alias information. |
| `agently_message_list` | `agently-cli message +list` | List messages. |
| `agently_message_read` | `agently-cli message +read` | Read one message by ID. |
| `agently_message_search` | `agently-cli message +search` | Search messages. |
| `agently_message_send` | `agently-cli message +send` | Send a message with confirmation-token flow. |
| `agently_message_reply` | `agently-cli message +reply` | Reply to a message. |
| `agently_message_forward` | `agently-cli message +forward` | Forward a message. |
| `agently_message_trash` | `agently-cli message +trash` | Move a message to trash. |
| `agently_attachment_upload` | `agently-cli attachment +upload` | Upload a local file for use as an attachment. |
| `agently_attachment_download` | `agently-cli attachment +download` | Download an attachment. |

## Common list/search filters

`agently_message_list` supports:

- `limit`
- `cursor`
- `dir`
- `after`
- `before`
- `is_unread`
- `has_attachments`

`agently_message_search` also supports:

- `q`
- `search_in`
- `from`
- `to`

## Read

`agently_message_read` requires:

- `id`

## Send / reply / forward

`agently_message_send` supports:

- `to`
- `cc`
- `bcc`
- `subject`
- `body`
- `body_file`
- `body_format`
- `attachment`
- `confirmation_token`

`agently_message_reply` supports:

- `id`
- `body`
- `body_file`
- `body_format`
- `cc`
- `bcc`
- `attachment`
- `confirmation_token`
- `reply_all`

`agently_message_forward` supports:

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

## Trash

`agently_message_trash` supports:

- `id`
- `confirmation_token`

It moves a message to trash rather than permanently deleting it. QQ Agent Mail
keeps trashed messages for 30 days.

## Attachments

`agently_attachment_upload` requires:

- `file`

`agently_attachment_download` requires:

- `message_id`
- `attachment_id`

Optional:

- `output`

## Confirmation

`agently-cli message +send`, `+reply`, `+forward`, and `+trash` use a
confirmation-token flow:

1. First call returns a confirmation payload.
2. Second call includes `confirmation_token` to complete the action.

The MCP server preserves this behavior and does not auto-confirm these actions.
