#!/usr/bin/env bash
# post-build.sh — 在 npm run build 成功后跑，把 Sponsor 区块更新到 dist/
#
# 用法:
#   npm run build           # 你本地/服务器上跑（非沙箱环境）
#   bash scripts/post-build.sh
#   # 然后同步 dist/ 到服务器

set -euo pipefail

cd "$(dirname "$0")/.."

if [ ! -d dist ]; then
  echo "❌ dist/ 不存在 — 先跑 npm run build"
  exit 1
fi

echo "=== 1. 复制 public/sponsor 图片 ==="
mkdir -p dist/sponsor
cp -v public/sponsor/wechat.png dist/sponsor/ 2>/dev/null || true
echo "✅ sponsor images in dist/"

echo ""
echo "=== 2. 替换 index.html Sponsor 区块 ==="

python3 <<'PYEOF'
import re, os

sponsor_html = '''<section class="cta-section" id="sponsor"> <div class="container"> <div class="section-label" style="justify-content: center; display: flex;"> // sponsor </div> <h2>Buy us a coffee</h2> <p>If this project helped you, fuel the next build. Every tip keeps the servers humming and the coffee flowing.</p> <div class="sponsor-grid"> <div class="sponsor-card"> <div class="sponsor-icon sponsor-icon-paypal"> <svg viewBox="0 0 24 24" width="40" height="40" fill="#009cde"><path d="M7.076 21.337H2.47a.641.641 0 0 1-.633-.74L4.944 1.726A.77.77 0 0 1 5.712 1h7.472c2.333 0 4.049.48 5.034 1.428.838.806 1.082 1.793 1.065 2.896-.018 1.131-.316 2.311-.854 3.522-.096.214-.195.424-.301.638-1.144 2.2-2.737 3.328-4.787 3.328H9.74c-.507 0-1 .411-1 .918l-.457 2.877-.033.188-.598 3.77z"/><path d="M18.604 8.53c-.835 2.136-2.602 3.78-5.167 3.78h-2.63c-.55 0-1 .45-1 1l-.566 3.572-.041.22-.594 3.754-.041.236H11.5c.55 0 1-.45 1-1l.572-3.611.041-.226.595-3.75.041-.238h.595c1.85 0 3.28-.74 4.28-2.207.415-.608.715-1.272.89-1.979l-.012-.01z" fill="#003087"/></svg> </div> <div class="sponsor-name">PayPal</div> <p class="sponsor-note">For supporters outside China</p> <a href="https://paypal.me/geoyuan" class="btn btn-primary" target="_blank" rel="noopener"> Donate </a> </div> <div class="sponsor-card"> <div class="sponsor-icon sponsor-icon-wechat"> <svg viewBox="0 0 24 24" width="40" height="40" fill="#07c160"><path d="M8.691 2.188C3.891 2.188 0 5.476 0 9.53c0 2.212 1.17 4.203 3.002 5.55a.59.59 0 0 1 .213.665l-.39 1.48c-.019.07-.048.141-.048.213 0 .163.13.295.29.295a.326.326 0 0 0 .167-.054l1.903-1.114a.864.864 0 0 1 .717-.098 10.16 10.16 0 0 0 2.837.403c.276 0 .543-.027.811-.05-.857-2.578.157-4.972 1.932-6.446 1.703-1.415 3.882-1.98 5.853-1.838-.576-3.583-4.196-6.348-8.596-6.348zM5.785 5.991c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178A1.17 1.17 0 0 1 4.623 7.17c0-.651.52-1.18 1.162-1.18zm5.813 0c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178 1.17 1.17 0 0 1-1.162-1.178c0-.651.52-1.18 1.162-1.18zm5.34 2.867c-1.797-.052-3.746.512-5.28 1.786-1.72 1.428-2.687 3.72-1.78 6.22.942 2.453 3.666 4.229 6.884 4.229.826 0 1.622-.12 2.361-.336a.722.722 0 0 1 .598.082l1.584.926a.272.272 0 0 0 .14.047c.134 0 .24-.111.24-.247 0-.06-.023-.12-.038-.177l-.327-1.233a.582.582 0 0 1-.023-.156.49.49 0 0 1 .201-.398C23.024 18.48 24 16.82 24 14.98c0-3.21-2.931-5.837-6.656-6.088V8.89c-.135-.01-.27-.027-.407-.032zm-2.735 3.13c.535 0 .969.44.969.982a.976.976 0 0 1-.969.983.976.976 0 0 1-.969-.983c0-.542.434-.982.97-.982zm4.844 0c.535 0 .969.44.969.982a.976.976 0 0 1-.969.983.976.976 0 0 1-.969-.983c0-.542.434-.982.969-.982z"/></svg> </div> <div class="sponsor-name">WeChat Pay</div> <p class="sponsor-note">Scan with the app to tip</p> <img src="/sponsor/wechat.png" alt="WeChat Pay" class="sponsor-qr" loading="lazy" onerror="this.onerror=null;this.replaceWith(Object.assign(document.createElement('div'),{className:'sponsor-qr-fallback',textContent:'QR not uploaded yet'}))"> </div> <div class="sponsor-card"> <div class="sponsor-icon sponsor-icon-btc"> <svg viewBox="0 0 24 24" width="40" height="40" fill="#f7931a"><path d="M23.687 14.02c-.388 1.557-2.424 2.867-4.949 3.162.255 1.71.93 2.864 2.277 3.854-1.074.51-2.25 1.039-3.732 1.253v-2.348c-1.747.318-3.756-.056-4.696-1.174-.535-.638-.757-1.387-.762-2.16.006-1.252.443-2.224 1.231-2.952.82-.757 2.126-1.12 3.584-1.332V10.03c-1.276.064-2.505.314-3.502.975-1.04.688-1.787 1.704-2.082 2.927-.372 1.539-.182 3.039.65 4.287.857 1.28 2.254 2.228 3.958 2.468v2.36c-.488.06-.984.09-1.477.09-1.472 0-2.883-.245-4.173-.674v-2.31c-1.52.044-3.068-.174-4.15-1.146-.72-.65-1.103-1.606-1.128-2.625.002-1.317.565-2.423 1.478-3.297.88-.846 2.144-1.435 3.576-1.717v-2.28c-1.49.218-2.794.752-3.528 1.51-1.01 1.048-1.284 2.685-.78 4.128-.35-.02-.695-.075-1.03-.154V6.84c1.215.588 2.63.93 4.158 1.042V5.56c1.478-.044 3.003-.268 4.188-.77.686-.292 1.3-.683 1.783-1.215.97-1.07 1.234-2.735.648-4.176.486.125.965.281 1.437.463v2.35c1.246-.233 2.412-.713 3.212-1.617.615-.697.88-1.642.852-2.587.003-1.105-.418-2.124-1.118-2.876L23.7 1.494c.844.912 1.297 2.184 1.297 3.574 0 1.386-.447 2.664-1.297 3.574v5.378c.695-.876 1.112-1.975 1.115-3.28 0-1.564-.66-2.98-1.733-3.953zM12.046 5.77c-.985.26-1.803.735-2.236 1.48-.395.643-.338 1.46-.02 2.11.307.628.997 1.088 1.908 1.278 1.023.214 2.255.117 3.136-.315.877-.427 1.436-1.137 1.522-1.985.086-.872-.266-1.724-1.02-2.28-.794-.585-2.054-.78-3.29-.288zm.604 8.17c-1.094.06-2.018.492-2.486 1.212-.446.68-.37 1.584.088 2.296.46.716 1.255 1.216 2.206 1.376 1.166.196 2.528-.068 3.496-.668.936-.58 1.456-1.47 1.424-2.448-.032-.968-.61-1.82-1.592-2.298-.984-.478-2.306-.518-3.136-.47z"/></svg> </div> <div class="sponsor-name">Bitcoin</div> <p class="sponsor-note">BTC on-chain</p> <div class="sponsor-addr"> <code class="addr-text" data-copy="bc1qes742zt83qwuc0mnnctx5j9030v7z739cq4hxm">bc1qes742zt…39cq4hxm</code> <button class="addr-copy" data-copy="bc1qes742zt83qwuc0mnnctx5j9030v7z739cq4hxm" title="Copy address">⎘</button> </div> </div> <div class="sponsor-card"> <div class="sponsor-icon sponsor-icon-eth"> <svg viewBox="0 0 24 24" width="40" height="40" fill="#627eea"><path d="M11.944 17.97L4.58 13.62 11.943 24l7.37-10.38-7.372 4.35h.003zM12.056 0L4.69 12.223l7.365 4.354 7.365-4.354L12.056 0z"/></svg> </div> <div class="sponsor-name">Ethereum</div> <p class="sponsor-note">ETH / ERC-20</p> <div class="sponsor-addr"> <code class="addr-text" data-copy="0xa5c9E81F858aDB133539b606A01850d4e51c0C8C">0xa5c9E81…0C8C</code> <button class="addr-copy" data-copy="0xa5c9E81F858aDB133539b606A01850d4e51c0C8C" title="Copy address">⎘</button> </div> </div> <div class="sponsor-card"> <div class="sponsor-icon sponsor-icon-sol"> <svg viewBox="0 0 24 24" width="40" height="40" fill="#14f195"><path d="M4.6 16.25a.7.7 0 01.5-.2h14.05a.35.35 0 01.25.6l-2.75 2.75a.7.7 0 01-.5.2H2.1a.35.35 0 01-.25-.6L4.6 16.25zm0-9a.7.7 0 01.5-.2h14.05a.35.35 0 01.25.6L16.65 10.4a.7.7 0 01-.5.2H2.1a.35.35 0 01-.25-.6l2.75-2.75zm14.8-4.5a.7.7 0 01.5.2l2.75 2.75a.35.35 0 01-.25.6H8.35a.7.7 0 01-.5-.2L5.1 3.95a.35.35 0 01.25-.6h14.05zm0 18a.7.7 0 01.5.2l2.75 2.75a.35.35 0 01-.25.6H8.35a.7.7 0 01-.5-.2L5.1 22.05a.35.35 0 01.25-.6h14.05z"/></svg> </div> <div class="sponsor-name">Solana</div> <p class="sponsor-note">SOL / SPL</p> <div class="sponsor-addr"> <code class="addr-text" data-copy="EMU144daDFvB5RhdtPYrkKS3SBiARf4HpHbefUsz7Um3">EMU144daDF…z7Um3</code> <button class="addr-copy" data-copy="EMU144daDFvB5RhdtPYrkKS3SBiARf4HpHbefUsz7Um3" title="Copy address">⎘</button> </div> </div> </div> </div> </section>'''

