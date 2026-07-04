# 开发

[English](development.md)

## 要求

- Go 1.26 或更新版本。
- 只有在本地构建/运行容器镜像时才需要 Docker。
- 只有真实集成烟测需要 `agently-cli`。

普通 Go 测试不要求本地开发机安装 Docker。

## 常用命令

```bash
go test ./...
go vet ./...
make build
make run
make compose-build
```

不使用 make 构建：

```bash
go build -ldflags "-X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" \
  -o bin/qq-agent-mail-mcp ./cmd/qq-agent-mail-mcp
```

## 测试

`internal/server` 包含一个使用 `httptest` 的 StreamableHTTP 集成测试。它需要允许监听一个临时本地端口。

## 版本

二进制版本通过以下方式注入：

```text
-ldflags "-X main.version=<version>"
```

`make build` 会优先使用最新 git tag，没有 tag 时回退到 `dev`。Release/GHCR 构建会通过 Docker build arg `VERSION` 注入 release tag。不要在 `main.go` 中手动维护发布版本号。

直接执行 `go build ./cmd/qq-agent-mail-mcp` 或直接执行 `docker compose up -d --build` 时，如果没有注入版本，二进制版本会是 `dev`。需要可追踪版本时，使用 `make build`、`make compose-build` 或 GHCR release workflow。

手动使用源码构建 compose 部署时，可以这样注入版本：

```bash
QQ_AGENT_MAIL_MCP_BUILD_VERSION="$(git describe --tags --always --dirty 2>/dev/null || echo dev)" \
  docker compose up -d --build
```

## Docker 镜像

```bash
docker compose build
docker compose up -d
```

运行时镜像使用：

- Debian slim / glibc，而不是 Alpine；
- Node.js，用于 npm 分发的 `agently-cli` wrapper；
- 非 root 的 `app` 用户；
- `tini` 作为 PID 1。

## git 提交要求

私有应用笔记、本地工作流、凭据、下载附件和烟测产物都不要进入 git。

项目专用笔记可放在 `private/` 或其它已忽略的本地路径。
