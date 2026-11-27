#!/usr/bin/env python3
"""写入 Braille logo — 自动 trim 空行 + 彩虹渐变"""

import os

OUT = os.path.join(os.path.dirname(__file__), "..", "ssh", "internal", "tui", "assets")
os.makedirs(OUT, exist_ok=True)

# 用户提供的原始 Braille — 先 trim 空行
RAW = """⠀⠀⠀⠀⠀⣠⣶⣶⣶⣶⣶⠂⣴⣶⣶⣶⣶⣶⡆⠀⢀⣴⣶⣶⣶⣆⠀⢲⣶⣶⠀⠀⢀⣶⣶⡆⣶⣶⡆⠀⠀⣶⣶⡆⠀⠀⣶⣶⣶⣆⠀⠀⢰⣶⣶⡄⠀⠀⣶⣶⡆⠀⠀⠀
⠀⠀⠀⢠⣾⣿⡿⠛⠛⠛⠋⠀⣿⣿⣿⠛⠛⠛⠃⣴⣿⣿⠛⠛⠻⣿⣷⡄⢿⣿⣧⠀⣼⣿⡟⠀⣿⣿⡇⠀⠀⣿⣿⡇⠀⢰⣿⡟⣿⣿⡀⠀⢸⣿⣿⣿⡄⠀⣿⣿⡇⠀⠀⠀
⠀⠀⠀⢸⣿⣿⠀⢠⣤⣤⣤⡄⣿⣿⣿⣤⣤⣤⠀⣿⣿⡇⠀⠀⠀⣿⣿⡇⠈⢿⣿⣶⣿⡟⠀⠀⣿⣿⡇⠀⠀⣿⣿⡇⠀⣿⣿⠃⢹⣿⣧⠀⢸⣿⣿⢿⣿⡄⣿⣿⡇⠀⠀⠀
⠀⠀⠀⢸⣿⣿⠀⠘⠛⣿⣿⡇⣿⣿⣿⠛⠛⠛⠀⣿⣿⡇⠀⠀⠀⣿⣿⡇⠀⠈⣿⣿⡟⠀⠀⠀⣿⣿⡇⠀⠀⣿⣿⡇⣸⣿⣟⣤⣼⣿⣿⡄⢸⣿⣿⠈⢿⣿⣿⣿⡇⠀⠀⠀
⠀⠀⠀⠘⢿⣿⣷⣤⣴⢿⣿⡇⣿⣿⣿⣤⣤⣤⡄⠻⣿⣿⣦⣤⣾⣿⡿⠃⠀⠀⣿⣿⡇⠀⠀⠀⢿⣿⣷⣤⣴⣿⣿⢃⣿⣿⠟⠛⠛⢻⣿⣧⢸⣿⣿⠀⠈⢿⣿⣿⡇⠀⠀⠀
⠀⠀⠀⠀⠀⠙⠛⠛⠁⠘⠛⠃⠛⠛⠛⠛⠛⠛⠃⠀⠈⠛⠛⠛⠛⠋⠀⠀⠀⠀⠛⠛⠃⠀⠀⠀⠀⠙⠛⠛⠛⠛⠁⠘⠛⠛⠀⠀⠀⠘⠛⠛⠚⠛⠛⠀⠀⠈⠛⠛⠃"""

# 合并多行 + trim 上下空行（只保留有 ⣿⣶⣤⣷ 等非空字符的行）
all_lines = [l.rstrip() for l in RAW.strip().split("\n")]
# trim 左右（如果有）
all_lines = [l.strip() for l in all_lines]
# trim 前后空行
while all_lines and all_lines[0].strip() == "":
    all_lines.pop(0)
while all_lines and all_lines[-1].strip() == "":
    all_lines.pop()

RAINBOW = [
    (255,  64, 128), (255, 140,  64), (255, 220,  64),
    ( 64, 255, 160), ( 64, 200, 255), (180, 100, 255),
    (255,  64, 200),
]

total_w = max(len(l) for l in all_lines)
seg_w = total_w / 7

def colorize(raw_line):
    out = ""
    for i, ch in enumerate(raw_line):
        seg_idx = min(int(i / seg_w), 6)
        color = RAINBOW[seg_idx]
        if ch in ("⠀", " "):
            out += f"\x1b[38;2;50;50;70m⠀\x1b[0m"
        else:
            out += f"\x1b[38;2;{color[0]};{color[1]};{color[2]}m{ch}\x1b[0m"
    return out

colorized = [colorize(l) for l in all_lines]

# 保存
braille_path = os.path.join(OUT, "geoyuan_logo_braille.txt")
with open(braille_path, "w") as f:
    f.write("\n".join(colorized) + "\n")
print(f"✅ Braille trim: {braille_path} ({total_w} cells × {len(colorized)} rows)")

# half-block（简化：每个非空 Braille cell → ▀）
def to_hb(raw_lines):
    out, prev = [], (-1, -1, -1)
    for raw in raw_lines:
        line = ""
        for i, ch in enumerate(raw):
            seg_idx = min(int(i / seg_w), 6)
            color = RAINBOW[seg_idx]
            is_empty = ch in ("⠀", " ")
            if not is_empty:
                if color != prev:
                    line += f"\x1b[38;2;{color[0]};{color[1]};{color[2]}m"
                    prev = color
                line += "▀"
            else:
                if prev != (-1, -1, -1):
                    line += "\x1b[39m"
                    prev = (-1, -1, -1)
                line += " "
        line += "\x1b[0m"
        out.append(line)
    return out

hb_path = os.path.join(OUT, "geoyuan_logo_halfblock.txt")
with open(hb_path, "w") as f:
    f.write("\n".join(to_hb(all_lines)) + "\n")
print(f"✅ Half-block: {hb_path}")

print(f"\n📺 Braille 彩色预览:")
for l in colorized:
    print(l)
print("\x1b[0m")
