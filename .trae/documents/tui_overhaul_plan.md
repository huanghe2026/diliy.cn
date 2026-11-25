# Geoyuan TUI 视觉重构 Plan

## 🎯 目标
1. **Logo**：识别度极低 → 高辨识度 6-8 行大像素字 + 多层彩虹描边 + 发光效果
2. **注册表单**：一行裸 textinput → 完整表单卡（label + placeholder + 字数提示 + 校验 + 多步 wizard）
3. **整体布局**：单屏堆叠 → `qodercli` 风格（三栏：顶部 logo 条 + 主内容 + 底部状态/hint 条）

## 📚 主流 TUI 参考

| 项目 | 风格 | 借鉴点 |
|---|---|---|
| **claude code** | 顶部大 logo + 中间对话 + 底部状态 | 顶部固定 header，logo 占 6-8 行 |
| **lazygit** | panel split，每块有彩色边框 + 标题 | 每个视图用 Border 包围，标题居中 |
| **fig** | wizard 表单，每步 label + hint | 注册/答题/结果页都有清晰字段 label |
| **btop** | 仪表盘风格 + 状态徽章 | 状态徽章（PENDING/PASSED/ONLINE）在 header 右侧 |

## 🔍 Repo 现状

### 文件结构
```
ssh/internal/tui/
├── app.go         # ~2100 行，所有 View/Update/Model 逻辑
├── styles.go      # 主题色板 + 3 个 ASCII logo + 渐变/边框/卡渲染函数
└── views_mint.go  # Mint 表单独立文件
```

### Logo 现状
- 当前：`geoyuanLogoBig`（rectangles 4 行，27 字符宽，彩虹渐变色）
- 备选：`geoyuanLogoDohone`（下划线大字 5 行）、`geoyuanLogoClassic`（block 6 行）
- 渲染：`RenderRainbowLogo()` 逐字符上色 → 问题是字体本身识别度就差

### 注册表单现状（viewRegister + registerInput）
```
⚡  Register          ← 只有一个标题
  > samhuang          ← 裸 prompt，无 label，无 placeholder
                      ← 无字数提示，无校验错误的友好提示
↵ confirm · Ctrl+C exit · :redeem <token>
```

### 全局布局现状
- Header bar 只在非 Register 阶段用 `headerBar()`（logo + breadcrumb + user）
- Register 阶段用 `headerBarCompact()`（只有 `◆ geoyuan · Join Geoyuan ·`）
- Welcome/Result 阶段没有统一的 panel/边框包裹
- 没有底部永久状态条

---

## 🏗 改动清单

### 1️⃣ Logo 重做（styles.go）

**方案**：用 `figlet -f Slant` 或 `dohone` 生成 **8 行高 + 80+ 字符宽**的大像素字，配合 **多层霓虹发光**效果。

```
                   ▄████▄   ███████ ▒█████   ██▓    ██░ ██ 
                  ▒██   ██▒ ██▒   █▜██▒  ██▒▓██░   ▓██░ ██▒
                  ░ ██▒▓██░▓██░  ██▒██░  ██▒▒██░   ▒██▀▀██░
                   ██▒▒ █ █ ▒▓█   ██░██   ██▒▒██░   ░▓█ ░██ 
                  ░██░  ██▒░██████▒╚██████▒▒███████▒░▓█▒░██▓
                  ░▒░  ░▒ ▒░ ░ ▒░▒░▒░ ╚═════░ ░ ▒ ▒░  ▒ ░░▒░▒
                  ░░  ░ ░ ░ ▒  ░ ▒ ▒░       ░ ░ ░ ░   ▒ ░▒░ ░
                   ░   ░      ░ ░ ░        ░ ░ ░ ░     ░  ░░ ░
```
（Slant 字体，8 行 × 80 字符，彩虹 + 描边 + 发光效果）

**函数改动**：
- `geoyuanLogoBig` → 换成 8 行 Slant 版
- 新增 `RenderGlowLogo()`：**先渲染一层 15% 透明度的"外发光"背景色，再渲染主彩虹 logo，再渲染 1px 暗色描边** → 三层叠加产生发光感
- `RenderRainbowLogo()` 保持兼容（小屏幕 fallback）

### 2️⃣ 注册表单重写（app.go + styles.go）

**从一行 textinput → 完整 Wizard 表单卡**：

```
┌─────────────────────────────────────────────────────────┐
│  ⚡  CREATE YOUR GEOYUAN ID                              │
│                                                          │
│  Username                                                │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│  > samhuang                                              │
│                                                          │
│  💡  3-20 chars · letters/numbers/hyphen only            │
│                                                          │
│  ✗  "admin" is reserved                                 │
│                                                          │
│  [↵] Confirm     [Esc] Quit     [?:redeem <token>]       │
└─────────────────────────────────────────────────────────┘
```

**新增**：
- `RenderRegisterCard()` 替换 RegisterWelcomeCard 为表单 wizard 卡
- 字段 label + 字数计数器 `(8/20)`
- placeholder `"enter your callsign..."`
- 实时校验（regex `^[a-zA-Z0-9_-]{3,20}$`）
- 错误用红框高亮 + 友好提示
- 底部 hintBar 明确三键

### 3️⃣ 全局布局重构（app.go）

