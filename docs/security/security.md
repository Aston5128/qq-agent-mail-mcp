# Security

[简体中文](security.zh-CN.md)

This project bridges an agent runtime to an email account. Treat the MCP server
as a sensitive service.

## Network exposure

MCP bearer-token authentication is not implemented yet. Until it is, do not
expose the server directly to an untrusted network.

Recommended options:

- bind to `127.0.0.1`;
- place the server and client on a private network;
- protect access with a reverse proxy, VPN, firewall, or equivalent control.

## Tool surface

The server exposes named tools only. It does not expose:

- raw shell execution;
- arbitrary CLI argument execution;
- delete/trash operations.

`agently-cli message +trash` is intentionally not forwarded in the current
version.

## Sending mail

Send, reply, and forward preserve the `agently-cli` confirmation-token flow. The
server does not auto-confirm outbound messages.

## Secrets and logs

Do not log:

- OAuth tokens;
- refresh tokens;
- message bodies;
- attachment contents.

CLI stderr and stdout error text are truncated before being returned.

## Attachments and paths

Attachments should be read from and written to controlled directories. Avoid
mounting broad host paths into the runtime container unless the agent is
supposed to access them.

## Application-specific policy

Keep application-specific allowlists, mail processing rules, and business logic
outside this generic bridge. If specialized behavior is needed later, add it as
explicitly named tools rather than hiding it inside the generic pass-through
tools.
