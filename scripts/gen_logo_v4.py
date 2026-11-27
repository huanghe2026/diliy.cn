#!/usr/bin/env python3
"""重新设计 Braille logo v4 — 超大粗体！

每字母 24 高 × 16 宽 → 6 Braille 行 × 8 cell
笔画填充率 100%（每 cell 8 点尽量满）

目标：终端里一眼就能认出 G-E-O-Y-U-A-N
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

# 每字母 16 宽 × 24 高（6 Braille 行）
W, H, GAP = 16, 24, 4

# ── 粗体 G（24高，右下大钩，全填充）──
G = []
for r in range(H):
    row = ""
    for c in range(W):
        is_top = r == 0 or r == 1
        is_bot = r == H-2 or r == H-1
        is_left = c <= 1 and not is_top  # 左竖
        is_right = c >= W-2 and (is_top or is_bot or r >= H//2-1)  # 右竖 + 下半钩
        is_mid = (r >= H//2-2 and r <= H//2+1) and c >= W//2 and c <= W-3  # G 中间的横
        if is_top or is_bot or is_left or is_right or is_mid:
            row += "1"
        else:
            row += "0"
    G.append(row)

# ── 粗体 E（24高，三横）──
E = []
for r in range(H):
    row = ""
    for c in range(W):
        is_top = r <= 1
        is_bot = r >= H-2
        is_mid = (r >= H//2-2 and r <= H//2+1)
        is_left = c <= 1
        is_fill = is_top or is_bot or is_mid
        if is_fill or (is_left and not is_bot):
            row += "1"
        else:
            row += "0"
    E.append(row)

# ── 粗体 O（24高，椭圆全填充边框）──
O = []
for r in range(H):
    row = ""
    cy = H / 2
    for c in range(W):
        cx = W / 2
        dx = (c - cx) / (W/2 - 1)
        dy = (r - cy) / (H/2 - 1.5)
        dist = dx*dx + dy*dy
        if 0.6 <= dist <= 1.15:
            row += "1"
        else:
            row += "0"
    O.append(row)

# ── 粗体 Y（24高，V收窄 + 厚竖）──
Y = []
for r in range(H):
    row = ""
    mid = H // 2 + 1
    if r < mid:
        # V 形：两边往中间收
        v_offset = int((r / mid) * (W // 2 - 1))
        left = W//2 - 1 - v_offset
        right = W//2 + v_offset
        for c in range(W):
            row += "1" if c <= left or c >= right else "0"
    else:
        # 单粗竖
        for c in range(W):
            row += "1" if c >= W//2 - 2 and c <= W//2 + 1 else "0"
    Y.append(row)

# ── 粗体 U（24高，双竖 + 厚底）──
U = []
for r in range(H):
    row = ""
    for c in range(W):
        is_left = c <= 1
        is_right = c >= W-2
        is_bot = r >= H-3
        is_fill = is_left or is_right or is_bot
        row += "1" if is_fill else "0"
    U.append(row)

# ── 粗体 A（24高，双竖 + 中横 + 尖顶）──
A = []
for r in range(H):
    row = ""
    mid = H // 2
    # 双竖（倾斜收窄 → 中横 → 直立）
    if r < mid:
        # 上半：从窄顶 → 宽底
        ratio = r / mid
        left = int(W//2 - ratio * (W//2 - 1))
        right = int(W//2 + ratio * (W//2 - 1))
    else:
        # 下半：直立双竖
        left, right = 1, W - 2
    for c in range(W):
        is_vert = c == left or c == right
        is_mid_bar = (r >= mid-1 and r <= mid+2) and c >= left and c <= right
        is_apex_top = r == 0 and c == W//2
        row += "1" if (is_vert and r > 0) or is_mid_bar or is_apex_top else "0"
    A.append(row)

# ── 粗体 N（24高，左竖 + 斜线 + 右竖全厚）──
N = []
for r in range(H):
    row = ""
    for c in range(W):
        is_left = c <= 1
        is_right = c >= W-2
        # 斜线：从 (0,0) 到 (W-1,H-1)，厚度 2
        slope = (c / (W-1)) * (H-1)
        is_diag = abs(r - slope) <= 1.5 and not is_left and not is_right
        row += "1" if is_left or is_right or is_diag else "0"
    N.append(row)

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

# ── 转 Braille ──
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

print(f"✅ 超大粗体 Braille: {out_path}")
print(f"   {total_w}px × {H}px → {len(braille_lines)} Braille rows × {len(braille_lines[0])//10} cells/row")
print(f"\n📺 预览 (无颜色):")
for l in braille_lines:
    # strip ANSI for plain preview
    plain = ''.join(ch for ch in l if ord(ch) > 31)
    print(plain)
print(f"\n📺 彩色预览:")
for l in braille_lines:
    print(l)
print("\x1b[0m")
