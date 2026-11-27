#!/usr/bin/env python3
"""从用户上传的 txt 读 Braille → trim 空行 + 彩虹渐变 → assets"""

import os

SRC = "/home/huanghe/.trae-cn/attachments/3331facc-0261-4313-a1f6-3a41458ce691_adf1a441-c828-40e1-b4a8-deea40e1a060_copy2paste-braille-art.txt"
OUT = os.path.join(os.path.dirname(__file__), "..", "ssh", "internal", "tui", "assets")
os.makedirs(OUT, exist_ok=True)

# 读源文件
with open(SRC, "r", encoding="utf-8") as f:
    raw = f.read()

# 按行 + rstrip + trim 前后空行
lines = [l.rstrip() for l in raw.strip().split("\n")]
while lines and all(ch in ("⠀", " ") for ch in lines[0].strip()):
    lines.pop(0)
while lines and all(ch in ("⠀", " ") for ch in lines[-1].strip()):
    lines.pop()

print(f"📜 源: {len(lines)} 行有效 Braille")

# 彩虹 7 色（G-E-O-Y-U-A-N 按位置分段）
RAINBOW = [
    (255,  64, 128),  # 玫红 G
    (255, 140,  64),  # 橙 E
    (255, 220,  64),  # 黄 O
    ( 64, 255, 160),  # 绿 Y
    ( 64, 200, 255),  # 青 U
    (180, 100, 255),  # 紫 A
    (255,  64, 200),  # 粉 N
]

total_w = max(len(l) for l in lines)
seg_w = max(total_w / 7, 1)

def colorize(line):
    out = ""
    for i, ch in enumerate(line):
        seg_idx = min(int(i / seg_w), 6)
        r, g, b = RAINBOW[seg_idx]
        if ch in ("⠀", " "):
            out += f"\x1b[38;2;50;50;70m⠀\x1b[0m"
        else:
            out += f"\x1b[38;2;{r};{g};{b}m{ch}\x1b[0m"
    return out

colorized = [colorize(l) for l in lines]

# 保存 braille
braille_path = os.path.join(OUT, "geoyuan_logo_braille.txt")
with open(braille_path, "w") as f:
    f.write("\n".join(colorized) + "\n")
print(f"✅ Braille: {braille_path} ({total_w}×{len(colorized)})")

# half-block 降级
def to_hb(line):
    out, prev = "", (-1, -1, -1)
    for i, ch in enumerate(line):
        seg_idx = min(int(i / seg_w), 6)
        r, g, b = RAINBOW[seg_idx]
        is_empty = ch in ("⠀", " ")
        if not is_empty:
            if (r, g, b) != prev:
                out += f"\x1b[38;2;{r};{g};{b}m"
                prev = (r, g, b)
            out += "▀"
        else:
            if prev != (-1, -1, -1):
                out += "\x1b[39m"
                prev = (-1, -1, -1)
            out += " "
    out += "\x1b[0m"
    return out

hb_path = os.path.join(OUT, "geoyuan_logo_halfblock.txt")
with open(hb_path, "w") as f:
    f.write("\n".join(to_hb(l) for l in lines) + "\n")
print(f"✅ Half-block: {hb_path}")

print(f"\n📺 预览 ({total_w} cells × {len(colorized)} rows):")
for l in colorized:
    print(l)
print("\x1b[0m")
