# Design notes

[简体中文](design.zh-CN.md)

This document captures the generic design of QQ Agent Mail MCP. It intentionally
does not include application-specific workflows.

## Goal

Expose QQ Agent Mail to MCP-capable agents through a small, inspectable wrapper
around `agently-cli`.

The server should make the CLI easier for agents to use without becoming an
email automation platform of its own.

## Non-goals

- No raw shell execution.
- No generic `agently_cli(args)` tool.
- No application-specific mail processing logic.
- No hidden automatic send confirmation.
- No long-term storage of message bodies or attachments.
- No delete/trash exposure in the current version.

## Structured pass-through

The wrapper translates typed MCP arguments into CLI flags.

It should not accept arbitrary CLI argument arrays because that would:

- make the tool harder for agents to understand;
- bypass validation;
- expand the security surface;
- make compatibility harder to document.

## Output shape

When `agently-cli` returns JSON, the server returns parsed JSON as structured MCP
content.

When `agently-cli` exits nonzero and stdout is JSON, the server preserves that
JSON and returns it as an `isError=true` MCP tool result. CLI-level errors such
as auth, not found, or invalid arguments should not become MCP protocol errors.

## Send confirmation

`agently-cli message +send`, `+reply`, and `+forward` are two-phase operations:

1. first call returns a confirmation payload;
2. second call includes the confirmation token.

The MCP server preserves this behavior.

## Delete / trash

`agently-cli message +trash` exists and uses a confirmation-token flow, but this
server does not expose it yet. If it is added later, it should have an explicit
MCP-level confirmation layer in addition to the CLI confirmation flow.

## Runtime model

The current implementation targets a remote StreamableHTTP sidecar. stdio may
be added later if a client runtime requires launching MCP servers locally.

## Language choice

Go is used because this project is a thin process wrapper with a small runtime
footprint and simple binary deployment.
