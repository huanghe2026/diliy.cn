#!/usr/bin/env python3
"""生成三层 fallback 的 GEOYUAN logo：

1. Braille 高密度 ⣿⠿⢿  (8 点/cell = 比半块还高 4× 分辨率)
2. Half-block ▀  (已存在, 2× 分辨率)
3. 纯文本加粗 (任何终端都行)

输出目录: ssh/internal/tui/assets/
"""

from PIL import Image, ImageDraw
import os

OUT = os.path.join(os.path.dirname(__file__), "..", "ssh", "internal", "tui", "assets")
os.makedirs(OUT, exist_ok=True)

RAINBOW = [
    (255,  64, 128),  # G 玫红
    (255, 140,  64),  # E 橙
    (255, 220,  64),  # O 黄
    ( 64, 255, 160),  # Y 绿
    ( 64, 200, 255),  # U 青
    (180, 100, 255),  # A 紫
    (255,  64, 200),  # N 粉
]

LETTER_W, LETTER_H, GAP = 6, 12, 2

G = ["111111","100000","100000","100111","100001","100001","111101","000001","000001","111111","000000","000000"]
E = ["111111","100000","100000","111110","100000","100000","100000","100000","100000","111111","000000","000000"]
O = ["011110","111111","100001","100001","100001","100001","100001","100001","100001","111111","011110","000000"]
Y = ["100001","100001","010010","001100","000100","000100","000100","000100","000100","000100","000100","000000"]
U = ["100001","100001","100001","100001","100001","100001","100001","100001","100001","111111","011110","000000"]
A = ["001100","010010","100001","100001","100001","111111","100001","100001","100001","100001","100001","000000"]
N = ["100001","110001","101001","100101","100011","100001","100001","100001","100001","100001","100001","000000"]
LETTERS = [G, E, O, Y, U, A, N]

total_w = len(LETTERS) * LETTER_W + (len(LETTERS) - 1) * GAP
img = Image.new("RGBA", (total_w, LETTER_H), (0, 0, 0, 0))
draw = ImageDraw.Draw(img)

x = 0
for li, letter in enumerate(LETTERS):
    color = RAINBOW[li]
    for row, line in enumerate(letter):
        for col, ch in enumerate(line):
            if ch == "1":
                draw.point((x + col, row), fill=(*color, 255))
    x += LETTER_W + GAP

# 2×放大 PNG（保留）
img_big = img.resize((total_w * 2, LETTER_H * 2), Image.NEAREST)
img_big.save(os.path.join(OUT, "geoyuan_logo.png"), "PNG")

# ── 方案 A: Braille 高密度（每 cell 8 点 = 4 行 × 2 列 = 4× 分辨率）──
# Braille 编码：U+2800 + (row*8 + col)
# 4 行像素 → 1 行 Braille cell
# 2 列像素 → 1 Braille cell 的左右半
def pixel_to_braille(px_top_left, px_top_right,
                     px_mid_left,  px_mid_right,
                     px_bot_left, px_bot_right,
                     px_bot2_left, px_bot2_right):
    """8 个像素 → 1 Braille 字符"""
    dots = [px_top_left, px_mid_left, px_bot_left, px_bot2_left,
            px_top_right, px_mid_right, px_bot_right, px_bot2_right]
    code = 0
    for i, on in enumerate(dots):
        if on:
            code |= (1 << i)
    return chr(0x2800 + code)

# 先把像素图 2× 放大，让 Braille 有更多点可填
img_big2 = img.resize((total_w * 2, LETTER_H * 2), Image.NEAREST)
w2, h2 = img_big2.size
px2 = img_big2.load()

braille_lines = []
for y in range(0, h2, 4):  # 每 4 行像素 = 1 行 Braille
    line = ""
    for x in range(0, w2, 2):  # 每 2 列像素 = 1 Braille cell 左右
        # 取 2×4 像素块
        cols = []
        for dx in range(2):
            for dy in range(4):
                if x+dx < w2 and y+dy < h2:
                    pt = px2[x+dx, y+dy]
                    cols.append(pt[3] > 0)  # alpha > 0 = 实点
                else:
                    cols.append(False)
        ch = pixel_to_braille(*cols)
        # 颜色：取第一个实点的颜色
        fg = None
        for dx in range(2):
            for dy in range(4):
                if x+dx < w2 and y+dy < h2:
                    pt = px2[x+dx, y+dy]
                    if pt[3] > 0:
                        fg = (pt[0], pt[1], pt[2])
                        break
            if fg:
                break
        if fg:
            line += f"\x1b[38;2;{fg[0]};{fg[1]};{fg[2]}m{ch}\x1b[0m"
        else:
            line += ch
    braille_lines.append(line)

