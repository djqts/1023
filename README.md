# Banner 指纹识别系统

接收一批网络扫描原始数据（`ip`、`port`、`banner`），输出协议、产品、版本、OS 线索和置信度。交付形态为 Go 的 client + server，Docker Compose 一键启动。

## 快速启动

```bash
cp .env.example .env
# 按需修改 .env 里的本地密码

docker compose up --build
```

启动链：

```text
mysql healthy -> server healthy -> client 执行
redis 并行启动，不阻塞 server
```

- server: `http://127.0.0.1:8080`
- health: `http://127.0.0.1:8080/health`
- 只发布 server 8080 到宿主回环；MySQL/Redis 不发布端口。
- 默认 profile 为 P0-B（MySQL + Redis 旁路）。
- 需要受限网络构建时，可在 `.env` 中覆盖 `GO_IMAGE` / `RUNTIME_IMAGE`。

重复执行 client：

```bash
docker compose run --rm client
```

停止：

```bash
docker compose down
```

### 受限环境（Docker Desktop buildx / 网络）

如果 `docker compose up --build` 在 Docker Desktop 上报 buildx 会话错误，或无法拉取官方基础镜像，使用本机 Docker 直连构建的等价脚本：

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\up.ps1
```

脚本会优先使用本机已有的 golang/distroless 镜像，构建 server/client 后执行 `docker compose up --no-build` 并运行 client。可通过 `GO_IMAGE` / `RUNTIME_IMAGE` 环境变量覆盖基础镜像，也可用 `-SkipClient` 只启动服务。

无宿主入口（airgap）编排：

```bash
docker compose -f docker-compose.yml -f compose.airgap.yaml up -d --no-build
```

`compose.airgap.yaml` 会清空 server 的端口发布，client 仍走 `internal` 网络访问 `server:8080`。

## P1 能力

| 能力 | 实现 |
|---|---|
| 二进制输入 | `banner_b64` + 非标准 `\xNN` JSON 转义兼容 |
| 客户端健壮性 | 自动分片（`-batch`）、5xx/429/超时重试（`-retries`）、部分成功输出并退出 3 |
| 规则运维 | 启动时校验规则 examples；Linux 下 `SIGHUP` 原子热加载，坏规则保留旧快照 |
| 协议扩展 | SMTP（Postfix/Exim/Exchange）、FileZilla Server、IIS、Pure-FTPd、dropbear 等 |
| 部署 | 多架构构建参数（`TARGETOS/TARGETARCH`）、airgap overlay |
| 性能工程 | `internal/engine` benchmark；参考约 3.5µs/op（单条、6 个典型 banner 轮询） |
| CI | `.github/workflows/ci.yml`：test / vet / build / compose config |

规则热加载（Linux 容器）：

```bash
docker kill -s HUP banner-fingerprint-server-1
```

## API

### `POST /fingerprint`

兼容裸数组与 `{"items":[...]}` 两种请求体，响应为裸数组：

```json
[
  {"ip":"1.2.3.4","port":22,"banner":"SSH-2.0-OpenSSH_8.9p1 Ubuntu-3"}
]
```

```json
[
  {
    "ip":"1.2.3.4",
    "port":22,
    "protocol":"SSH",
    "product":"OpenSSH",
    "version":"8.9p1",
    "os_hint":"Ubuntu",
    "confidence":0.95
  }
]
```

认不出时：

```json
{"ip":"1.2.3.23","port":12345,"protocol":"unknown","product":"","version":"","os_hint":"","confidence":0}
```

协议值按自测示例使用大写：`SSH`、`HTTP`、`MySQL`、`Redis`、`FTP`、`SMTP`；未知为小写 `unknown`。

### `GET /health`

返回服务状态、版本、规则数量/摘要、依赖状态和运行时长。规则已加载但 MySQL/Redis 故障时返回 `200 degraded`；规则未加载/停机返回 `503`。

## 本地运行（不经过 Docker）

```bash
go run ./cmd/server
go run ./cmd/client -server http://127.0.0.1:8080 -input data/input.json
```

JSON 输出：

```bash
go run ./cmd/client -server http://127.0.0.1:8080 -input data/input.json -json
```

测试：

```bash
go test ./...
go vet ./...
go test ./internal/engine -run '^$' -bench BenchmarkFingerprint -benchtime=100x
```

## 规则

规则位于 `rules/*.yaml`，由 server 启动时加载、校验、编译；代码只做通用匹配。新增协议/产品不需要改 Go 代码：

```yaml
- id: http-nginx
  protocol: HTTP
  product: nginx
  priority: 100
  confidence: 0.9
  patterns:
    - '(?im)^Server:\s*nginx/(?P<version>[0-9][^\s]*)'
  os_hint_patterns:
    - '(?i)\((Ubuntu|Debian|CentOS|FreeBSD)\)'
```

规则文件通过 Compose 只读挂载到 `/app/rules`；镜像内不包含规则。Linux 下发送 `SIGHUP` 会重新加载并原子替换快照；新规则编译失败时保留旧快照。

## 目录

```text
cmd/server/              server 入口与 healthcheck 子命令
cmd/client/              client CLI
internal/engine/         规则加载/校验/匹配
internal/httpapi/        Gin 路由、JSON 解析、限额、健康检查
internal/domain/         领域类型与端口
internal/app/            应用层端口/命令/结果
internal/infra/          MySQL/Redis/RabbitMQ 适配骨架
internal/platform/       配置、错误码、日志、ID/时钟
pkg/api/                 共享 HTTP DTO 与容错 JSON 解码
pkg/client/              HTTP client 实现
rules/                   YAML 识别规则
deploy/                  Dockerfile、MySQL 初始化 SQL
scripts/up.ps1           受限环境直连构建 + 启动脚本
compose.airgap.yaml      无宿主端口覆盖
.github/workflows/ci.yml CI
data/input.json          自测输入
docs/                    需求文档、架构设计
```

更多设计见 [需求文档](docs/REQUIREMENTS.md) 与 [架构设计](docs/ARCHITECTURE.md)。

## 当前实现边界

- P0 同步识别链路已可运行；P1 的客户端分片/重试、规则热加载、examples 校验、SMTP/FileZilla 扩展、airgap、多架构参数、benchmark、CI 已落地。
- RabbitMQ 异步任务与 gRPC 仍是可选架构扩展，不进入默认 Compose；P0/P1 验收不依赖它们。
- MySQL 审计与 Redis 健康探测为旁路：连接失败只降级，不阻断识别。
- 规则匹配覆盖 SSH、HTTP（nginx/Apache/Jetty/Microsoft-IIS）、MySQL、Redis、FTP（ProFTPD/vsFTPd/Pure-FTPd/FileZilla/Microsoft FTP）、SMTP（Postfix/Exim/Exchange）。
- 二进制 banner 支持 `banner_b64`；也兼容 `\xNN` 形式的非标准 JSON 转义。
