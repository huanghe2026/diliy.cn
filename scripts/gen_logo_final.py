#!/usr/bin/env python3
"""最终版 Braille + Half-block + Fallback 三合一
Braille 用 figlet 生成大 ASCII block 字（填充率 100%），每个字符 cell 填 8 个点 = 真正粗体！
"""

from PIL import Image, ImageDraw, ImageFont
import os, subprocess, shutil

OUT = os.path.join(os.path.dirname(__file__), "..", "ssh", "internal", "tui", "assets")
os.makedirs(OUT, exist_ok=True)

RAINBOW = [
    (255,  64, 128), (255, 140,  64), (255, 220,  64),
    ( 64, 255, 160), ( 64, 200, 255), (180, 100, 255),
    (255,  64, 200),
]

# ── 方法 1: figlet ANSI block 字 → 高分辨率 Braille ──
# figlet 生成的标准 "Standard" font 字母 7 行高、8 列宽
# 我们用 "-f Bloody" 或 "-f Doom" 更粗
def figlet_text(text, font="Slant"):
    """用 figlet 生成大 ASCII art（没有 figlet 就 fallback 手写 block）"""
    if shutil.which("figlet"):
        result = subprocess.run(
            ["figlet", "-f", font, "-t", text],
            capture_output=True, text=True, timeout=5
        )
        if result.returncode == 0 and result.stdout.strip():
            return result.stdout.rstrip("\n")
    return None

# figlet 字体优先级（从粗到细）
for font in ["Doh", "Basic", "Banner", "Standard", "Slant"]:
    art = figlet_text("GEOYUAN", font)
    if art:
        print(f"✅ figlet font={font}")
        print(art)
        break

# ── 渲染成像素图 → Braille ──
# 如果 figlet 不可用，手写 block pixel
if not art:
    print("figlet 不可用，手写 block pixel 字母")
    W, H, GAP = 10, 14, 3
    
    G = [
        "0111111110","1111111111","1100000011","1100000011","1100111111",
        "1100111111","1100000011","1100000011","1100000011","1111111111","0111111110","0000000000","0000000000","0000000000"]
    E = [
        "1111111110","1111111111","1100000000","1100000000","1111111110",
        "1111111111","1100000000","1100000000","1100000000","1111111111","1111111110","0000000000","0000000000","0000000000"]
    O = [
        "0011111100","0111111110","1100000011","1100000011","1100000011",
        "1100000011","1100000011","1100000011","1100000011","1100000011","0111111110","0011111100","0000000000","0000000000"]
    Y = [
        "1100000011","1100000011","0110000110","0011001100","0001111000",
        "0001111000","0000110000","0000110000","0000110000","0000110000","0000110000","0000110000","0000110000","0000000000"]
    U = [
        "1100000011","1100000011","1100000011","1100000011","1100000011",
        "1100000011","1100000011","1100000011","1100000011","1111111111","0111111110","0000000000","0000000000","0000000000"]
    A = [
        "0001111000","0011001100","0110000110","1100000011","1100000011",
        "1111111111","1100000011","1100000011","1100000011","1100000011","1100000011","1100000011","0000000000","0000000000"]
    N = [
        "1100000011","1110000111","1101001011","1100110011","1100110011",
        "1100011011","1100001011","1100000011","1100000011","1100000011","1100000011","1100000011","0000000000","0000000000"]
    LETTERS = [G, E, O, Y, U, A, N]
    total_w = len(LETTERS) * W + (len(LETTERS) - 1) * GAP
    img = Image.new("RGBA", (total_w, H), (0,0,0,0))
    draw = ImageDraw.Draw(img)
    x = 0
    for li, letter in enumerate(LETTERS):
        color = RAINBOW[li]
        for row, line in enumerate(letter):
            for col, ch in enumerate(line):
                if ch == "1":
                    draw.point((x+col, row), fill=(*color, 255))
        x += W + GAP
    print(f"✅ 手写 pixel: {total_w}×{H}")

else:
    # figlet 生成的 ASCII → 转像素图
    lines = art.split("\n")
    H = len(lines)
    W = max(len(l) for l in lines)
    # 颜色分配：每行按列位置分配彩虹色
    img = Image.new("RGBA", (W, H), (0,0,0,0))
    draw = ImageDraw.Draw(img)
    for r, line in enumerate(lines):
        for c, ch in enumerate(line):
            if ch != " ":
                color = RAINBOW[int(c / W * len(RAINBOW)) % len(RAINBOW)]
                draw.point((c, r), fill=(*color, 255))
    total_w, H = W, H
    print(f"✅ figlet pixel: {W}×{H}")

# ── Braille 转换（放大 3× 让每个 cell 8 点更清晰）──
big = img.resize((img.size[0]*3, img.size[1]*3), Image.NEAREST)
w2, h2 = big.size
px = big.load()

def to_braille_lines(px, w, h):
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

braille_lines = to_braille_lines(px, w2, h2)
print(f"\n📺 Braille {len(braille_lines)} rows × {w2//2} cols")

# ── Half-block 版本 ──
hb_lines = []
prev_fg = prev_bg = (-1,-1,-1)
for y_top in range(0, big.size[1], 2):
    y_bot = min(y_top+1, big.size[1]-1)
    line = ""
    for x in range(big.size[0]):
        pt_t = big.load()[x, y_top]
        pt_b = big.load()[x, y_bot]
        fg = (pt_t[0], pt_t[1], pt_t[2]) if pt_t[3] > 0 else None
        bg = (pt_b[0], pt_b[1], pt_b[2]) if pt_b[3] > 0 else None
        if fg is None and bg is None:
            line += " "
            prev_fg = prev_bg = (-1,-1,-1)
            continue
        esc = ""
        if fg != prev_fg:
            esc += f"\x1b[38;2;{fg[0]};{fg[1]};{fg[2]}m" if fg else "\x1b[39m"
            prev_fg = fg if fg else (-1,-1,-1)
        if bg != prev_bg:
            esc += f"\x1b[48;2;{bg[0]};{bg[1]};{bg[2]}m" if bg else "\x1b[49m"
            prev_bg = bg if bg else (-1,-1,-1)
        line += esc + ("▀" if fg else " ")
    line += "\x1b[0m"
    hb_lines.append(line)

# ── 保存 ──
braille_path = os.path.join(OUT, "geoyuan_logo_braille.txt")
with open(braille_path, "w") as f:
    f.write("\n".join(braille_lines) + "\n")
print(f"✅ Braille: {braille_path}")

hb_path = os.path.join(OUT, "geoyuan_logo_halfblock.txt")
with open(hb_path, "w") as f:
    f.write("\n".join(hb_lines) + "\n")
print(f"✅ Half-block: {hb_path}")

# ── 简单预览（去 ANSI）──
print(f"\n📺 Braille 纯预览:")
for l in braille_lines:
    plain = ''.join(c for c in l if ord(c) > 31)
    print(plain[:80] + ("..." if len(plain) > 80 else ""))

print(f"\n📺 Half-block 纯预览:")
for l in hb_lines:
    plain = ''.join(c for c in l if c not in '\x1b[0123456789;')
    print(plain[:80] + ("..." if len(plain) > 80 else ""))