braille_path = os.path.join(OUT, "geoyuan_logo_braille.txt")
with open(braille_path, "w") as f:
    f.write("\n".join(braille_lines) + "\n")
print(f"✅ Braille: {braille_path} ({w2//2} cells × {len(braille_lines)} lines)")
print(f"   预览:")
for l in braille_lines:
    print(l)
print("\x1b[0m")

# ── 方案 B: Half-block ▀（保留现有逻辑）──
def gen_halfblock():
    w, h = img.size
    px = img.load()
    lines = []
    prev_fg = prev_bg = (-1, -1, -1)

    for y_top in range(0, h, 2):
        y_bot = min(y_top + 1, h - 1)
        line = ""
        for x in range(w):
            pt = px[x, y_top]
            pb = px[x, y_bot]
            fg = (pt[0], pt[1], pt[2]) if pt[3] > 0 else None
            bg = (pb[0], pb[1], pb[2]) if pb[3] > 0 else None

            if fg is None and bg is None:
                line += "\x1b[0m "
                prev_fg = prev_bg = (-1, -1, -1)
                continue

            esc = ""
            if fg != prev_fg:
                esc += f"\x1b[38;2;{fg[0]};{fg[1]};{fg[2]}m" if fg else "\x1b[39m"
                prev_fg = fg if fg else (-1, -1, -1)
            if bg != prev_bg:
                esc += f"\x1b[48;2;{bg[0]};{bg[1]};{bg[2]}m" if bg else "\x1b[49m"
                prev_bg = bg if bg else (-1, -1, -1)
            line += esc + ("▀" if fg else " ")
        line += "\x1b[0m"
        lines.append(line)
    return "\n".join(lines) + "\n"

hb = gen_halfblock()
hb_path = os.path.join(OUT, "geoyuan_logo_halfblock.txt")
with open(hb_path, "w") as f:
    f.write(hb)
print(f"\n✅ Half-block: {hb_path}")

# ── 方案 C: 纯文本加粗（兜底）──
# 用 Unicode bold 字符
bold_text = "𝗚𝗘𝗢𝗬𝗨𝗔𝗡"  # Unicode Mathematical Bold
plain_path = os.path.join(OUT, "geoyuan_logo_fallback.txt")
with open(plain_path, "w") as f:
    f.write(f"\x1b[1;35m{'═'*30}\n")
    f.write(f"\x1b[1;33m  GEOYUAN  \x1b[0m — \x1b[1;36mgeoyuan.com\x1b[0m\n")
    f.write(f"\x1b[1;35m{'═'*30}\x1b[0m\n")
print(f"✅ Fallback: {plain_path}")

# ── 方案 D: 启动 splash（渐变淡入）──
# 额外一个带 "powered by Polygon Amoy" 的 splash 横幅
splash = []
splash.append("\x1b[1;35m" + "╔" + "═"*52 + "╗")
splash.append("\x1b[1;35m║" + "\x1b[0m 🌍  \x1b[1;33mGEOYUAN\x1b[0m — \x1b[36mGeo-bound NFT Cards\x1b[0m" + " "*12 + "\x1b[1;35m║")
splash.append("\x1b[1;35m║" + "\x1b[0m  \x1b[35mPolygon Amoy · Bubble Tea SSH TUI\x1b[0m" + " "*8 + "\x1b[1;35m║")
splash.append("\x1b[1;35m╚" + "═"*52 + "╝\x1b[0m")

splash_path = os.path.join(OUT, "geoyuan_splash.txt")
with open(splash_path, "w") as f:
    f.write("\n".join(splash) + "\n")
print(f"✅ Splash: {splash_path}")
print(f"\n📺 Splash 预览:")
for l in splash:
    print(l)
print("\x1b[0m")
