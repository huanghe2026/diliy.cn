#!/usr/bin/env bash
# geoyuan SSH 测试服务 —— 远程部署脚本（在阿里云服务器上运行）
#
# 作用：
#   1. 安装静态二进制 + 创建专用用户 + systemd 常驻
#   2. 把系统 sshd 从 22 挪到 2222，让出 22 给 geoyuan 服务
#
# 安全设计（防锁死）：
#   - 改 sshd 前，先在 2222 临时起一个 sshd，让你用新窗口验证能登录；
#     你不确认 yes，脚本绝不动主 sshd 配置。
#   - 主 sshd 重启失败时自动回滚到原配置（22），不会把你锁在外面。
#   - 全程请保持运行本脚本的这个 SSH 窗口不要关闭（保命通道）。
#
# 用法（在远程服务器上）：
#   chmod +x deploy.sh
#   sudo ./deploy.sh          # 推荐用 sudo 跑，避免中途反复输密码
#   # 或不用 sudo：./deploy.sh （会自动调 sudo，需多次输密码）
#
set -euo pipefail

# ──────────── 可配置变量 ────────────
NEW_SSHD_PORT=2222          # 系统 sshd 新端口
GEOYUAN_PORT=22             # geoyuan 测试服务端口
APP_DIR=/opt/geoyuan-ssh    # 安装目录
APP_USER=geoyuan            # 运行服务的专用用户
SSHD_CFG=/etc/ssh/sshd_config
SSHD_DROPIN_DIR=/etc/ssh/sshd_config.d
PORT_DROPIN="$SSHD_DROPIN_DIR/00-geoyuan-port.conf"
BACKUP="${SSHD_CFG}.geoyuan.bak.$(date +%Y%m%d%H%M%S)"
SSHD_UNIT=""              # 运行时探测：ssh（Debian/Ubuntu）或 sshd（RHEL 系）
SSHD_SOCKET_ACTIVE=0      # Ubuntu 24.04+ 默认用 ssh.socket 套接字激活
# ────────────────────────────────────

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
info()  { echo -e "${GREEN}[INFO]${NC} $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }
err()   { echo -e "${RED}[FAIL]${NC} $*"; }
die()   { err "$*"; exit 1; }

# 确认在远程、二进制存在
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN="$SCRIPT_DIR/geoyuan-ssh"
[[ -x "$BIN" ]] || die "找不到可执行二进制 $BIN ，请把 geoyuan-ssh 和本脚本放同一目录"

# 提权：若不是 root 就用 sudo 重跑自己（保留环境）
if [[ $EUID -ne 0 ]]; then
  info "非 root 运行，将通过 sudo 提权重跑本脚本"
  exec sudo -E "$0" "$@"
fi

info "以 root 身份运行（$(whoami) on $(hostname)）"

#############################################
# 0. 前置检查
#############################################
command -v sshd   >/dev/null || die "未找到 sshd"
command -v systemctl >/dev/null || die "未找到 systemctl"
command -v setcap >/dev/null || die "未找到 setcap（apt install libcap2-bin）"
command -v ss     >/dev/null || die "未找到 ss（iproute2）"
command -v python3 >/dev/null || die "未找到 python3（修改 sshd 配置需要）"
[[ -f "$SSHD_CFG" ]] || die "未找到 $SSHD_CFG"

# 二进制架构与本机匹配（避免拷错平台）
if command -v file >/dev/null 2>&1; then
  BIN_DESC="$(file -b "$BIN")"
  case "$(uname -m)" in
    x86_64)
      grep -qi 'x86-64' <<<"$BIN_DESC" || die "二进制架构与本机 x86_64 不符：$BIN_DESC（重新构建：GOOS=linux GOARCH=amd64 go build -o geoyuan-ssh .）"
      ;;
    aarch64|arm64)
      grep -qi 'aarch64' <<<"$BIN_DESC" || die "二进制架构与本机 $(uname -m) 不符：$BIN_DESC（重新构建：GOOS=linux GOARCH=arm64 go build -o geoyuan-ssh .）"
      ;;
  esac
  info "二进制架构校验通过（$(uname -m)）"
fi

# 探测 sshd 的 systemd 服务名，并检查是否套接字激活
if systemctl list-unit-files --type=service 2>/dev/null | grep -q '^ssh\.service'; then
  SSHD_UNIT=ssh
