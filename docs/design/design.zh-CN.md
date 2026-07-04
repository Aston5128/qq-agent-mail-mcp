# 设计说明

[English](design.md)

本文档描述 QQ Agent Mail MCP 的通用设计。它不包含任何具体应用的工作流。

## 目标

通过一个轻量、可检查的 `agently-cli` 封装层，把 QQ Agent Mail 能力暴露给支持 MCP 的 Agent。

Server 应该让 Agent 更方便地使用 CLI，而不是自己变成一个邮件自动化平台。

## 非目标

- 不暴露原始 shell 执行。
- 不提供通用的 `agently_cli(args)` 工具。
- 不包含具体应用的邮件处理逻辑。
- 不隐藏自动发送确认。
- 不长期存储邮件正文或附件。
- 当前版本不暴露删除/回收站能力。

## 结构化透传

封装层把类型化 MCP 参数翻译成 CLI flags。

不应该接受任意 CLI 参数数组，因为那样会：

- 让工具更难被 Agent 理解；
- 绕过校验；
- 扩大安全攻击面；
- 让兼容性更难维护。

## 输出形状

当 `agently-cli` 返回 JSON 时，server 会把解析后的 JSON 作为 MCP structured content 返回。

当 `agently-cli` 非零退出且 stdout 是 JSON 时，server 会保留这份 JSON，并作为 `isError=true` 的 MCP tool result 返回。auth、not found、invalid arguments 这类 CLI 级错误不应该变成 MCP 协议级错误。

## 发送确认

`agently-cli message +send`、`+reply`、`+forward` 都是两阶段操作：

1. 第一次调用返回确认载荷；
2. 第二次调用带上 confirmation token。

MCP server 保留这个行为。

## 删除 / 回收站

`agently-cli message +trash` 已存在，并且也使用 confirmation-token 流程，但当前 server 不暴露它。未来如果增加，应该在 CLI confirmation 之外再加 MCP 层显式确认。

## 运行时模型

当前实现目标是远程 StreamableHTTP sidecar。如果某些 client runtime 需要本地启动 MCP server，未来可以增加 stdio。

## 语言选择

项目使用 Go，因为这是一个轻量进程封装层，需要较小运行时占用和简单的二进制部署。
