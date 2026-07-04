# 更新日志

> English: [CHANGELOG.md](../../CHANGELOG.md)

## 0.0.3

- 新增容器化部署：多阶段 Dockerfile（MCP server 构建 + agently-cli 运行时）与 docker-compose。
- 通过 `AGENTLY_CLI_CONFIG_DIR` 命名卷持久化 agently-cli 凭证，容器重启后保持登录。
- 以非 root 用户运行，`tini` 作为 PID 1，保证信号正确转发与子进程回收。
- 将 agently-cli 锁定到已验证的 1.0.6 版本（可通过 build arg 覆盖）。
- 新增 `.dockerignore`。

## 0.0.2

- 将 `agently-cli` 非零退出时的 stdout JSON 作为结构化 MCP 工具错误返回。
- 保留 CLI 退出码、stderr、归一化的 code/message 以及原始 stdout JSON。
- 将 `error.code` 与真实鉴权形状 `error.type` 同时归一化到顶层 `error.code`。
- 为 StreamableHTTP 上未登录（unauthorized）类 CLI 失败增加回归测试覆盖。

## 0.0.1

- 初始 StreamableHTTP PoC。
- 暴露具名 MCP 工具用于 QQ Agent Mail 操作。
- 将带类型的 MCP 入参转发给 `agently-cli` 并返回解析后的 stdout JSON。
