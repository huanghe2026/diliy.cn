#!/usr/bin/env python3
"""写入用户提供的 Braille GEOYUAN + 彩虹渐变色 + 同步生成 halfblock"""

import os

OUT = os.path.join(os.path.dirname(__file__), "..", "ssh", "internal", "tui", "assets")
os.makedirs(OUT, exist_ok=True)

# ── 用户提供的原始 Braille art（6 行）──
BRAILLE_LINES_RAW = [
    "⠀⠀⠀⠀⠀⣠⣶⣶⣶⣶⣶⠂⣴⣶⣶⣶⣶⣶⡆⠀⢀⣴⣶⣶⣶⣆⠀⢲⣶⣶⠀⠀⢀⣶⣶⡆⣶⣶⡆⠀⠀⣶⣶⡆⠀⠀⣶⣶⣶⣆⠀⠀⢰⣶⣶⡄⠀⠀⣶⣶⡆⠀⠀⠀",
    "⠀⠀⠀⢠⣾⣿⡿⠛⠛⠛⠋⠀⣿⣿⣿⠛⠛⠛⠃⣴⣿⣿⠛⠛⠻⣿⣷⡄⢿⣿⣧⠀⣼⣿⡟⠀⣿⣿⡇⠀⠀⣿⣿⡇⠀⢰⣿⡟⣿⣿⡀⠀⢸⣿⣿⣿⡄⠀⣿⣿⡇⠀⠀⠀",
    "⠀⠀⠀⢸⣿⣿⠀⢠⣤⣤⣤⡄⣿⣿⣿⣤⣤⣤⠀⣿⣿⡇⠀⠀⠀⣿⣿⡇⠈⢿⣿⣶⣿⡟⠀⠀⣿⣿⡇⠀⠀⣿⣿⡇⠀⣿⣿⠃⢹⣿⣧⠀⢸⣿⣿⢿⣿⡄⣿⣿⡇⠀⠀⠀",
    "⠀⠀⠀⢸⣿⣿⠀⠘⠛⣿⣿⡇⣿⣿⣿⠛⠛⠛⠀⣿⣿⡇⠀⠀⠀⣿⣿⡇⠀⠈⣿⣿⡟⠀⠀⠀⣿⣿⡇⠀⠀⣿⣿⡇⣸⣿⣟⣤⣼⣿⣿⡄⢸⣿⣿⠈⢿⣿⣿⣿⡇⠀⠀⠀",
    "⠀⠀⠀⠘⢿⣿⣷⣤⣴⢿⣿⡇⣿⣿⣿⣤⣤⣤⡄⠻⣿⣿⣦⣤⣾⣿⡿⠃⠀⠀⣿⣿⡇⠀⠀⠀⢿⣿⣷⣤⣴⣿⣿⢃⣿⣿⠟⠛⠛⢻⣿⣧⢸⣿⣿⠀⠈⢿⣿⣿⡇⠀⠀⠀",
    "⠀⠀⠀⠀⠀⠙⠛⠛⠁⠘⠛⠃⠛⠛⠛⠛⠛⠛⠃⠀⠈⠛⠛⠛⠛⠋⠀⠀⠀⠀⠀⠛⠛⠃⠀⠀⠀⠀⠙⠛⠛⠛⠛⠁⠘⠛⠛⠀⠀⠀⠘⠛⠛⠚⠛⠛⠀⠀⠈⠛⠛⠃",
]

# ── 彩虹色（7 字母）──
RAINBOW = [
    (255,  64, 128),  # G
    (255, 140,  64),  # E
    (255, 220,  64),  # O
    ( 64, 255, 160),  # Y
    ( 64, 200, 255),  # U
    (180, 100, 255),  # A
    (255,  64, 200),  # N
]

# ── 按字符位置分 7 段（每字母 ~19 cells）──
# 计算总宽后平均分 7 段
total_w = max(len(l) for l in BRAILLE_LINES_RAW)
seg_w = total_w / 7

def colorize_braille_line(raw_line):
    """给一行 Braille 逐字符加彩虹渐变色"""
    out = ""
    for i, ch in enumerate(raw_line):
        # 判断所属字母段
        seg_idx = min(int(i / seg_w), 6)
        color = RAINBOW[seg_idx]
        # 空格用深灰（淡色背景感）
        if ch == "⠀" or ch == " ":
            out += f"\x1b[38;2;60;60;80m{ch}\x1b[0m"
        else:
            out += f"\x1b[38;2;{color[0]};{color[1]};{color[2]}m{ch}\x1b[0m"
    return out

colorized = [colorize_braille_line(l) for l in BRAILLE_LINES_RAW]

# ── 保存 Braille ──
braille_path = os.path.join(OUT, "geoyuan_logo_braille.txt")
with open(braille_path, "w") as f:
    f.write("\n".join(colorized) + "\n")
print(f"✅ Braille: {braille_path} ({total_w} cells × 6 rows)")

# ── 同步生成 Half-block（用同样的字符分段逻辑，只是把 Braille cell 换成 ▀）──
# 方法：每个 Braille cell 对应 half-block 的一个 ▀ cell
# 颜色分配：同上 7 段彩虹
# 先把 Braille cell 映射到 half-block

def to_halfblock_from_braille(raw_lines):
    """把 Braille art 转成 half-block ▀"""
    out = []
    prev_fg = prev_bg = (-1, -1, -1)
    for raw_line in raw_lines:
        line = ""
        for i, ch in enumerate(raw_line):
            seg_idx = min(int(i / seg_w), 6)
            color = RAINBOW[seg_idx]
            is_empty = ch == "⠀" or ch == " "
            
            # half-block: 一个 Braille cell → 一个 ▀
            # Braille 有 8 点密度信息，但我们简化：
            #   cell 非空 → ▀ 前景色 = 字母色，背景透明
            #   cell 空   → 空格
            if not is_empty:
                fg = color
                esc = ""
                if fg != prev_fg:
                    esc += f"\x1b[38;2;{fg[0]};{fg[1]};{fg[2]}m"
                    prev_fg = fg
                line += esc + "▀"
            else:
                if prev_fg != (-1, -1, -1):
                    line += "\x1b[39m"
                    prev_fg = (-1, -1, -1)
                line += " "
        line += "\x1b[0m"
        out.append(line)
    return out

hb_lines = to_halfblock_from_braille(BRAILLE_LINES_RAW)
hb_path = os.path.join(OUT, "geoyuan_logo_halfblock.txt")
with open(hb_path, "w") as f:
    f.write("\n".join(hb_lines) + "\n")
print(f"✅ Half-block: {hb_path} ({total_w} cells × 6 rows)")

# ── 预览 ──
print(f"\n📺 Braille 彩色预览:")
for l in colorized:
    print(l)
print("\x1b[0m")
