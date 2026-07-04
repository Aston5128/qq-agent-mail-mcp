# QQ Agent Mail MCP

[English](../../README.md) | **简体中文**

QQ Agent Mail MCP 是一个轻量级 MCP Server，用来通过腾讯 QQ 邮箱的 `agently-cli` 暴露 QQ Agent 邮箱能力。

这个项目定位为一层很薄的结构化透传层：把 `agently-cli` 中有用的邮件操作包装成 MCP tools，同时把 CLI 安装和 OAuth 授权状态从具体 Agent 运行环境中隔离出来。

## 当前状态

0.0.3 PoC 已实现。

当前已经完成：

- Go 版 StreamableHTTP MCP Server；
- 结构化 MCP tools 到 `agently-cli` 的 1:1 透传；
- CLI stdout JSON 解析为 MCP structured content；
- `agently-cli` 非零退出时，stdout 中的 JSON 错误会作为结构化 tool error 返回；
- 本地 fake `agently-cli` 集成测试；
- 容器化部署（`Dockerfile` + `docker-compose.yml`），且 `agently-cli` 的 OAuth 凭证持久化已在真实宿主机验证（token 能在 `docker compose down` / `up` 后保留）。

暂未实现：

- MCP bearer token 认证；
- `agently-cli` OAuth 管理命令；
- stdio transport。

## 为什么需要这个项目

把 `agently-cli` 直接安装到每个 Agent 容器里会比较脆弱。容器重建可能同时清掉 CLI 和授权状态。

这个项目把这些职责拆开：

```text
Agent runtime
  -> MCP tools
     -> QQ Agent Mail MCP
        -> agently-cli
           -> QQ Agent Mail
```

Agent 运行时只使用 MCP。这个 Server 负责 CLI 二进制、OAuth 状态、邮件操作和附件下载。

## 设计原则

- 薄封装：MCP tools 尽量直接映射到 `agently-cli` 命令，少放业务逻辑。
- 结构化透传：尽量保留 CLI 的 JSON 输出结构。
- 不暴露任意命令执行：提供具名 tools，而不是一个原始 `args` 执行器。
- 小内存占用：优先简单实现，不引入不必要的后台 worker。
- 显式发送确认：保留 `agently-cli` 的两阶段发送确认流程。
- 易部署：优先支持容器 sidecar；如果目标运行时要求，再支持 stdio 回退形态。

## 已实现工具

第一版工具贴近稳定的 CLI 操作：

| MCP tool | 底层命令 |
| --- | --- |
| `agently_me` | `agently-cli +me` |
| `agently_message_list` | `agently-cli message +list` |
| `agently_message_read` | `agently-cli message +read` |
| `agently_message_search` | `agently-cli message +search` |
| `agently_message_send` | `agently-cli message +send` |
| `agently_message_reply` | `agently-cli message +reply` |
| `agently_message_forward` | `agently-cli message +forward` |
| `agently_attachment_upload` | `agently-cli attachment +upload` |
| `agently_attachment_download` | `agently-cli attachment +download` |

工具名刻意贴近 CLI。这样 Server 更容易审计，也避免把具体应用工作流写进通用邮箱桥接层。

## 快速开始

本地测试：

```bash
go test ./...
```

构建二进制。版本号从最新 git tag 自动注入 —— 用 `make build` 最省事：

```bash
make build
# 或者不用 make：
go build -ldflags "-X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 0.0.3)" \
  -o bin/qq-agent-mail-mcp ./cmd/qq-agent-mail-mcp
```

启动 StreamableHTTP Server：

```bash
QQ_AGENTLY_CLI_BIN=agently-cli \
QQ_AGENT_MAIL_MCP_BIND=127.0.0.1:8765 \
go run ./cmd/qq-agent-mail-mcp
```

常用环境变量：

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `QQ_AGENTLY_CLI_BIN` | `agently-cli` | `agently-cli` 可执行文件路径。 |
| `QQ_AGENT_MAIL_MCP_BIND` | `127.0.0.1:8765` | StreamableHTTP 监听地址。 |
| `QQ_AGENT_MAIL_MCP_ENV_FILE` | `./.env` | 启动时加载的 `.env` 文件路径（见下文）。 |