elif systemctl list-unit-files --type=service 2>/dev/null | grep -q '^sshd\.service'; then
  SSHD_UNIT=sshd
else
  die "未找到 ssh.service 或 sshd.service，无法安全迁移 sshd"
fi
if systemctl is-active --quiet "${SSHD_UNIT}.socket" 2>/dev/null; then
  SSHD_SOCKET_ACTIVE=1
  warn "检测到 ${SSHD_UNIT}.socket 套接字激活（新版 Ubuntu 默认），迁移时需先停用 socket"
fi
info "sshd 服务单元：$SSHD_UNIT"

# 检查目标端口当前占用
info "检查端口占用..."
if ss -ltn "( sport = :$GEOYUAN_PORT )" | grep -q ":$GEOYUAN_PORT "; then
  warn "$GEOYUAN_PORT 端口当前被占用（应为系统 sshd，稍后会挪走）"
fi
if ss -ltn "( sport = :$NEW_SSHD_PORT )" | grep -q ":$NEW_SSHD_PORT "; then
  die "$NEW_SSHD_PORT 已被占用，请先释放或换一个端口"
fi

#############################################
# 1. 安装二进制 + 专用用户 + 数据目录
#############################################
info "安装二进制到 $APP_DIR ..."
mkdir -p "$APP_DIR/data"
cp "$BIN" "$APP_DIR/geoyuan-ssh"
chmod 755 "$APP_DIR/geoyuan-ssh"

id "$APP_USER" >/dev/null 2>&1 || useradd --system --no-create-home --shell /usr/sbin/nologin "$APP_USER"
chown -R "$APP_USER:$APP_USER" "$APP_DIR"

# 赋予绑定低端口能力（非 root 绑 22 需要）
setcap 'cap_net_bind_service=+ep' "$APP_DIR/geoyuan-ssh"
info "setcap 完成：$APP_DIR/geoyuan-ssh"

#############################################
# 2. systemd unit
#############################################
UNIT=/etc/systemd/system/geoyuan-ssh.service
info "写入 systemd unit: $UNIT"
cat > "$UNIT" <<EOF
[Unit]
Description=geoyuan SSH quiz service
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=$APP_USER
Group=$APP_USER
WorkingDirectory=$APP_DIR
Environment=GEOYUAN_ADDR=:$GEOYUAN_PORT
Environment=GEOYUAN_DATA_DIR=$APP_DIR/data
ExecStart=$APP_DIR/geoyuan-ssh
Restart=on-failure
RestartSec=3
# 允许非 root 绑定低端口
AmbientCapabilities=CAP_NET_BIND_SERVICE
CapabilityBoundingSet=CAP_NET_BIND_SERVICE
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$APP_DIR/data

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
info "systemd unit 已写入，暂未启动（等 sshd 让出 $GEOYUAN_PORT）"

#############################################
# 3. 防火墙放行 2222（ufw / firewalld，如启用）
#############################################
if command -v ufw >/dev/null 2>&1 && ufw status | grep -q "Status: active"; then
  info "ufw 已启用，放行 $NEW_SSHD_PORT/tcp ..."
  ufw allow "$NEW_SSHD_PORT/tcp" || warn "ufw 放行失败，请手动处理"
fi
if command -v firewall-cmd >/dev/null 2>&1 && systemctl is-active --quiet firewalld 2>/dev/null; then
  info "firewalld 已启用，放行 $NEW_SSHD_PORT/tcp ..."
  firewall-cmd --permanent --add-port="$NEW_SSHD_PORT/tcp" >/dev/null || warn "firewalld 放行失败，请手动处理"
  firewall-cmd --reload >/dev/null || warn "firewalld reload 失败，请手动处理"
fi
warn "若阿里云安全组未放行 $NEW_SSHD_PORT 入方向，请先到阿里云控制台放行，否则下面验证 2222 会连不上！"

#############################################
# 4. 【关键】sshd 安全切换：先在 2222 验证
#############################################
info "=== sshd 端口切换（安全模式）==="
info "当前系统 sshd 还在 22，你的当前连接不受影响。"

