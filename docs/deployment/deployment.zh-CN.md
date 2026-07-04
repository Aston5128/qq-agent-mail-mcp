# 部署

[English](deployment.md)

当前推荐使用 Docker Compose 部署。优先路径是使用 `docker-compose.ghcr.yml` 拉取已发布的 GitHub Container Registry 镜像；如果希望部署机直接从源码构建，则使用默认 `docker-compose.yml`。

因为当前没有内置认证，推荐部署在 Hermes 同宿主机或其它可信私有环境中。
镜像会把 Go MCP server 和 `agently-cli` 打包到同一个轻量运行时容器里。

## 使用已发布镜像启动

```bash
docker compose -f docker-compose.ghcr.yml pull
docker compose -f docker-compose.ghcr.yml up -d
```

然后授权 CLI：

```bash
docker compose -f docker-compose.ghcr.yml exec qq-agent-mail-mcp agently-cli auth login
docker compose -f docker-compose.ghcr.yml exec qq-agent-mail-mcp agently-cli +me
```

如果要固定某个镜像 tag，设置：

```env
QQ_AGENT_MAIL_MCP_VERSION=v0.0.5
```

## 从源码构建

如果希望本地镜像 tag 和二进制版本可追踪，使用：

```bash
make compose-build
```

或者手动注入版本：

```bash
QQ_AGENT_MAIL_MCP_BUILD_VERSION="$(git describe --tags --always --dirty 2>/dev/null || echo dev)" \
  docker compose up -d --build
```

如果直接运行 `docker compose up -d --build`，本地镜像 tag 和二进制版本默认是 `dev`。

然后授权 CLI：

```bash
docker compose exec qq-agent-mail-mcp agently-cli auth login
docker compose exec qq-agent-mail-mcp agently-cli +me
```

默认 endpoint：

```text
http://<host>:8765
```

## 端口

compose 文件默认发布：

```yaml
ports:
  - "8765:8765"
```

如果只想允许宿主机访问，改成：

```yaml
ports:
  - "127.0.0.1:8765:8765"
```

## Volumes

项目挂载两个 volume：

```yaml
volumes:
  - agently-data:/var/lib/agently
  - agently-keyring:/home/app/.local/share/agently-cli
```

`agently-data` 存放 `config.json`。`agently-keyring` 存放加密后的 OAuth token 和密钥材料。两个都需要，否则凭据持久化不可靠。

细节见 [认证与凭据持久化](../authentication/authentication.zh-CN.md)。

## 环境变量

server 读取普通环境变量，也可以加载 `.env` 文件。真实环境变量优先级高于 `.env`。

常用变量：

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `QQ_AGENT_MAIL_MCP_BIND` | `127.0.0.1:8765` | StreamableHTTP 监听地址。 |
| `QQ_AGENTLY_CLI_BIN` | `agently-cli` | `agently-cli` 路径。 |
| `QQ_AGENT_MAIL_MCP_ENV_FILE` | `./.env` | 可选 env 文件路径。 |

Docker 内部会把 `QQ_AGENT_MAIL_MCP_BIND` 设置为 `0.0.0.0:8765`。

## 推荐宿主机目录

手动部署时建议：

```text
/opt/qq-agent-mail-mcp/        # 应用文件、二进制、compose 文件
/etc/qq-agent-mail-mcp/        # 环境变量文件和本地配置
/var/lib/qq-agent-mail-mcp/    # 持久化认证状态和运行数据
/var/cache/qq-agent-mail-mcp/  # 临时下载和缓存
```

如果另一个 Agent 已经安装在 `/opt/<agent-name>` 下，这个服务应作为兄弟目录存在，而不是放进那个 Agent 应用目录里面。

## 升级

使用已发布镜像：

```bash
docker compose -f docker-compose.ghcr.yml pull
docker compose -f docker-compose.ghcr.yml up -d
docker compose -f docker-compose.ghcr.yml exec qq-agent-mail-mcp agently-cli +me
```

使用本地源码构建：

```bash
docker compose build --pull
docker compose up -d
docker compose exec qq-agent-mail-mcp agently-cli +me
```

如果 `+me` 仍然成功，说明升级后凭据仍然保留。
