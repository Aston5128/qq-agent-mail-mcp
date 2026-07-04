# 目录结构

本文档描述 QQ Agent Mail MCP 项目的目录组织方式。

## 顶层结构

```text
qq-agent-mail-mcp/
├── .agents/                    # Agent 配置和文档
│   └── core/                   # 核心配置
│       └── directory-structure.md  # 本文件
├── bin/                        # 编译产物
│   └── qq-agent-mail-mcp      # 编译后的二进制文件
├── cmd/                        # 可执行入口
│   └── qq-agent-mail-mcp/
│       ├── main.go             # 程序入口
│       └── main_test.go        # 入口测试
├── docs/                       # 项目文档
│   ├── authentication/         # 鉴权与凭证持久化
│   │   ├── authentication.md
│   │   └── authentication.zh-CN.md
│   ├── changelogs/             # 非英文 changelog
│   │   └── CHANGELOG.zh-CN.md  # 中文 changelog
│   ├── deployment/             # 部署说明
│   │   ├── deployment.md
│   │   └── deployment.zh-CN.md
│   ├── design/                 # 设计说明
│   │   ├── design.md
│   │   └── design.zh-CN.md
│   ├── development/            # 开发说明
│   │   ├── development.md
│   │   └── development.zh-CN.md
│   ├── errors/                 # 错误响应说明
│   │   ├── errors.md
│   │   └── errors.zh-CN.md
│   ├── readme/                 # 非英文 README
│   │   └── README.zh-CN.md     # 中文 README
│   ├── security/               # 安全说明
│   │   ├── security.md
│   │   └── security.zh-CN.md
│   ├── superpowers/            # 高级功能规划
│   │   └── plans/
│   │       ├── 2026-07-03-streamable-http-poc.md
│   │       └── 2026-07-04-structured-cli-errors-0.0.2.md
│   └── tools/                  # MCP 工具说明
│       ├── tools.md
│       └── tools.zh-CN.md
├── internal/                   # 内部实现（不对外暴露）
│   ├── agently/                # agently-cli 封装层
│   │   ├── runner.go           # CLI 命令执行器
│   │   ├── runner_test.go      # 执行器测试
│   │   ├── tools.go            # MCP 工具定义
│   │   └── tools_test.go       # 工具测试
│   ├── config/                 # 运行时配置加载
│   │   ├── config.go           # 环境变量 / .env 加载
│   │   └── config_test.go      # 配置加载测试
│   └── server/                 # MCP Server 实现
│       ├── server.go           # Server 核心逻辑 + StreamableHTTP
│       └── streamable_http_test.go  # StreamableHTTP 测试
├── .github/                    # GitHub 配置（CI/CD）
│   └── workflows/
│       └── publish-ghcr.yml    # Release 时构建并推送 GHCR 镜像
├── .dockerignore               # Docker 构建忽略规则
├── .env.example                # 环境变量配置示例
├── .gitignore                  # Git 忽略规则
├── Dockerfile                  # 多阶段镜像构建（MCP server + agently-cli）
├── Makefile                    # build/run/compose-build/test/vet/clean 目标
├── docker-compose.yml          # 容器编排：源码构建（凭证卷 + 端口 + env）
├── docker-compose.ghcr.yml     # 容器编排：拉取 GHCR 已发布镜像
├── go.mod                      # Go 模块定义
├── go.sum                      # Go 依赖校验
├── README.md                   # 英文 README（canonical）
└── CHANGELOG.md                # 英文 changelog（canonical）
```

## 目录说明

### `.agents/`

存放 Agent 相关的配置和文档。`core/` 子目录存放核心配置文件，包括目录结构说明等。

### `bin/`

编译产物目录。存放 `go build` 生成的可执行文件。该目录已加入 `.gitignore`，不会提交到版本控制。

### `cmd/`

可执行程序入口。遵循 Go 项目的标准布局，每个子目录对应一个可执行程序。

- `main.go`：程序入口，负责解析配置、初始化依赖、启动 Server
- `main_test.go`：入口级别的集成测试

### `docs/`

项目文档目录。每个主题目录同时提供英文版与 `zh-CN` 版本。

- `authentication/`：鉴权与凭证持久化说明
- `changelogs/`：非英文 changelog（如 `CHANGELOG.zh-CN.md`）
- `deployment/`：部署说明
- `design/`：设计说明（MCP 桥接层的设计原则与工具映射）
- `development/`：开发说明
- `errors/`：错误响应说明
- `readme/`：非英文 README（如 `README.zh-CN.md`）
- `security/`：安全说明
- `superpowers/plans/`：高级功能规划和 PoC 方案
- `tools/`：MCP 工具说明

### `internal/`

内部实现包，Go 的 `internal` 机制确保这些包不能被外部项目导入。

#### `internal/agently/`

封装与 `agently-cli` 的交互逻辑。

- `runner.go`：CLI 命令执行器，负责调用 `agently-cli` 并解析输出
- `tools.go`：MCP 工具定义，将 MCP 工具映射到 CLI 命令

#### `internal/config/`

运行时配置加载。从环境变量（及可选 `.env` 文件）读取监听地址、agently-cli 路径、MCP 鉴权等参数；真实环境变量优先于 `.env`。

- `config.go`：配置结构与环境变量绑定
- `config_test.go`：配置加载测试

#### `internal/server/`

MCP Server 的核心实现。

- `server.go`：Server 核心逻辑，包括工具注册、请求处理
- StreamableHTTP transport 实现

### 根目录文件

- `go.mod` / `go.sum`：Go 模块依赖管理
- `Makefile`：`build` / `run` / `compose-build` / `test` / `vet` / `clean` 目标；版本号通过 ldflags 从 git tag 注入，无 tag 时回退 `dev`
- `README.md`：英文 README（canonical，其他语言见 `docs/readme/`）
- `CHANGELOG.md`：英文 changelog（canonical，其他语言见 `docs/changelogs/`）
- `.gitignore`：Git 忽略规则，包括 `bin/`、`.env`、本地配置等
- `.env.example`：环境变量配置示例（`.env` 已 gitignore）
- `Dockerfile`：多阶段镜像构建 —— Go 编译 MCP server，运行时层预装 `agently-cli`
- `docker-compose.yml`：容器编排（源码构建）—— `AGENTLY_CLI_CONFIG_DIR` 凭证卷、端口映射、可选 `.env`，镜像 tag 由 `QQ_AGENT_MAIL_MCP_BUILD_VERSION` 控制（默认 `dev`）
- `docker-compose.ghcr.yml`：容器编排（GHCR 拉取）—— 拉取 `ghcr.io/aston5128/qq-agent-mail-mcp`，镜像 tag 由 `QQ_AGENT_MAIL_MCP_VERSION` 控制（默认 `latest`）
- `.dockerignore`：Docker 构建上下文忽略规则
- `.github/workflows/publish-ghcr.yml`：Release 时构建并推送 GHCR 镜像（tag 为 release tag + `latest`）

## 设计原则

1. **标准 Go 布局**：遵循 Go 社区推荐的项目结构（`cmd/` + `internal/`）
2. **内部实现隐藏**：所有实现代码放在 `internal/` 下，不对外暴露
3. **文档集中管理**：英文 `README.md` / `CHANGELOG.md` 作为 canonical 放根目录（工具默认在此查找），其他语言版本放 `docs/readme/` 与 `docs/changelogs/`
4. **编译产物分离**：`bin/` 目录独立，便于清理和忽略
