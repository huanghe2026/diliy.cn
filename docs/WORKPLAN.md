# diliy.cn（Geoyuan）改进工作计划

> 版本：v2 · 2026-11-28
> 审批状态：待审批
> 聚焦范围：**TUI 功能修复（最高优先）** → 品牌对齐 → 运营增长

---

## 1. 背景与目标

`diliy.cn` 的 `ssh diliy.cn` TUI 主流程已打通（注册→准入测试→主菜单→Mint/My Cards/Marketplace/Chat），但部分界面功能未做完整。经代码逐层排查，确认存在以下"没做完"的缺口。

**目标**：先补齐 TUI 界面的功能断点，再推进品牌对齐与运营。

---

## 2. TUI 功能缺口诊断（已核实代码）

| # | 缺口 | 现状（代码依据） | 影响 |
|---|---|---|---|
| G1 | **Marketplace 无卡片详情页** | `viewMarket` 列表只渲染 Title/价格/Seller/cell/latlon；`Card.Desc/Color/Res/Timestamp/Owner` 字段存在但未展示，`MarketListing.Color` 也未渲染 | 用户购买前看不到卡片完整信息（描述/颜色/分辨率/铸造时间） |
| G2 | **My Cards 无选中详情** | `viewMyCards` 平铺列表，无方向键选中与 Enter 进详情；`Card.TxHash/Owner/Res/Timestamp` 未展示 | 无法查看单张卡完整链上信息 |
| G3 | ~~旧合约 7 张卡未迁移~~ | 旧合约 `0x2797...` 的 7 张卡**已决定作废**（不迁移，测试数据全部重新 Mint） | 无（作废） |
| G4 | **Marketplace 列表缺 Color 渲染** | `MarketListing.Color` 有数据，`viewMarket` 未画色块 | 列表辨识度低 |

> 说明：Chat 已支持公共/私聊(`:w`)/`:online`/`/me`/`:redeem`；Enter 触发 Buy 已在 `ef85a53` 修复（代码确认 B/Enter 同 case）。以上不在"未完成"之列。

---

## 3. 分阶段计划

### 🔴 第一期：TUI 功能修复（本期执行范围，最高优先）

| # | 任务 | 涉及文件 | 产出 |
|---|---|---|---|
| 1.1 | Marketplace 卡片详情页：Enter 选中 → 详情（完整 Desc/Color 色块/Res/H3/经纬度/Seller/Timestamp），详情内 B 购买、Esc 返回 | `internal/tui/app.go`（新增 view + phase）、`internal/tui/styles.go` | 详情页 |
| 1.2 | My Cards 选中详情：方向键选中 → Enter 详情（含 TxHash/Owner），详情内 L 挂售 / M 邻居 / Esc 返回 | `internal/tui/app.go` | 详情交互 |
| 1.3 | Marketplace 列表补 Color 色块渲染 | `internal/tui/app.go` | 列表辨识 |
| 1.4 | ~~旧合约迁移~~（已取消：旧卡作废，全部重新 Mint） | — | 不迁移 |
| 1.5 | Discover 全网浏览：遍历链上所有已铸卡片（totalSupply + tokenByIndex），按页浏览 | `internal/card/`、`internal/tui/app.go` | 全网卡片入口 |
| 1.6 | 个人主页/他人资料：从卡片 owner 或聊天点用户，看 display_name/tagline + 全部卡 | `internal/store/`、`internal/tui/app.go` | 个人主页 |

### 🟡 第二期：品牌与体验对齐

| # | 任务 | 角色 | 涉及文件 |
|---|---|---|---|
| 2.1 | 重写网站叙事，统一为「地理 NFT + 社交」 | 营销+设计 | `src/pages/index.astro`、`src/i18n/ui.ts`、`src/data/team.ts` |
| 2.2 | 网站终端 demo 与真实 TUI 对齐 | 美工 | `src/components/SshHero.astro` |
| 2.3 | 统一 Web/TUI logo/配色/字体（地理色系） | 美工 | `src/styles/global.css`、`ssh/internal/tui/styles.go` |
| 2.4 | 清理未提交脚本 `gen_logo_*.py`、`ssh/scripts/` | 程序员 | 仓库 |

### 🟢 第三期：运营与增长

| # | 任务 | 角色 | 产出 |
|---|---|---|---|
| 3.1 | 生产监控告警 + 日志轮转 + DB 巡检 | 运营 | 稳定性 |
| 3.2 | 重构 deploy.sh，去除 sed 改 sshd 锁死风险 | 运营 | 安全部署 |
| 3.3 | onboarding 漏斗分析（web→ssh→mint） | 营销 | 增长方案 |

---

## 4. 跨项目统一清单（贯穿各期）

### 4.1 品牌与文案
- 统一品牌名与叙事：`diliy.cn` / `diliy` / `GYIDpro` / `gyid.geoyuan.com` 共用「Geoyuan 地理元」品牌故事线。

### 4.2 链与钱包
- 统一链：Polygon Amoy（chainID=80002）。
- 统一合约：GeoyuanCard `0xCB441824e3A67Fb30e89A78d9dcFf4F9E618D748`（新合约）。
- 统一钱包派生：BIP44 `m/44'/60'/0'/0/{userID}`，服务端 seed，私钥不落盘。

### 4.3 nginx 与部署
- 统一 vhost 命名、systemd 服务名（geoyuan-ssh / geoyuan-api / nginx）、资源约束（nginx worker=2、TCP 256KB、swappiness=10）。

---

## 5. 验收标准（第一期）

> 状态：**代码已完成**，`go build` / `go vet` / `go test` 全部通过。以下为生产部署后的实机验收项。

- [x] Marketplace 列表展示 Color 色块。
- [x] Marketplace 可 Enter 进卡片详情页，B 可购买、Esc 返回。
- [x] My Cards 可方向键选中、Enter 进详情（含 Owner），L/M/Esc 正常。
- [x] Discover 全网浏览（主菜单第 7 项，遍历链上所有已铸卡片，分页）。
- [x] 个人主页（详情页按 U 查看 owner 资料 + 全部卡，owner 反查用户）。
- [x] ~~旧合约迁移~~ 已取消：旧卡作废，全部重新 Mint。

### 待生产验证（需部署后实测）
- [ ] 实机 `ssh diliy.cn` 走通：Market → Enter 详情 → Esc 返回；My Cards → 选中 → 详情。
- [ ] Discover 分页与链上数据一致。

---

## 6. 风险与依赖

| 风险 | 缓解 |
|---|---|
| TUI 改动影响线上 | 本地 `go build` + 单元测试通过后再部署 |
| 详情页需新增 Phase | 复用现有 headerBar/statusLine 统一布局，不破坏一致性 |
