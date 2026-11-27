# geoyuan SSH 准入测试服务

复刻 [pico.sh/getting-started](https://pico.sh/getting-started) 的入口模式：用户在终端执行 `ssh diliy.cn`，用 SSH 公钥连上服务后，在终端 TUI 里完成 30 道技术综合选择题，达到通过线即自动加入 `geoyuan` team。

服务基于 `golang.org/x/crypto/ssh` 自建（不依赖系统 sshd），TUI 基于 `bubbletea`，数据存本地 SQLite（`modernc.org/sqlite`，纯 Go 无 cgo）。

## 工作流程

1. 用户 `ssh diliy.cn` → 服务端 `PublicKeyCallback` **放行任意公钥**（含未知指纹），准入与否由答题结果决定，不在认证阶段拒绝。
2. 首次连接的公钥没有绑定用户 → TUI 提示输入用户名（校验唯一性后与公钥指纹绑定）。
3. 进入 30 道四选一测试：`↑/↓` 或 `j/k` 选择，`1-4` 直选并跳到下一题，回车确认。
4. 交卷后即时显示分数、对错回顾与解析。
   - **≥ 24/30（80%）** → `approved`，写入 team 名单，下次重连直接进入欢迎页。
   - 未通过 → `pending`，可随时重连重测。
5. host key 首次运行自动生成 ed25519 并持久化，重启复用（客户端 known_hosts 稳定）。

## 目录结构

```
ssh/
├── main.go                      # 入口：加载 host key、监听、每连接起 goroutine
├── go.mod / go.sum
├── .gitignore                   # 忽略 data/、*.db
├── internal/
│   ├── questions/questions.go   # 30 题题库（编程/OS/网络/安全）
│   ├── store/store.go           # SQLite 存储（users / attempts）
│   ├── tui/app.go               # bubbletea 答题界面
│   └── sshserver/server.go      # x/crypto/ssh 服务端：认证 + PTY + 路由 TUI
└── data/                        # 运行时生成（geoyuan.db + ssh_host_ed25519_key），已 gitignore
```

## 构建与运行

```bash
cd ssh
go build -o geoyuan-ssh .

# 开发：监听 2222
./geoyuan-ssh
# 或指定地址
GEOYUAN_ADDR=:2222 ./geoyuan-ssh
```

环境变量：

| 变量 | 默认 | 说明 |
|---|---|---|
| `GEOYUAN_ADDR` | `:2222` | 监听地址 |
| `GEOYUAN_DATA_DIR` | `data` | 数据目录（db + host key） |

客户端连接（开发）：

```bash
ssh -p 2222 localhost
```

## 映射到 22 端口（生产，不修改系统 sshd）

目标是让用户 `ssh diliy.cn`（默认 22）连到本服务。**不要改系统 `/etc/ssh/sshd_config`**，用以下任一方式：

**方式 A：setcap 绑定低端口（推荐，非 root 运行）**

```bash
sudo setcap 'cap_net_bind_service=+ep' /path/to/geoyuan-ssh
GEOYUAN_ADDR=:22 ./geoyuan-ssh
```

**方式 B：iptables 端口转发（服务仍跑 2222）**

```bash
sudo iptables -t nat -A PREROUTING -p tcp --dport 22 -j REDIRECT --to-port 2222
# 本机回环也要：
sudo iptables -t nat -A OUTPUT -p tcp -d 127.0.0.1 --dport 22 -j REDIRECT --to-port 2222
```

注意：若服务器 22 端口已被系统 sshd 占用，需先停掉或改用方式 B 把外部 22 转到 2222（系统 sshd 改监听其它端口或内网）。

> 部署提示：CI/CD（`.github/workflows/deploy-go.yml`）通过系统 sshd 的 2222 端口 SSH 部署，服务名在 Debian/Ubuntu 上是 `ssh.service`（不是 `sshd.service`），reload/restart 用 `systemctl reload ssh`。若部署报 `handshake failed`，先确认 sshd 的 `AllowUsers`/`AllowGroups` 未限制 GitHub runner IP。

## systemd 常驻

```ini
# /etc/systemd/system/geoyuan-ssh.service
[Unit]
Description=geoyuan SSH quiz service
After=network.target

[Service]
Type=simple
User=geoyuan
WorkingDirectory=/opt/geoyuan-ssh
Environment=GEOYUAN_ADDR=:2222
ExecStart=/opt/geoyuan-ssh/geoyuan-ssh
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now geoyuan-ssh
```

## 题库与阈值

- 题库固定 30 题，每次全考，定义在 `internal/questions/questions.go`，可直接编辑增改（`Answer` 为正确选项索引：A=0 … D=3）。
- 通过阈值常量 `PassThreshold`（默认 24）在 `internal/tui/app.go` 顶部，按需调整。

## 数据模型

- `users(id, username UNIQUE, pubkey_fingerprint UNIQUE, status, created_at)` — `status` 为 `pending` / `approved`
- `attempts(id, user_id, score, total, passed, created_at)` — 每次答题记录；通过时事务内同步把用户置为 `approved`

SQLite 开启 WAL + busy_timeout，`RecordAttempt` 用事务保证 attempt 记录与状态更新的原子性。