# 4a. 临时在 2222 起一个 sshd 实例（不改任何配置文件）
info "在 $NEW_SSHD_PORT 临时启动一个 sshd 实例用于验证..."
# 用 -p 覆盖端口，复用主配置；-o 确保不与主配置冲突
/usr/sbin/sshd -p "$NEW_SSHD_PORT" -o "PidFile=/run/sshd-geoyuan-tmp.pid"
TMP_PID=$(cat /run/sshd-geoyuan-tmp.pid 2>/dev/null || pgrep -f "sshd -p $NEW_SSHD_PORT" | head -1)
info "临时 sshd 已启动 (pid=${TMP_PID:-unknown})，监听 $NEW_SSHD_PORT"

sleep 1
if ! ss -ltn | grep -q ":$NEW_SSHD_PORT "; then
  die "临时 sshd 未能监听 $NEW_SSHD_PORT，脚本中止。主 sshd 未受影响。"
fi

# 4b. 让用户验证
echo
echo -e "${YELLOW}>>> 请现在打开一个【新的终端窗口】执行：${NC}"
echo -e "    ${GREEN}ssh -p $NEW_SSHD_PORT $(logname 2>/dev/null || echo '<你的用户名>')@$(hostname -I | awk '{print $1}')${NC}"
echo -e "    ${YELLOW}（或用公网IP/域名）确认能正常登录到系统 shell。${NC}"
echo -e "    ${YELLOW}验证通过后回到本窗口输入 yes；若连不上输入 no，脚本会清理临时 sshd 并中止，原 22 完全不动。${NC}"
read -r -p "2222 端口能正常登录吗？(yes/no): " ans
if [[ "$ans" != "yes" ]]; then
  warn "你选择中止。正在清理临时 sshd..."
  kill "$TMP_PID" 2>/dev/null || pkill -f "sshd -p $NEW_SSHD_PORT" 2>/dev/null || true
  die "已中止，系统 sshd 仍在 22，未做任何改动。"
fi

# 4c. 验证通过：先停掉临时 sshd（释放 2222），再切主配置
info "验证通过，停止临时 sshd 并切换主配置..."
kill "$TMP_PID" 2>/dev/null || pkill -f "sshd -p $NEW_SSHD_PORT" 2>/dev/null || true
sleep 1

# 备份
cp "$SSHD_CFG" "$BACKUP"
info "已备份原配置到 $BACKUP"

# 统一的重启 / 回滚入口（兼容 socket 激活）
restart_sshd() {
  if [[ "$SSHD_SOCKET_ACTIVE" == "1" ]]; then
    systemctl disable --now "${SSHD_UNIT}.socket" >/dev/null 2>&1 || true
    systemctl enable --now "$SSHD_UNIT"
  else
    systemctl restart "$SSHD_UNIT"
  fi
}
rollback_sshd() {
  rm -f "$PORT_DROPIN"
  cp "$BACKUP" "$SSHD_CFG"
  if [[ "$SSHD_SOCKET_ACTIVE" == "1" ]]; then
    systemctl disable --now "$SSHD_UNIT" >/dev/null 2>&1 || true
    systemctl enable --now "${SSHD_UNIT}.socket" >/dev/null 2>&1 || true
  else
    systemctl restart "$SSHD_UNIT" >/dev/null 2>&1 || true
  fi
}

# 4d. 改端口：优先写 sshd_config.d drop-in（Include 在顶部，优先级最高），
#     同时兜底改主配置，兼容没有 Include 的系统。
mkdir -p "$SSHD_DROPIN_DIR"
printf 'Port %s\n' "$NEW_SSHD_PORT" > "$PORT_DROPIN"
info "已写入 $PORT_DROPIN"
python3 - "$SSHD_CFG" "$NEW_SSHD_PORT" <<'PY'
import sys, re
path, port = sys.argv[1], sys.argv[2]
with open(path) as f:
    lines = f.readlines()
out, seen = [], False
for ln in lines:
    m = re.match(r'^(\s*#?\s*)(Port\s+)(\d+)(.*)$', ln)
    if m:
        out.append(f"{m.group(1)}{m.group(2)}{port}{m.group(4)}\n")
        seen = True
    else:
        out.append(ln)
if not seen:
    out.append(f"\nPort {port}\n")
with open(path, 'w') as f:
    f.writelines(out)
print("Port ->", port, "already_present=", seen)
PY

