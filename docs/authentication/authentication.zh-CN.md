# 认证与凭据持久化

[English](authentication.md)

QQ Agent Mail MCP 不自己实现 QQ 邮箱 OAuth。OAuth 授权由 `agently-cli` 负责。

## 绑定账号

在 MCP server 实际运行的同一环境中执行登录：

```bash
agently-cli auth login
agently-cli +me
```

使用 Docker Compose 时：

```bash
docker compose exec qq-agent-mail-mcp agently-cli auth login
docker compose exec qq-agent-mail-mcp agently-cli +me
```

推荐模型是一个 MCP server 实例绑定一个 QQ Agent Mail 账号。如果需要多个账号，就运行多个实例。

## 需要持久化什么

在已验证的容器运行环境中，`agently-cli` 会持久化两类数据：

```text
/var/lib/agently                         # config.json (AGENTLY_CLI_CONFIG_DIR)
/home/app/.local/share/agently-cli       # master.key + bootstrap_token.enc
```

真正的 OAuth token 存在 `$HOME/.local/share/agently-cli/` 下的文件型 keychain 中，而不是 `AGENTLY_CLI_CONFIG_DIR`。

因此项目自带 compose 文件挂载了两个命名卷：

```yaml
volumes:
  - agently-data:/var/lib/agently
  - agently-keyring:/home/app/.local/share/agently-cli
```

两个都需要。如果缺少 keychain volume，容器重建后 `agently-cli +me` 可能失效。

## 验证持久化

部署或 volume 配置变更后，重新验证：

```bash
docker compose exec qq-agent-mail-mcp agently-cli auth login
docker compose exec qq-agent-mail-mcp agently-cli +me
docker compose down
docker compose up -d
docker compose exec qq-agent-mail-mcp agently-cli +me
```

如果最后一次 `+me` 不需要重新登录仍然成功，说明凭据已经持久化。

## MCP 认证

这里有两层信任边界：

```text
agently-cli -> QQ Agent Mail
MCP client  -> QQ Agent Mail MCP
```

第一层由 `agently-cli` 处理。当前 server 版本尚未实现 MCP 请求认证。在 bearer-token 认证实现前，请只把 server 放在 localhost、私有网络中，或通过可信反向代理 / 防火墙保护。

## 尚未实现

- `qq-agent-mail-mcp auth login`
- `qq-agent-mail-mcp auth status`
- `qq-agent-mail-mcp auth refresh`
- 启动时账号 fail-fast 校验
- MCP bearer-token 认证