**三栏固定结构**：
```
┌─ TOP BAR（永远可见）──────────────────────────────────┐
│  [LOGO]  [breadcrumb]              [user·IP·online]  │
│  ═════════════════════════════════════════════════════  │
├─ MAIN CONTENT（phase 切换）───────────────────────────┤
│  注册 / 答题 / 主菜单 / Mint / Cards / Market / Chat  │
│  每块用 lipgloss Border + 标题包裹                      │
├─ BOTTOM BAR（永远可见）───────────────────────────────┤
│  [phase·status]  [wallet·bal]  [keys·? help]  [Esc←]  │
└─────────────────────────────────────────────────────────┘
```

**新增**：
- `bottomBar()` 函数：phase 徽章 + 链上状态（部署钱包余额 + 用户钱包地址）+ 全局键位提示
- 所有 Phase 的 View 都先套 `lipgloss.Border()` 再填内容
- 主菜单改成 **卡片式 panel**（不是一行一行的 cursor），每项有 icon + 大标题 + 描述 + 数字角标

### 4️⃣ 主菜单视觉升级（viewMenu）

**从单行列表 → Panel Card 网格**：
```
┌─ GEOYUAN MAIN MENU ──────────────────────────────────┐
│                                                       │
│  ┌─ 🎴 Mint ──────┐  ┌─ 🃏 My Cards ─────────────┐  │
│  │  New GeoyuanCard│  │  3 owned · 1 for sale    │  │
│  │  H3 pin · AI q │  │  + neighbors              │  │
│  └────────────────┘  └───────────────────────────┘  │
│                                                       │
│  ┌─ 🛒 Marketplace ─┐  ┌─ 💬 Crew Chat ──────────┐  │
│  │  4 listings      │  │  online · 2 msgs unread │  │
│  │  browse · buy    │  │  hangout with builders  │  │
│  └─────────────────┘  └──────────────────────────┘  │
│                                                       │
│  ┌─ 🌐 geoyuan.com ─┐  ┌─ 🚪 Exit ───────────────┐  │
│  │  portal · docs   │  │  SSH session shutdown   │  │
│  └─────────────────┘  └──────────────────────────┘  │
│                                                       │
└───────────────────────────────────────────────────────┘
  ↑↓ 选卡 · Enter 进入 · 1-6 直达 · ? 帮助 · Esc 退出
```
（cursor 就是整个 panel 高亮 border + 发光效果）

### 5️⃣ 其他视图统一 panel 包裹

| 视图 | 当前 | 改后 |
|---|---|---|
| Welcome | 无 border | Border + 标题 + tagline |
| Quiz | 裸行 | Border + 进度条 `5/30 ▓▓▓▓░░░` |
| Result | RenderCongratsPass | Border + 分数仪表盘 + 三行统计 |
| Mint | views_mint.go 独立文件 | Border + 步骤指示 `Step 1/3` |
| MyCards | 裸列表 | Border + tab（Owned / For Sale / Neighbors） |
| Market | 裸列表 | Border + 分页 `Page 1/3` + 价格筛选 |
| Chat | 裸消息流 | Border + 在线徽章 + 未读红点 |
| ListCard | 裸列表 | Border + 金额验证提示 |

---

## 📂 Files to Change

| 文件 | 改什么 | 行数估算 |
|---|---|---|
| `styles.go` | 新 Slant logo + `RenderGlowLogo()` + 新 panel 边框组件 + 注册表单渲染函数 | +150 行 |
| `app.go` | 全局三栏布局重构 + `bottomBar()` + `viewRegister` 完整重写 + `viewMenu` panel 网格 + 所有 View 套 Border | +400 / -200 ≈ +200 |
| `views_mint.go` | Border 包裹 + 步骤指示 | +30 |

---

## 🔗 Dependencies

1. **Logo 生成**：需要本机有 `figlet` 或手动生成 8 行 Slant 字体 → 无额外 Go 依赖
2. **lipgloss Border**：项目已有，直接用 `lipgloss.RoundedBorder()` / `lipgloss.DoubleBorder()` / `lipgloss.ThickBorder()`
3. **Panel 高亮动画**：cursorBorder vs normalBorder，切换更新即可，不需要新 tick
4. **用户钱包地址**：`m.userWalletAddr()` 已经能算，bottomBar 直接显示前 6 后 4 位
5. **部署钱包余额**：需要 `chain.BalanceWei()` 调用，已经在 chain 包里有，app.go 能直接调用

---

## ✅ Validation

| 测试 | 预期 |
|---|---|
| `go build ./...` | 0 error |
| `go vet ./...` | 0 warning |
| `ssh diliy.cn` 24 行小终端 | 不溢出，布局降级 gracefully |
| `ssh diliy.cn` 50 行宽终端 | 完整三栏 + 大 logo + panel 网格 |
| Register 页 | 表单卡 + label + placeholder + 实时校验 |
| Esc 帮助 + 退出确认 | 仍然工作 |
| Quiz 进度条 | `▓▓▓▓▓▓▓▓░░░ 22/30` 精确显示 |
| bottomBar 钱包地址 | `0xa5c9...C8C` 格式正确 |
| 链上状态 | bottomBar 显示 `0.048 MATIC ✅` |

---

## ⚠️ Risks

| 风险 | 处理 |
|---|---|
| Slant 字体在窄终端溢出 | 加 `termWidth < 80` fallback 到 Dohone 5 行版 |
| Panel 网格渲染性能 | lipgloss 足够快，panel 数量固定（6 个），无性能担忧 |
| 注册表单改动影响 Redeem 流程 | `:redeem <token>` 逻辑不动，只改 UI 层 |
| Quiz 进度条不兼容现有 textinput | 进度条单独一行，不干扰 textinput.Blink tick |
| 改动量大（+600 行）引入回归 | 分 commit：logo → register → layout → menu → panels，每步单独 build+vet+deploy |
