#!/usr/bin/env python3
"""生成 Braille 高密度启动欢迎横幅（代替朴素 splash）。

设计思路：
- Braille 每 cell 8 点 = 比 ASCII 密 4×
- 用 Braille 画一个 "霓虹发光" 横幅
- 中央: 小版 GEOYUAN Braille 字（4行）
- 下方: 彩虹分隔线 + tagline + 链上徽章
- 纯 Braille + truecolor escape，终端宽 ≥ 80 最佳

输出: ssh/internal/tui/assets/geoyuan_splash_braille.txt
"""

from PIL import Image, ImageDraw, ImageFont
import os

OUT = os.path.join(os.path.dirname(__file__), "..", "ssh", "internal", "tui", "assets")
os.makedirs(OUT, exist_ok=True)

# ── 方法 1: 画 Braille 版 GEOYUAN（4 行高 × 紧凑） ──
# 用 10×6 block letter（比主 logo 小，适合横幅）

RAINBOW = [
    (255,  64, 128),  # G
    (255, 140,  64),  # E
    (255, 220,  64),  # O
    ( 64, 255, 160),  # Y
    ( 64, 200, 255),  # U
    (180, 100, 255),  # A
    (255,  64, 200),  # N
]

# 紧凑字母 4×4（每字 4 宽 × 8 高 → 4 Braille 行）
W, H, GAP = 4, 8, 1

G4 = ["1111","1001","1001","1001","1111","0001","0001","1111"]
E4 = ["1111","1000","1000","1110","1000","1000","1000","1111"]
O4 = ["1111","1001","1001","1001","1001","1001","1001","1111"]
Y4 = ["1001","1001","0110","0110","0010","0010","0010","0010"]
U4 = ["1001","1001","1001","1001","1001","1001","1001","1111"]
A4 = ["0110","1001","1001","1001","1111","1001","1001","1001"]
N4 = ["1001","1101","1011","1001","1001","1001","1001","1001"]
LETTERS = [G4, E4, O4, Y4, U4, A4, N4]

total_w = len(LETTERS) * W + (len(LETTERS) - 1) * GAP
img = Image.new("RGBA", (total_w, H), (0, 0, 0, 0))
draw = ImageDraw.Draw(img)

x = 0
for li, letter in enumerate(LETTERS):
    color = RAINBOW[li]
    for row, line in enumerate(letter):
        for col, ch in enumerate(line):
            if ch == "1":
                draw.point((x + col, row), fill=(*color, 255))
    x += W + GAP

# ── 转 Braille ──
def to_braille_lines(img_src, scale=2):
    """把像素图放大 scale× → 转 Braille（每 cell 4×2 像素）"""
    big = img_src.resize((img_src.size[0]*scale, img_src.size[1]*scale), Image.NEAREST)
    w, h = big.size
    px = big.load()
    out = []
    for y in range(0, h, 4):
        line = ""
        for x in range(0, w, 2):
            # 取 2×4 块
            dots = []
            fg = None
            for dx in range(2):
                for dy in range(4):
                    if x+dx < w and y+dy < h:
                        pt = px[x+dx, y+dy]
                        on = pt[3] > 0
                        dots.append(on)
                        if on and fg is None:
                            fg = (pt[0], pt[1], pt[2])
                    else:
                        dots.append(False)
            code = 0
            for i, d in enumerate(dots):
                if d:
                    code |= (1 << i)
            ch = chr(0x2800 + code)
            if fg:
                line += f"\x1b[38;2;{fg[0]};{fg[1]};{fg[2]}m{ch}\x1b[0m"
            else:
                line += ch
        out.append(line)
    return out

braille_logo = to_braille_lines(img, scale=3)  # 3× 放大 → 更密

# ── 生成彩虹渐变线 ──
def gradient_line(text, colors):
    """给文本加彩虹渐变（逐字符）"""
    out = ""
    for i, ch in enumerate(text):
        c = colors[i % len(colors)]
        out += f"\x1b[38;2;{c[0]};{c[1]};{c[2]}m{ch}"
    out += "\x1b[0m"
    return out

rainbow_bar = gradient_line("═" * 60, RAINBOW)

# ── 组装横幅 ──
# 横幅结构:
#   ⣿ Braille logo (居中)
#   ═══════════════════════════════════════════════════════════
#   🌍  Geo-bound NFT Cards · Polygon Amoy · Bubble Tea SSH TUI
#   ═══════════════════════════════════════════════════════════

splash_lines = []

# Braille logo（居中）
max_len = max(len(l) for l in braille_logo)
for bl in braille_logo:
    padding = max(0, (72 - max_len) // 2)
    splash_lines.append(" " * padding + bl)

# 分隔线
splash_lines.append(rainbow_bar)

# 标签行（彩虹强调）
tagline = (
    f"\x1b[1;35m  🌍  \x1b[0m"
    f"\x1b[1;33mGeo-bound NFT Cards\x1b[0m"
    f"\x1b[36m  ·  Polygon Amoy  ·  Bubble Tea SSH TUI\x1b[0m"
)
splash_lines.append(tagline)

# 底部分隔
splash_lines.append(rainbow_bar)

# 版权/版本行（dim）
splash_lines.append(
    "\x1b[2m  powered by Polygon · geoyuan.com · v0.5 P5 Marketplace\x1b[0m"
)

out_path = os.path.join(OUT, "geoyuan_splash_braille.txt")
with open(out_path, "w") as f:
    f.write("\n".join(splash_lines) + "\n")

print(f"✅ Braille Splash: {out_path}")
print(f"   共 {len(splash_lines)} 行")
print(f"\n📺 预览:")
for l in splash_lines:
    print(l)
print("\x1b[0m")
