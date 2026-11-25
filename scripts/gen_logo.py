#!/usr/bin/env python3
"""生成 GEOYUAN 半块 ANSI logo（优化版，最少 escape 代码）。

输出: ssh/internal/tui/assets/geoyuan_logo_halfblock.txt
这个文件直接 embed 进 Go binary，渲染时不用任何额外库。
"""

from PIL import Image, ImageDraw
import os

OUT_DIR = os.path.join(os.path.dirname(__file__), "..", "ssh", "internal", "tui", "assets")
os.makedirs(OUT_DIR, exist_ok=True)

# 每个字母 6 宽 × 12 高，7 字母 + 6 个 2px 间距
LETTER_W, LETTER_H, GAP = 6, 12, 2

RAINBOW = [
    (255,  64, 128),  # G 玫红
    (255, 140,  64),  # E 橙
    (255, 220,  64),  # O 黄
    ( 64, 255, 160),  # Y 绿
    ( 64, 200, 255),  # U 青
    (180, 100, 255),  # A 紫
    (255,  64, 200),  # N 粉
]

G = [
    "111111","100000","100000","100111","100001","100001",
    "111101","000001","000001","111111","000000","000000",
]
E = [
    "111111","100000","100000","111110","100000","100000",
    "100000","100000","100000","111111","000000","000000",
]
O = [
    "011110","111111","100001","100001","100001","100001",
    "100001","100001","100001","111111","011110","000000",
]
Y = [
    "100001","100001","010010","001100","000100","000100",
    "000100","000100","000100","000100","000100","000000",
]
U = [
    "100001","100001","100001","100001","100001","100001",
    "100001","100001","100001","111111","011110","000000",
]
A = [
    "001100","010010","100001","100001","100001","111111",
    "100001","100001","100001","100001","100001","000000",
]
N = [
    "100001","110001","101001","100101","100011","100001",
    "100001","100001","100001","100001","100001","000000",
]

LETTERS = [G, E, O, Y, U, A, N]

# 构建像素图（12 高 × 54 宽）
total_w = len(LETTERS) * LETTER_W + (len(LETTERS) - 1) * GAP
img = Image.new("RGBA", (total_w, LETTER_H), (0, 0, 0, 0))
draw = ImageDraw.Draw(img)

x = 0
for li, letter in enumerate(LETTERS):
    color = RAINBOW[li]
    for row, line in enumerate(letter):
        for col, ch in enumerate(line):
            if ch == "1":
                draw.point((x + col, row), fill=(color[0], color[1], color[2], 255))
    x += LETTER_W + GAP

# 2× 放大 → PNG（给未来 Sixel/Kitty 用）
img_big = img.resize((total_w * 2, LETTER_H * 2), Image.NEAREST)
png_path = os.path.join(OUT_DIR, "geoyuan_logo.png")
img_big.save(png_path, "PNG")
print(f"✅ PNG: {png_path} ({img_big.size[0]}×{img_big.size[1]})")

# ── 生成优化版半块 ANSI（最少 escape 序列）──
# 策略：逐 cell 渲染，相邻 cell 颜色相同时复用 escape
w, h = img.size
px = img.load()

out_lines = []

# 添加顶部发光（可选）：在 logo 上面加 1 行柔和光晕
# （暂时不加，让 logo 更紧凑）

for y_top in range(0, h, 2):  # 每 2 像素 = 1 个 text line
    y_bot = min(y_top + 1, h - 1)
    line = ""
    prev_fg = (-1, -1, -1)
    prev_bg = (-1, -1, -1)

    for x in range(w):
        pt = px[x, y_top]   # RGBA
        pb = px[x, y_bot]

        fg = (pt[0], pt[1], pt[2]) if pt[3] > 0 else None
        bg = (pb[0], pb[1], pb[2]) if pb[3] > 0 else None

        if fg is None and bg is None:
            # 透明 → 纯空格 + reset
            line += "\x1b[0m "
            prev_fg = prev_bg = (-1, -1, -1)
            continue

        esc = ""
        if fg != prev_fg:
            if fg is None:
                esc += "\x1b[39m"
            else:
                esc += f"\x1b[38;2;{fg[0]};{fg[1]};{fg[2]}m"
            prev_fg = fg if fg else (-1, -1, -1)

        if bg != prev_bg:
            if bg is None:
                esc += "\x1b[49m"
            else:
                esc += f"\x1b[48;2;{bg[0]};{bg[1]};{bg[2]}m"
            prev_bg = bg if bg else (-1, -1, -1)

        if fg is not None:
            ch = "▀"
        elif bg is not None:
            ch = " "
        else:
            ch = " "
        line += esc + ch

    line += "\x1b[0m"
    out_lines.append(line)

txt_path = os.path.join(OUT_DIR, "geoyuan_logo_halfblock.txt")
with open(txt_path, "w") as f:
    f.write("\n".join(out_lines) + "\n")

size_kb = os.path.getsize(txt_path) / 1024
print(f"✅ 半块 ANSI: {txt_path} ({w} cells × {len(out_lines)} lines, {size_kb:.1f} KB)")

# 预览到终端
print("\n📺 预览:")
for l in out_lines:
    print(l)
print("\x1b[0m")