环境变量也可以通过 `.env` 文件提供。启动时 server 会从工作目录加载 `.env`（或 `QQ_AGENT_MAIL_MCP_ENV_FILE` 指定的路径），且不会覆盖环境中已存在的变量 —— 真实环境变量始终优先。文件不存在则静默跳过。模板见 [`.env.example`](.env.example)；本地运行时复制为 `.env`（`.env` 已被 gitignore）。

## 透传边界

Server 应该透传受支持的命令参数，而不是透传任意 shell 参数。

推荐：

```json
{
  "tool": "agently_message_list",
  "arguments": {
    "limit": 10,
    "dir": "inbox",
    "has_attachments": true
  }
}
```

避免：

```json
{
  "tool": "agently_cli",
  "arguments": {
    "args": ["message", "+list", "--limit", "10"]
  }
}
```

结构化形式更安全、更容易被 Agent 发现和使用，也更符合 MCP tool schema 的表达方式。

## 错误返回

如果 `agently-cli` 非零退出，Server 会区分两类情况：

- 如果 stdout 是 JSON：把它解析后放进 MCP tool error 的 `structuredContent.error.stdout`；
- 如果 stdout 不是 JSON：保留脱敏后的 `stdout_text` 和 `stderr`。

例如未登录、消息不存在、参数错误等 CLI 级错误，会作为 `isError=true` 的 tool result 返回，而不是 MCP 协议级错误。这样上层 Agent 可以读取：

```json
{
  "ok": false,
  "error": {
    "type": "agently_cli_error",
    "code": "auth",
    "message": "login required",
    "exit_code": 7,
    "stdout": {
      "ok": false,
      "error": {
        "type": "auth",
        "message": "login required"
      }
    }
  }
}
```

## 发送行为

`agently-cli message +send` 使用两阶段确认流程：

1. 第一次调用返回 confirmation token 和邮件摘要。
2. 调用方再次发送并带上 `--confirmation-token`，才真正完成投递。

MCP Server 应该保留这个行为：

- 如果没有传入 confirmation token，就返回确认响应；
- 如果传入了 confirmation token，就完成发送请求；
- 通用工具里不添加隐藏的自动确认逻辑。

具体应用的白名单策略可以由调用方实现，或者未来通过明确命名的专用工具叠加在通用工具之上。

## 认证模型

`agently-cli auth login` 会执行 OAuth 授权，并把凭据保存在系统凭据存储中。

MCP Server 不应该自己实现 QQ 邮箱 OAuth。账号绑定应该是一个本地管理动作：

```bash
qq-agent-mail-mcp auth login
qq-agent-mail-mcp auth status
qq-agent-mail-mcp auth refresh
```

这些命令预期在内部调用对应的 `agently-cli` 操作。它们是管理命令，不是暴露给 Agent 的 MCP tools。

推荐模型是：一个 MCP Server 实例绑定一个 QQ Agent 邮箱账号。如果需要多个账号，就运行多个 MCP 实例。

Server 启动时应该调用 `agently-cli +me`，确认 CLI 已授权。如果配置了 `QQ_AGENT_MAIL_EXPECTED_ACCOUNT`，还应该校验当前授权账号是否匹配；如果不匹配，应直接启动失败。

凭据持久化已使用项目自带的 `docker-compose.yml` 在真实 Linux 宿主机上验证通过。

`agently-cli` 把真正的 OAuth token 存在**文件型 keychain**里，既不在系统 keychain（精简运行时镜像里没有），也不在 `AGENTLY_CLI_CONFIG_DIR`（后者只放 `config.json`）。token 实际落在：

```text
$HOME/.local/share/agently-cli/master.key          # 加密密钥
$HOME/.local/share/agently-cli/bootstrap_token.enc # 加密后的 OAuth token
```

因此持久化的关键是把**这个目录**挂成 volume。compose 文件把 `agently-keyring` 挂到 `/home/app/.local/share/agently-cli`；这样挂载后，`docker compose down` / `up` 之后再跑 `agently-cli +me` 仍然成功。

部署有任何改动后，重新验证持久化：

```bash
agently-cli auth login
docker compose down && docker compose up -d
agently-cli +me            # 不重新登录仍授权成功 => 已持久化
```

## MCP 访问控制

这里有两层不同的信任边界：

```text
agently-cli -> QQ Agent Mail
MCP client  -> QQ Agent Mail MCP
```

第一层由 `agently-cli` 负责。第二层在 MCP Server 监听网络端口时，应该由这个 Server 自己负责。

推荐模式：

