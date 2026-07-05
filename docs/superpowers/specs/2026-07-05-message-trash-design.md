# Message Trash Tool Design

## Context

QQ Agent Mail MCP currently exposes typed MCP tools for read, search, send,
reply, forward, and attachment operations. It intentionally avoids a generic
`agently_cli(args)` executor and maps each MCP tool to a specific
`agently-cli` command.

The project documentation previously left `agently-cli message +trash` out of
the public MCP surface. The new requirement is to expose the same trash
operation through MCP.

## Goal

Add a named MCP tool, `agently_message_trash`, that moves one message to the QQ
Agent Mail trash folder by forwarding to:

```text
agently-cli message +trash
```

The tool name should use `trash`, not `delete`, because the backing CLI command
is `trash` and the behavior is recoverable trashing rather than permanent
deletion. QQ Agent Mail keeps trashed messages for 30 days.

## Non-Goals

- Do not add permanent delete support.
- Do not add a generic CLI pass-through.
- Do not add an MCP-specific second confirmation layer.
- Do not auto-confirm the CLI confirmation-token flow.

## Tool Shape

Add `MessageTrashInput` in `internal/agently/tools.go`:

- `id`: required message ID.
- `confirmation_token`: optional token returned by the first CLI trash call.

Its `Args()` method should produce:

```text
message +trash --id <id>
```

When `confirmation_token` is provided, it should append:

```text
--confirmation-token <token>
```

The field name should match existing send, reply, and forward tools:
`confirmation_token`.

## Server Registration

Register the input type as an MCP tool named `agently_message_trash` in
`internal/server/server.go`.

The description should make the recoverable behavior clear, for example:

```text
Move a message to trash using agently-cli confirmation flow.
```

The generic forwarding path remains unchanged:

1. MCP client calls `agently_message_trash`.
2. The typed input builds CLI arguments.
3. `Runner.Run` executes `agently-cli`.
4. Parsed JSON stdout is returned as MCP structured content.
5. Nonzero CLI exits are returned as existing structured MCP tool errors.

## Confirmation Behavior

The MCP server should not add its own extra confirmation flag. The operation is
recoverable because messages go to trash and are retained by QQ Agent Mail for
30 days.

The server should still preserve the CLI confirmation-token behavior. If the
CLI first call returns a confirmation payload, the client can call the same tool
again with `confirmation_token`.

## Documentation Updates

Update English and Chinese docs so they no longer say delete/trash operations
are unexposed. The docs should instead state:

- `agently_message_trash` is available.
- It maps to `agently-cli message +trash`.
- It moves a message to trash rather than permanently deleting it.
- QQ Agent Mail keeps trashed messages for 30 days.
- The server preserves the CLI confirmation-token flow and does not auto-confirm
  the operation.

Affected docs:

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

## Testing

Add focused coverage without broad refactoring:

- `internal/agently/tools_test.go`: verify `MessageTrashInput.Args()` with only
  `id`.
- `internal/agently/tools_test.go`: verify `MessageTrashInput.Args()` with
  `confirmation_token`.
- `internal/server/streamable_http_test.go`: verify `agently_message_trash`
  is registered and can be called through StreamableHTTP, using the existing
  fake CLI pattern.

Run:

```bash
go test ./...
go vet ./...
```
