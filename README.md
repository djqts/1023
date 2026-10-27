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

协议值按自测示例使用大写：`SSH`、`HTTP`、`MySQL`、`Redis`、`FTP`；未知为小写 `unknown`。

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

规则文件通过 Compose 只读挂载到 `/app/rules`；镜像内不包含规则。

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
deploy/                  Dockerfile、MySQL 迁移
data/input.json          自测输入
docs/                    需求文档、架构设计
```

更多设计见 [需求文档](docs/REQUIREMENTS.md) 与 [架构设计](docs/ARCHITECTURE.md)。

## 当前实现边界

- P0 同步识别链路已可运行；RabbitMQ/gRPC 为 P1 骨架/可选。
- MySQL 审计与 Redis 健康探测为旁路：连接失败只降级，不阻断识别。
- 规则匹配覆盖 SSH、HTTP（nginx/Apache/Jetty/Microsoft-IIS）、MySQL、Redis、FTP（ProFTPD/vsFTPd/Pure-FTPd/Microsoft FTP）。
- 二进制 banner 支持 `banner_b64`；也兼容 `\xNN` 形式的非标准 JSON 转义。
