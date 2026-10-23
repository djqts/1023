# 规则目录

Compose 会把本目录以只读方式挂载到 `/app/rules`。

实现阶段在此放置按协议分组的 YAML 规则包，例如：

- `10-ssh.yaml`
- `20-http.yaml`
- `30-mysql.yaml`
- `40-redis.yaml`
- `50-ftp.yaml`

规则文件由 server 启动时加载、校验、编译；镜像内不包含规则。