| 模式 | 使用场景 | 行为 |
| --- | --- | --- |
| `stdio` | Agent 本地启动 MCP 命令 | 不监听网络端口，不额外做 MCP 认证。 |
| `none` | 仅 localhost 开发或隔离测试网络 | 不要求 MCP auth token。 |
| `bearer` | HTTP/SSE 或任何远程 MCP transport | 每个请求都要求 bearer token。 |

推荐默认值：

- stdio transport：`QQ_AGENT_MAIL_MCP_AUTH_MODE=stdio`；
- HTTP/SSE transport：`QQ_AGENT_MAIL_MCP_AUTH_MODE=bearer`；
- `none` 必须显式配置，并且只建议用于本地测试。

示例：

```env
QQ_AGENT_MAIL_MCP_TRANSPORT=http
QQ_AGENT_MAIL_MCP_BIND=127.0.0.1:8765
QQ_AGENT_MAIL_MCP_AUTH_MODE=bearer
QQ_AGENT_MAIL_MCP_AUTH_TOKEN_FILE=/run/secrets/qq-agent-mail-mcp-token
QQ_AGENT_MAIL_EXPECTED_ACCOUNT=example@agent.qq.com
```

## 部署形态

### Sidecar MCP Server

如果 Agent 运行时可以连接远程 MCP Server，优先使用这种形态：

```text
agent-container
  -> remote MCP

qq-agent-mail-mcp
  -> agently-cli
  -> persistent credential volume
```

### 本地 stdio MCP 命令

如果 Agent 运行时只支持启动本地 MCP 命令，则使用回退形态：

```text
agent-container
  -> stdio MCP command
     -> agently-cli
     -> mounted credential volume
```

Sidecar 形态可以让 Agent 容器更小。stdio 形态则兼容那些不支持远程 MCP 连接的运行时。

## 推荐宿主机目录

建议把 MCP Server 和具体 Agent 应用目录分开。

推荐目录：

```text
/opt/qq-agent-mail-mcp/        # 应用文件、二进制、compose 文件
/etc/qq-agent-mail-mcp/        # 环境变量文件和本地配置
/var/lib/qq-agent-mail-mcp/    # 持久化认证状态和运行数据
/var/cache/qq-agent-mail-mcp/  # 临时附件下载和缓存
```

如果另一个 Agent 已经安装在 `/opt/<agent-name>` 下，这个项目通常应该作为兄弟目录存在，而不是放进它里面：

```text
/opt/<agent-name>
/opt/qq-agent-mail-mcp
```

这样可以让邮箱桥接层保持可复用，也避免它的 OAuth 状态或部署生命周期被某个具体 Agent 应用绑定。

## 安全注意事项

- 不暴露原始 shell 或任意 CLI 参数执行器。
- 不记录 OAuth token、refresh token、邮件正文或附件内容。
- 附件下载到受控工作目录。
- 拒绝输出路径中的路径穿越。
- 返回结构化错误，但如果原始 CLI 输出可能包含敏感信息，不要直接透出完整内容。
- 应用专用邮件策略不要塞进通用工具；如果确实需要，后续通过明确命名的专用工具增加。

## 实现语言

Server 将使用 Go 实现。

选择 Go 的原因：

- 内存占用小；
- 单二进制部署简单；
- 很适合做轻量 CLI wrapper；
- 适合作为小型 sidecar 服务长期运行。

之前考虑过 TypeScript，因为 MCP 的 TypeScript SDK 较成熟，且 `agently-cli` 通过 npm 分发；但这个项目更看重较小的运行时占用和简单的单二进制部署。

## 开发环境

本地需要准备：

- Go，建议使用当前稳定版；
- `agently-cli`，仅用于烟测和运行时集成验证；
- 本地开发机不要求安装 Docker。

macOS + Homebrew：

```bash
brew install go
go version
```

开发机保持干净。Docker 或 Linux 容器中的凭据持久化烟测应该放到目标服务器，或者一次性的 Linux 测试环境中完成，不要求在本地开发机安装 Docker。

第一版实现前，还应该先在部署宿主机上做一次 `agently-cli` 凭据持久化烟测，再决定是否使用 sidecar 部署。

## 仓库卫生

公开仓库文件只描述通用 MCP 邮箱桥接能力。

本地应用笔记、私有工作流、凭据、下载附件和烟测产物都不要进入 git。

忽略路径见 `.gitignore`。