# Add JS for copy buttons (injected before </body>)
copy_script = '''<script>
document.querySelectorAll('.addr-copy, .addr-text').forEach(el => {
  el.addEventListener('click', async () => {
    const text = el.dataset.copy || el.parentElement?.dataset.copy || '';
    if (!text) return;
    try { await navigator.clipboard.writeText(text); }
    catch {
      const ta = document.createElement('textarea');
      ta.value = text; document.body.appendChild(ta);
      ta.select(); document.execCommand('copy'); document.body.removeChild(ta);
    }
  });
});
</script>'''

for fname in ['index.html', 'zh/index.html']:
    path = os.path.join('dist', fname)
    if not os.path.exists(path):
        print(f"⚠️  {path} not found, skip")
        continue
    with open(path) as f:
        html = f.read()

    # Replace Sponsor section
    pattern = r'<section class="cta-section" id="sponsor">.*?</section>'
    new_html, n = re.subn(pattern, sponsor_html, html, count=1, flags=re.DOTALL)
    if n == 0:
        print(f"⚠️  Sponsor section NOT FOUND in {fname}")
    else:
        print(f"✅ {fname}: Sponsor updated ({n} match)")

    # Inject copy script before </body>
    if '</body>' in new_html and 'addr-copy' not in new_html:
        new_html = new_html.replace('</body>', copy_script + '</body>', 1)
        print(f"✅ {fname}: copy script injected")

    with open(path, 'w') as f:
        f.write(new_html)

