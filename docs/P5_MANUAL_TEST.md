# Geoyuan P5 手动测试步骤（SSH TUI）

> **版本**: P5 市场功能 · 合约 v2 · commit `fc25557`
> **环境**: Polygon Amoy (chainID=80002) · 合约 `0xCB441824e3A67Fb30e89A78d9dcFf4F9E618D748`
> **前置**: 服务器已开机，所有服务 active，部署钱包余额 ≥ 0.3 MATIC

---

## 📋 测试前准备（必做）

| # | 操作 | 预期结果 | 怎么确认 |
|---|---|---|---|
| 1 | 确认服务全绿 | 3 service active | `ssh -p 2222 root@8.136.127.182 'systemctl is-active geoyuan-ssh geoyuan-api nginx'` 输出 3 个 `active` |
| 2 | 确认链上合约活着 | totalSupply ≥ 1 | `curl -s -X POST https://polygon-amoy-bor-rpc.publicnode.com -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"to":"0xCB441824e3A67Fb30e89A78d9dcFf4F9E618D748","data":"0x18160ddd"},"latest"]}'` → 返回 `0x01` 或更大 |
| 3 | 确认部署钱包有 gas | balance ≥ 0.3 MATIC | [Polygonscan 查 0xa5c9E8...](https://amoy.polygonscan.com/address/0xa5c9E81f858aDB133539b606A01850d4e51c0C8C) |
| 4 | 打开一个 SSH 窗口 | `ssh diliy.cn` 进入 TUI | 看到 Geoyuan logo + Register 页 |

---

## Phase 1 — 注册 + 答题（路径确认）

| Step | 操作 | 按键 | 预期屏幕 | ✅/❌ |
|---|---|---|---|---|
| 1.1 | 输入新用户名（不要重复） | 输入后 Enter | 出现 AI Quiz 第 1 题 | ☐ |
| 1.2 | 答前 5 题（或全部 30 题） | 数字键 1/2/3/4 选答案 | 最后看到绿色"✓ 全部通过"或"⚠ 未通过" | ☐ |
| 1.3 | 通过后等 2s | — | 自动进入主菜单（彩虹 logo + 6 个选项） | ☐ |

**常见失败**：答题未通过 → 出现重试页（3 选项），选第一项 🎯 重新挑战

---

## Phase 2 — Mint 第一张卡（链上写入验证）

| Step | 操作 | 按键 | 预期屏幕 | 链上验证 | ✅/❌ |
|---|---|---|---|---|---|
| 2.1 | 主菜单选 🎴 Mint a Geoyuan Card | ↑↓ 选 第 0 项 → Enter | 进入 Mint 视图，显示 H3 坐标 + 自动填充的 Lat/Lon | — | ☐ |
| 2.2 | 输入标题（必填） | 输入文字 → Tab | 光标跳到下一个字段 | — | ☐ |
| 2.3 | 输入描述（可选） | — | — | — | ☐ |
| 2.4 | 选颜色 | 方向键或数字 | 颜色方块高亮 | — | ☐ |
| 2.5 | 确认 Mint | Enter 到 "Mint on Polygon" → Enter | 显示 "⛓ Submitting to Polygon Amoy... (10-60s)" | — | ☐ |
| 2.6 | 等 receipt | — | 显示 "✓ Mint successful! Token #N" + 卡片信息 | `totalSupply` 从 1 → 2 | ☐ |
| 2.7 | 回主菜单 | Esc | 主菜单 6 选项 | — | ☐ |

**链上验证命令**（Mint 前后各跑一次）：
```bash
# 部署钱包（也是测试钱包）totalSupply
curl -s -X POST https://polygon-amoy-bor-rpc.publicnode.com \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"to":"0xCB441824e3A67Fb30e89A78d9dcFf4F9E618D748","data":"0x18160ddd"},"latest"]}'
```
→ 结果从 `"0x01"` 变成 `"0x02"` = Mint 成功

---

## Phase 3 — My Cards（链上读取 + 邻居匹配）

| Step | 操作 | 按键 | 预期屏幕 | ✅/❌ |
|---|---|---|---|---|
| 3.1 | 主菜单选 🃏 My Cards | ↑↓ 选 第 1 项 → Enter | "Loading... (10-30s)" → 然后列出你刚才 Mint 的卡 | ☐ |
| 3.2 | 看到卡片详情 | — | 显示 Token #N、Title、Color ■、H3 index、Lat/Lon | ☐ |
| 3.3 | 按 M 看邻居 | M | "Neighbor Match — k=2 around 0x..." 显示邻居卡（如果有的话） | ☐ |
| 3.4 | 按 Esc 回 My Cards | Esc | 回到你的卡片列表 | ☐ |
| 3.5 | 按 L 准备挂售 | L | 进入 "📦 List Card for Sale" 视图 | ☐ |

---

## Phase 4 — 挂售（**P5 新增**）

| Step | 操作 | 按键 | 预期屏幕 | 链上验证 | ✅/❌ |
|---|---|---|---|---|---|
| 4.1 | 在 List Card 视图选一张卡 | ↑↓ 选 → Enter 价格输入框 | 光标在 "MATIC 价格: " 输入框 | — | ☐ |
| 4.2 | 输入价格（如 `0.05` = 0.05 MATIC，或 `0` = 免费） | 输入数字 → Enter | 显示 "⛓ Submitting listing... (10-60s)" | — | ☐ |
| 4.3 | 等挂售成功 | — | 显示 "✓ Listing successful! Press Esc to return" | `listingCount` 从 1 → 2 | ☐ |
| 4.4 | 按 Esc 回 My Cards | Esc | 回到 My Cards（注意：你的卡可能从列表消失——因为已挂售到合约托管） | — | ☐ |

**链上验证命令**（挂售前后各跑一次）：
```bash
# 查挂售总数
curl -s -X POST https://polygon-amoy-bor-rpc.publicnode.com \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"to":"0xCB441824e3A67Fb30e89A78d9dcFf4F9E618D748","data":"0xa9b07c26"},"latest"]}'
```
→ 结果从 `"0x01"` 变成 `"0x02"` = 挂售成功

```bash
# 查你的卡是否在合约托管里（ownerOf 应该是合约地址本身）
# 把 1 换成你的 tokenId
curl -s -X POST https://polygon-amoy-bor-rpc.publicnode.com \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"to":"0xCB441824e3A67Fb30e89A78d9dcFf4F9E618D748","data":"0x6352211e00000000000000000000000000000000000000000000000000000000000000002"},"latest"]}'
```
→ 返回以 `0xCB441824...` 开头的地址 = 托管成功

---

## Phase 5 — Marketplace 浏览（**P5 新增**）

| Step | 操作 | 按键 | 预期屏幕 | ✅/❌ |
|---|---|---|---|---|
| 5.1 | 主菜单选 🛒 Marketplace | ↑↓ 选 第 2 项 → Enter | "Loading Geoyuan Marketplace..." → 列出所有挂售卡 | ☐ |
| 5.2 | 看到测试卡 #1 | — | 列表显示 "#1 P5 Test Card   ◎ 0.00 MATIC" | ☐ |
| 5.3 | 看到自己挂售的卡（如果 Phase 4 完成） | — | 列表显示你的 Token #N + Title + MATIC 价 | ☐ |
| 5.4 | 方向键导航 | ↑↓ 或 jk | 黄色 ▸ 光标移动，右侧显示 Seller 地址 + H3 cell | ☐ |
| 5.5 | R 刷新 | R | 重新从链上拉取 | ☐ |

**空挂售边界测试**：
- 如果 `listingCount = 0`，应该显示 "📭 No cards listed yet. Be the first..."

---

## Phase 6 — 购买（**P5 新增**）

| Step | 操作 | 按键 | 预期屏幕 | 链上验证 | ✅/❌ |
|---|---|---|---|---|---|
| 6.1 | Marketplace 里选中测试卡 #1 或自己挂售的卡 | ↑↓ 选中 | 黄色 ▸ 在目标卡上 | — | ☐ |
| 6.2 | 按 B 购买 | **B**（**注意**：Enter 不触发购买！只有 B 键） | 显示 "⛓ Buying... (10-60s)" | — | ☐ |
| 6.3 | 等待交易完成 | — | 显示 "✓ Bought Token #N!" + 你的新钱包地址 | `listingCount` 减 1；你的钱包余额增加 token | ☐ |
| 6.4 | 回 My Cards 看是否到账 | Esc → 主菜单 → My Cards | 应该能看到刚买的 token | — | ☐ |

**链上验证命令**（购买前后）：
```bash
# listingCount 应该减 1
curl -s -X POST https://polygon-amoy-bor-rpc.publicnode.com \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"to":"0xCB441824e3A67Fb30e89A78d9dcFf4F9E618D748","data":"0xa9b07c26"},"latest"]}'
```

```bash
# 查新 ownerOf — 应该是买家钱包地址
# 用 Go 工具查你的钱包：
cd ~/myweb/diliy.cn/ssh && go run ./cmd/mintas -uid <你的用户ID> -addr-only
# 然后查 ownerOf：
curl -s -X POST https://polygon-amoy-bor-rpc.publicnode.com \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"to":"0xCB441824e3A67Fb30e89A78d9dcFf4F9E618D748","data":"0x6352211e00000000000000000000000000000000000000000000000000000000000000001"},"latest"]}'
```

---

## Phase 7 — Crew Chat（回归）

| Step | 操作 | 按键 | 预期屏幕 | ✅/❌ |
|---|---|---|---|---|
| 7.1 | 主菜单选 💬 Crew Chat | ↑↓ 选 第 3 项 → Enter | 进入聊天室，看到历史消息 | ☐ |
| 7.2 | 发一条消息 | 输入文字 → Enter | 消息出现在聊天区，其他连接者能看到 | ☐ |
| 7.3 | 退出聊天 | Esc | 回主菜单 | ☐ |

---

## Phase 8 — 退出

| Step | 操作 | 按键 | 预期 | ✅/❌ |
|---|---|---|---|---|
| 8.1 | 主菜单选 🚪 Exit | 选 第 5 项 → Enter 或直接 q | SSH 连接关闭，回到本地 shell | ☐ |

---

## 🔧 按键速查表

| 视图 | 导航 | 确认 | 其他 |
|---|---|---|---|
| **主菜单** | ↑↓ / jk | Enter 或 Space | q/esc 退出 |
| **My Cards** | M=邻居, L=挂售, R=刷新 | — | Esc 回主菜单 |
| **Marketplace** | ↑↓ / jk 选卡 | **B = 购买**（Enter 没用！） | N=下一页, P=上一页, R=刷新, Esc=回主菜单 |
| **List Card** | ↑↓ 选卡 | Enter = 确认挂售 | Esc 取消 |
| **Mint** | Tab 跳字段 | Enter 确认 | Esc 取消 |
| **Quiz** | 1/2/3/4 选答案 | — | — |

---

## 🚨 已知限制

| 限制 | 原因 | 后续 |
|---|---|---|
| **旧合约 7 张卡不自动迁移** | P5 是全新合约 `0xCB44...`，旧的在 `0x2797...` | 写迁移脚本或在 TUI 加桥接 |
| **Enter 不触发 Buy** | 代码里 B/Enter 同 case 但又单独判断 `msg.String() != "b"` 提前 return | 已知，后续修 |
| **My Cards 挂售后卡消失** | token 转到合约托管了，ownerOf 变成合约地址 | 设计如此 |
| **Enter chat 后无法用聊天功能回归** | — | — |

---

## 📊 结果清单（测试完填表）

| 测试项 | 通过/失败 | 备注 |
|---|---|---|
| P1 注册 + Quiz | ☐ ✅ ☐ ❌ | |
| P2 Mint GeoyuanCard | ☐ ✅ ☐ ❌ | totalSupply 前: ____ → 后: ____ |
| P3 My Cards 读取 | ☐ ✅ ☐ ❌ | |
| P3.5 邻居匹配 | ☐ ✅ ☐ ❌ | |
| **P4 挂售（L 键）** | ☐ ✅ ☐ ❌ | listingCount 前: ____ → 后: ____ |
| **P5 Marketplace 浏览** | ☐ ✅ ☐ ❌ | |
| **P6 购买（B 键）** | ☐ ✅ ☐ ❌ | listingCount 前: ____ → 后: ____ |
| P7 Crew Chat | ☐ ✅ ☐ ❌ | |
| P8 退出 | ☐ ✅ ☐ ❌ | |

---

## 🔍 生产/链上一键巡检（脚本）

```bash
# 保存为 check.sh
#!/bin/bash
echo "=== 1. 服务状态 ==="
ssh -p 2222 root@8.136.127.182 'systemctl is-active geoyuan-ssh geoyuan-api nginx'

echo "=== 2. 端口监听 ==="
ssh -p 2222 root@8.136.127.182 'ss -tlnp | grep -E ":(22|2222|3002|80)\b"'

echo "=== 3. 链上合约 ==="
RPC="https://polygon-amoy-bor-rpc.publicnode.com"
CONTRACT="0xCB441824e3A67Fb30e89A78d9dcFf4F9E618D748"

# totalSupply
echo -n "totalSupply: "
curl -s -X POST $RPC -H 'Content-Type: application/json' \
  -d "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"eth_call\",\"params\":[{\"to\":\"$CONTRACT\",\"data\":\"0x18160ddd\"},\"latest\"]}" \
  | python3 -c "import sys,json; print(int(json.load(sys.stdin)['result'],16))"

# listingCount
echo -n "listingCount: "
curl -s -X POST $RPC -H 'Content-Type: application/json' \
  -d "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"eth_call\",\"params\":[{\"to\":\"$CONTRACT\",\"data\":\"0xa9b07c26\"},\"latest\"]}" \
  | python3 -c "import sys,json; print(int(json.load(sys.stdin)['result'],16))"

# 部署钱包余额
echo -n "deployer balance (MATIC): "
curl -s -X POST $RPC -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"eth_getBalance","params":["0xa5c9E81f858aDB133539b606A01850d4e51c0C8C","latest"]}' \
  | python3 -c "import sys,json; b=int(json.load(sys.stdin)['result'],16); print(f'{b/1e18:.4f}')"

echo "=== 4. 区块浏览器 ==="
echo "新合约: https://amoy.polygonscan.com/address/$CONTRACT"
echo "旧合约: https://amoy.polygonscan.com/address/0x2797C96bCC61622ca9Fa93e8478d67FB5De28e3E"
```
