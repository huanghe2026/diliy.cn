# diliy.cn SSH 准入测试服务 实施计划

## 需求

复刻 [pico.sh/getting-started](https://pico.sh/getting-started) 的入口模式：用户在终端 `ssh diliy.cn`，用公钥认证连上自建的 SSH 服务，进入一个**终端 TUI 测试**（30 道四选一 ABCD 题），答题通过即自动加入 geoyuan team。

## 已确认决策（来自用户）

| 维度 | 选择 |
|---|---|
| 题目领域 | 技术综合（编程 + 操作系统 + 网络 + 安全基础） |
| 通过后动作 | 达阈值即自动写入 team 名单（approved） |
| 交互形式 | 终端 TUI（方向键选 A-D，回车确认） |
| 数据存储 | 本地 SQLite |

## Repository Research

- `/home/huanghe/myweb/diliy.cn` 目前是纯 Astro 静态站（`geoyuan-site`），英文 `/`、中文 `/zh/`，`outDir: ./dist`。**无任何 Go 代码、无数据库、无后端服务**。
- 按钮文案已改为 `ssh diliy.cn`（上一轮完成），首页 `hero_terminal` 是 `huanghe@hangzhou.geoyuan.com:~# ./apply.sh` 的终端风格，与 TUI 调性一致。
- SSH 服务是**全新独立 Go 模块**，与 Astro 前端解耦（前端只负责展示 `ssh diliy.cn` 这个入口提示）。

## 模块结构（新建 `diliy.cn/ssh/`）

```
diliy.cn/ssh/
├── go.mod                          # module geoyuan.cn/ssh
├── main.go                         # 入口：加载 host key、启动 ssh server
├── README.md                       # 构建/部署说明
├── .gitignore                      # 忽略 *.db、host key
├── internal/
│   ├── questions/
│   │   └── questions.go            # 30 题题库（我生成，带答案+解析）
│   ├── store/
│   │   └── store.go                # SQLite：users / attempts 表，CRUD
│   ├── tui/
│   │   └── app.go                  # bubbletea：答题界面 + 结果页
│   └── sshserver/
│       └── server.go               # x/crypto/ssh：公钥认证 → 判定 → 进入 TUI
└── data/
    ├── geoyuan.db                  # 运行时生成（gitignore）
    └── ssh_host_ed25519_key        # host key，首次运行生成（gitignore）
```

## 依赖与关键选型（默认值，审批时可调整）

- **TUI 库**：`github.com/charmbracelet/bubbletea`（Go 生态标准 TUI，支持方向键/回车，与 PTY 兼容好）
- **SSH**：`golang.org/x/crypto/ssh`（用户指定）
- **SQLite**：`modernc.org/sqlite`（纯 Go，免 cgo，跨平台编译部署简单）
- **通过阈值**：≥ 24/30（80%）标记为 approved
- **题库**：固定一套 30 题（我生成），每次全考；后续可扩展为随机抽题
- **用户名**：首次连接时 TUI 提示输入用户名，校验唯一性后绑定其 SSH 公钥指纹
- **端口**：开发监听 `:2222`，生产通过端口转发/setcap 映射 `diliy.cn:22`；**绝不修改系统 sshd 配置**
- **重测策略**：未通过可重测（每次记录 attempt，取最新状态）；通过后锁定 approved

## 实施步骤（依赖顺序）

1. **初始化 Go module + 依赖**：`ssh/go.mod`，`go get` x/crypto/ssh、bubbletea、modernc.org/sqlite
2. **题库 `questions.go`**：我生成 30 道技术综合题（编程/OS/网络/安全各若干），结构 `{Q string, Options [4]string, Answer int, Explain string}`
3. **`store/store.go`**：SQLite 建表 `users(id, username, pubkey_fingerprint UNIQUE, status, created_at)`、`attempts(id, user_id, score, total, passed, created_at)`；封装 UpsertUser / RecordAttempt / GetUserByFingerprint 等
4. **`tui/app.go`**：bubbletea Model —— 显示题目与 4 选项、↑↓ 或 1-4 选择、回车下一题、进度条、最后显示分数与通过/失败结果
5. **`sshserver/server.go`**：
   - `ssh.ServerConfig`，`PublicKeyCallback` 按指纹查 user
   - 首次连接（无 user）→ 分配用户名 → 进入 TUI 测试
   - 已有 user 且 approved → 显示 "Welcome to geoyuan team" 欢迎页
   - 已有 user 未通过 → 可选择重测
   - 测试结束后写 attempts，达阈值则更新 users.status=approved
   - 请求 PTY，把 PTY 的 reader/writer 接给 bubbletea（`tea.WithInput` / `tea.WithOutput`）
6. **`main.go`**：加载/生成 host key（ed25519），监听端口，每个连接起 goroutine
7. **本地端到端验证**：`go run .` 起服务 → `ssh -p 2222 localhost` 跑通首次注册→答题→通过→重连欢迎的全流程；再测失败分支
8. **`README.md`**：构建、host key 生成、端口映射、systemd 常驻部署说明

## 验证

- `cd ssh && go build ./...` 编译通过
- `go vet ./...` 无告警
- 本地 `ssh -p 2222 localhost`：
  - 首次连接能注册用户名、进入 TUI、方向键选择正常、回车提交
  - 30 题答完显示分数，≥24 显示 approved 欢迎，<24 显示未通过
  - 断开重连：approved 用户直接进欢迎页；未通过用户可重测
  - SQLite 中 users/attempts 数据正确写入
- 终端兼容性：在普通 xterm/终端下 TUI 不花屏（bubbletea 默认处理）

## 风险与处理

- **22 端口权限**：非 root 不能监听 1024 以下端口 → 开发用 2222；生产用 `setcap 'cap_net_bind_service=+ep'` 或 nginx/防火墙转发 22→2222，不动系统 sshd
- **bubbletea × ssh PTY 适配**：需把 ssh channel 的 PTY IO 正确转交给 bubbletea；若遇 raw mode 问题，回退到 bubbletea 的 `tea.WithInput(reader)` + 手动设置 raw（x/crypto/ssh 的 Pty 回调已提供窗口大小）
- **公钥认证首次无用户**：`PublicKeyCallback` 对未知指纹也要返回成功（让连接进来），真正的"准入"由测试结果决定，而非认证阶段拒绝
- **SQLite 并发写**：多连接同时答题 → 开启 WAL 模式 + 每条连接独立语句，避免锁冲突
- **host key 持久化**：首次运行生成并写文件，重启复用（否则用户每次连接都提示 host key 变更）；key 文件加入 gitignore
- **scope 控制**：本次只做 SSH 测试准入，不接 geoyuan 后续生产环境的实际授权（approved 标志先落地，环境接入后续迭代）
