# QQ Agent Mail MCP

[English](../../README.md) | **简体中文**

QQ Agent Mail MCP 是一个轻量级 StreamableHTTP MCP Server，通过 [QQ Agent 邮箱](https://agent.qq.com) 的 `agently-cli` 暴露 QQ Agent 邮箱能力。

基于 `agently-cli` 本身已有的能力封装，将

它把选定的 `agently-cli` 邮件操作包装成结构化 MCP tools，同时把 CLI 安装和 OAuth 凭据状态从具体 Agent 运行环境中隔离出来。

## 功能

- Go 编写的 StreamableHTTP MCP Server。
- 对 `agently-cli` 做薄封装和结构化透传。
- 将 CLI stdout JSON 解析为 MCP structured content。
- 将 CLI 非零退出返回为结构化 MCP tool error。
- 支持 Docker 部署，并持久化 `agently-cli` OAuth 凭据。
- 不暴露原始 shell 或任意 CLI 参数执行器。

当前限制：

- 尚未实现 MCP bearer-token 认证。
- 尚未实现 stdio transport。
- 尚未由本 server 暴露 `agently-cli` OAuth 管理命令。
- 刻意不暴露删除/回收站操作。

## 为什么需要这个项目

起因是因为我的 Hermes Agent 是通过 docker 容器化部署的，是可以将 `agently-cli` 直接安装到容器内，但是每次 ```docker compose down``` 都会丢失。也是可以通过编辑 Hermes Agent 的 docker-compose.yaml 避免，但是总觉得不够优雅，故写了一个 MCP server 来转发。

这个项目把职责拆开：

```text
Agent runtime
  -> MCP tools
     -> QQ Agent Mail MCP
        -> agently-cli
           -> QQ Agent Mail
```

## Docker Compose 快速部署

这里有两个 compose 文件：

```text
docker-compose.ghcr.yml   # 推荐：拉取已发布的 GHCR 镜像
docker-compose.yml        # 源码构建：在本机从源码构建镜像
```

### 推荐：拉取 GitHub Package 镜像

拉取最新的代码

```bash
git clone https://github.com/Aston5128/qq-agent-mail-mcp.git
cd qq-agent-mail-mcp
```

拉取并启动已发布镜像：

```bash
docker compose -f docker-compose.ghcr.yml pull
docker compose -f docker-compose.ghcr.yml up -d
```

在运行中的容器内授权 `agently-cli`：

```bash
# 运行这行之后会输出一个 url，复制到任意浏览器中使用微信进行授权
docker compose -f docker-compose.ghcr.yml exec qq-agent-mail-mcp agently-cli auth login
docker compose -f docker-compose.ghcr.yml exec qq-agent-mail-mcp agently-cli +me
```

默认 StreamableHTTP endpoint：

```text
http://<host>:8765
```

compose 文件会持久化 `agently-cli` 的 config 和 OAuth keychain 数据，因此授权状态可以在 `docker compose down` / `up` 后保留。

如果要固定某个镜像 tag，可以在 compose 文件旁边创建 `.env`：

```env
QQ_AGENT_MAIL_MCP_VERSION=v0.0.5
```

然后继续使用上面的 GHCR compose 命令。

### 可选：源码构建

如果希望部署机直接从当前源码树构建镜像：

```bash
make compose-build
```

两个 compose 文件使用相同的端口、volume、环境变量和容器名；区别只在于镜像是从 GHCR 拉取，还是从 `Dockerfile` 本地构建。

直接运行 `docker compose up -d --build` 仍然可用，但如果没有设置 `QQ_AGENT_MAIL_MCP_BUILD_VERSION`，本地镜像 tag 和二进制版本都会是 `dev`。

## MCP client 配置

将支持 MCP 的 Agent 配置为连接上面的远程 StreamableHTTP server。

在 MCP bearer-token 认证实现前，不要把这个 server 直接暴露到不可信网络。建议只绑定 localhost、放在私有网络中，或通过反向代理 / 防火墙保护。

## Tools

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

参数和行为说明见 [Tools](../tools/tools.zh-CN.md)。

## 本地开发

```bash
go test ./...
go vet ./...
make build
```

从源码运行：

```bash
QQ_AGENTLY_CLI_BIN=agently-cli \
QQ_AGENT_MAIL_MCP_BIND=127.0.0.1:8765 \
go run ./cmd/qq-agent-mail-mcp
```

## 文档

- [部署](../deployment/deployment.zh-CN.md)
- [认证与凭据持久化](../authentication/authentication.zh-CN.md)
- [Tools](../tools/tools.zh-CN.md)
- [错误返回](../errors/errors.zh-CN.md)
- [安全](../security/security.zh-CN.md)
- [设计说明](../design/design.zh-CN.md)
- [开发](../development/development.zh-CN.md)
- [更新日志](../changelogs/CHANGELOG.zh-CN.md)
