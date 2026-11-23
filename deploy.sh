#!/usr/bin/env bash
set -euo pipefail

echo "=== Step 1: Fix sshd AllowUsers ==="
if grep -qE '^AllowUsers|^AllowGroups' /etc/ssh/sshd_config; then
  cp /etc/ssh/sshd_config /etc/ssh/sshd_config.bak.$(date +%s)
  sed -i 's/^AllowUsers/# AllowUsers (disabled)/' /etc/ssh/sshd_config
  sed -i 's/^AllowGroups/# AllowGroups (disabled)/' /etc/ssh/sshd_config
  rm -f /etc/ssh/sshd_config.d/*.conf
  sshd -t && systemctl reload sshd && echo "sshd unlocked"
else
  echo "sshd already open"
fi

echo "=== Step 2: Install Node.js ==="
if ! command -v node &>/dev/null || [[ "$(node --version | cut -d. -f1 | sed 's/v//')" -lt 18 ]]; then
  curl -fsSL https://deb.nodesource.com/setup_22.x | bash -
  apt-get install -y nodejs
fi
node --version

echo "=== Step 3: Clone/pull repo ==="
cd /root
if [ -d diliy.cn ]; then
  cd diliy.cn && git pull --ff-only || { git fetch --all; git reset --hard origin/main; }
else
  git clone --depth=1 https://github.com/huanghe2026/diliy.cn.git
  cd diliy.cn
fi

echo "=== Step 4: Build ==="
npm install --no-audit --no-fund
npm run build

echo "=== Step 5: Deploy to nginx ==="
mkdir -p /var/www/diliy.cn
rsync -avz --delete dist/ /var/www/diliy.cn/
nginx -s reload || systemctl reload nginx || true

echo "=== Step 6: Verify ==="
if curl -s http://localhost/ | grep -q 'bc1qes742zt'; then
  echo "DEPLOY SUCCESS - Sponsor online"
  curl -s http://localhost/ | grep -o 'bc1qes742zt[a-z0-9]*'
else
  echo "DEPLOY FAILED - check /root/diliy.cn"
  exit 1
fi
