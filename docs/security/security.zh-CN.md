# 安全

[English](security.md)

这个项目把 Agent 运行时连接到一个邮箱账号。请把 MCP server 当作敏感服务对待。

## 网络暴露

当前尚未实现 MCP bearer-token 认证。在实现前，不要把 server 直接暴露到不可信网络。

推荐选项：

- 只绑定 `127.0.0.1`；
- 把 server 和 client 放在私有网络中；
- 通过反向代理、VPN、防火墙或等效控制保护访问。

## 工具面

server 只暴露具名 tools，不暴露：

- 原始 shell 执行；
- 任意 CLI 参数执行；
- 永久删除操作。

`agently_message_trash` 会转发 `agently-cli message +trash`，将邮件移入回收站，而不是永久删除。QQ Agent Mail 的回收站邮件会保留 30 天。

## 确认操作

发送、回复、转发、移入回收站都会保留 `agently-cli` 的 confirmation-token 流程。server 不会自动确认这些操作。

## Secrets 和日志

不要记录：

- OAuth token；
- refresh token；
- 邮件正文；
- 附件内容。

CLI stderr 和 stdout 错误文本在返回前会截断。

## 附件和路径

附件应该从受控目录读取并写入受控目录。除非 Agent 确实需要访问，否则不要把宽泛的宿主机路径挂进运行时容器。

## 应用专用策略

应用专用白名单、邮件处理规则和业务逻辑应放在这个通用 bridge 之外。如果未来需要专用行为，应通过明确命名的 tools 增加，而不是藏在通用透传工具里。
