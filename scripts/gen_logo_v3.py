#!/usr/bin/env python3
"""重新设计 Braille logo — 粗体厚描边 GEOYUAN

之前的字母太薄太细，Braille 点虽然密但笔画没对齐。
新设计：每字母 14 高 × 12 宽（Braille 7 行 × 6 细胞），粗体全填充。

输出: ssh/internal/tui/assets/geoyuan_logo_braille.txt (覆盖旧版)
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

# 每字母 14 高 × 12 宽（粗体！）
# 用 1/0 矩阵，1=填色, 0=空
W, H, GAP = 12, 14, 4

# ── 粗体 G（右下大钩）──
G = [
    "011111111100",  # 1 full top bar
    "111111111110",
    "110000000011",
    "110000000011",
    "110000000011",
    "110001111011",  # G 的右下横
    "110001111111",
    "110000000011",
    "110000000011",
    "110000000011",
    "110000000011",
    "111111111111",  # 1 full bottom bar
    "011111111100",
    "000000000000",
]

# ── 粗体 E（三横）──
E = [
    "111111111110",
    "111111111111",  # full top
    "110000000000",
    "110000000000",
    "110000000000",
    "111111111110",  # middle bar
    "111111111111",
    "110000000000",
    "110000000000",
    "110000000000",
    "110000000000",
    "111111111110",  # bottom
    "111111111111",
    "000000000000",
]

# ── 粗体 O（完整椭圆）──
O = [
    "001111111100",
    "011111111110",
    "111100001111",
    "110000000011",
    "110000000011",
    "110000000011",
    "110000000011",
    "110000000011",
    "110000000011",
    "110000000011",
    "111100001111",
    "011111111110",
    "001111111100",
    "000000000000",
]

# ── 粗体 Y（V 形收窄，更清晰）──
Y = [
    "110000000011",
    "111000000111",  # top V
    "111100001111",
    "011110011110",
    "001111111100",  # V 交汇处
    "000111111000",
    "000011110000",
    "000011110000",  # 单竖 thick
    "000011110000",
    "000011110000",
    "000011110000",
    "000011110000",
    "000011110000",
    "000000000000",
]

# ── 粗体 U（双竖 + 厚底）──
U = [
    "110000000011",
    "110000000011",
    "110000000011",
    "110000000011",
    "110000000011",
    "110000000011",
    "110000000011",
    "110000000011",
    "110000000011",
    "110000000011",
    "110000000011",
    "111111111111",
    "011111111110",
    "000000000000",
]

# ── 粗体 A（双竖 + 中横）──
A = [
    "000011110000",
    "000111111000",
    "001110011100",
    "011100001110",
    "111000000111",
    "110000000011",
    "110011110011",  # middle bar thick
    "110011110011",
    "110000000011",
    "110000000011",
    "110000000011",
    "110000000011",
    "110000000011",
    "000000000000",
]

# ── 粗体 N（左竖 + 斜线 + 右竖）──
N = [
    "110000000011",
    "111000000111",  # diagonal thick
    "111100001111",
    "110110011011",
    "110011110011",
    "110001110011",
    "110000110011",
    "110000110011",
    "110000110011",
    "110000110011",
    "110000110011",
    "110000110011",
    "110000110011",
    "000000000000",
]

LETTERS = [G, E, O, Y, U, A, N]

# ── 渲染像素图 ──
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

# ── 转 Braille（每 2×4 像素块 = 1 cell）──
def to_braille(img_src):
    w, h = img_src.size
    px = img_src.load()
    out = []
    for y in range(0, h, 4):
        line = ""
        for x in range(0, w, 2):
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

braille_lines = to_braille(img)

# 保存
out_path = os.path.join(OUT, "geoyuan_logo_braille.txt")
with open(out_path, "w") as f:
    f.write("\n".join(braille_lines) + "\n")

print(f"✅ 粗体 Braille logo: {out_path}")
print(f"   {total_w}px wide × {H}px tall → {len(braille_lines)} Braille rows")
print(f"\n📺 预览:")
for l in braille_lines:
    print(l)
print("\x1b[0m")