# 4e. 语法检查 + 重启；失败自动回滚
info "检查 sshd 配置语法..."
if ! sshd -t; then
  err "sshd 配置语法错误，回滚..."
  rollback_sshd
  die "已回滚到原配置，sshd 未重启。请检查 $SSHD_CFG"
fi

info "重启 sshd（迁移到 $NEW_SSHD_PORT）..."
if ! restart_sshd; then
  err "sshd 重启失败，立即回滚..."
  rollback_sshd
  die "已回滚到原配置（22），请排查后重试。"
fi

sleep 2
if ! ss -ltn | grep -q ":$NEW_SSHD_PORT "; then
  err "sshd 重启后未监听 $NEW_SSHD_PORT，回滚..."
  rollback_sshd
  die "已回滚。当前窗口（原 22 连接）应仍可用。"
fi
info "sshd 已成功迁移到 $NEW_SSHD_PORT，22 已释放"
info "【确认】请新开窗口 ssh -p $NEW_SSHD_PORT ... 再次确认主 sshd 正常，然后回来继续"
read -r -p "主 sshd 在 $NEW_SSHD_PORT 正常吗？(yes/no): " ans2
[[ "$ans2" == "yes" ]] || die "你选择中止。22 已释放但 geoyuan 未启动，可手动 systemctl restart ssh 恢复(配置已是 $NEW_SSHD_PORT)。"

#############################################
# 5. 启动 geoyuan 占 22
#############################################
info "启动 geoyuan 服务，占用 $GEOYUAN_PORT ..."
systemctl enable --now geoyuan-ssh.service || true
sleep 2
if ! systemctl is-active --quiet geoyuan-ssh.service; then
  err "geoyuan-ssh 启动失败，自动把 sshd 恢复到 $GEOYUAN_PORT ..."
  rollback_sshd
  die "已回滚 sshd 到 $GEOYUAN_PORT。请查看 journalctl -u geoyuan-ssh"
fi

if ss -ltn | grep -q ":$GEOYUAN_PORT "; then
  info "geoyuan 已监听 $GEOYUAN_PORT ✅"
else
  err "geoyuan 未监听 $GEOYUAN_PORT，自动把 sshd 恢复到 $GEOYUAN_PORT ..."
  rollback_sshd
  die "已回滚 sshd 到 $GEOYUAN_PORT，请查看日志"
fi

#############################################
# 6. 生成一键回滚脚本 + 完成
#############################################
cat > "$APP_DIR/rollback-sshd.sh" <<ROLLBACK
#!/usr/bin/env bash
# 一键回滚：停 geoyuan，把系统 sshd 恢复到 22
set -euo pipefail
systemctl disable --now geoyuan-ssh.service || true
rm -f "$PORT_DROPIN"
cp "$BACKUP" "$SSHD_CFG"
if [[ "$SSHD_SOCKET_ACTIVE" == "1" ]]; then
  systemctl disable --now "$SSHD_UNIT" || true
  systemctl enable --now "${SSHD_UNIT}.socket"
else
  systemctl restart "$SSHD_UNIT"
fi
ss -ltn | grep -q ':22 ' && echo "sshd 已恢复监听 22"
ROLLBACK
chmod 755 "$APP_DIR/rollback-sshd.sh"

echo
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN} 部署完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo "  用户测试入口 :  ssh $(hostname -I | awk '{print $1}')   (端口 $GEOYUAN_PORT)"
echo "                 或 ssh diliy.cn（若域名解析到本机）"
echo "  系统管理登录 :  ssh -p $NEW_SSHD_PORT $(hostname -I | awk '{print $1}')"
echo "  服务状态     :  systemctl status geoyuan-ssh"
echo "  服务日志     :  journalctl -u geoyuan-ssh -f"
echo "  数据目录     :  $APP_DIR/data (geoyuan.db + host key)"
echo "  系统 sshd    :  systemctl status $SSHD_UNIT（现监听 $NEW_SSHD_PORT）"
echo "  原 sshd 备份 :  $BACKUP"
echo "  一键回滚     :  sudo $APP_DIR/rollback-sshd.sh"
echo
warn "提醒：阿里云安全组需放行 $NEW_SSHD_PORT 入方向，否则你无法用 -p $NEW_SSHD_PORT 登录管理！"