PYEOF

echo ""
echo "=== 3. 追加 CSS 到 dist 的 _astro/*.css ==="
cat >> dist/_astro/*.css 2>/dev/null <<'CSSEOF'
.sponsor-icon { width:56px; height:56px; display:flex; align-items:center; justify-content:center; border-radius:12px; background:rgba(255,255,255,.04); border:1px solid var(--border); margin-bottom:2px; }
.sponsor-icon-paypal { background:rgba(0,156,222,.12); border-color:rgba(0,156,222,.3); }
.sponsor-icon-wechat { background:rgba(7,193,96,.12); border-color:rgba(7,193,96,.3); }
.sponsor-icon-btc    { background:rgba(247,147,26,.12); border-color:rgba(247,147,26,.3); }
.sponsor-icon-eth    { background:rgba(98,126,234,.12); border-color:rgba(98,126,234,.3); }
.sponsor-icon-sol    { background:rgba(20,241,149,.12); border-color:rgba(20,241,149,.3); }
.sponsor-addr { display:flex; align-items:center; gap:6px; background:rgba(0,0,0,.3); border:1px solid var(--border); border-radius:6px; padding:6px 8px; margin-top:auto; width:100%; max-width:220px; }
.addr-text { flex:1; font-family:'JetBrains Mono',monospace; font-size:.72rem; color:var(--muted); overflow:hidden; white-space:nowrap; text-overflow:ellipsis; cursor:pointer; user-select:all; }
.addr-copy { flex-shrink:0; background:transparent; border:none; color:var(--accent); font-size:1rem; cursor:pointer; padding:2px 6px; border-radius:4px; transition:background .2s; }
.addr-copy:hover { background:rgba(255,255,255,.08); }
CSSEOF

echo ""
echo "=== ✅ Done! dist/ ready to deploy ==="
ls -lh dist/index.html dist/sponsor/wechat.png
