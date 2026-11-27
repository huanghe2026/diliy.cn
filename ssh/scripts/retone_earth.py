#!/usr/bin/env python3
"""Geoyuan Braille logo 配色：彩虹 7 色 → 大地色系（地形→地质→深海）"""

import re

# 大地色系（Geoyuan 主题：地形→地质→深海）
# 原始彩虹色 → 新大地色
COLOR_MAP = {
    # 彩虹红 → 地形绿（植被）
    "255;64;128":  "34;197;94",    # ■ green-500
    # 彩虹橙 → 地质橙（土壤）
    "255;140;64":  "249;115;22",   # ■ orange-500
    # 彩虹黄 → 地质金（矿物质）
    "255;220;64":  "234;179;8",    # ■ yellow-500
    # 彩虹绿 → 森林绿（深林）
    "64;255;160":  "16;185;129",   # ■ emerald-500
    # 彩虹青 → 河流蓝
    "64;200;255":  "6;182;212",    # ■ cyan-500
    # 彩虹紫 → 海洋蓝
    "180;100;255": "59;130;246",   # ■ blue-500
    # 彩虹粉 → 深海蓝
    "255;64;200":  "29;78;216",    # ■ blue-700
    # 深灰背景（保留）
    "50;50;70":    "40;44;58",     # 略调深灰适配大地系
}

def convert_logo(path):
    with open(path, "r") as f:
        content = f.read()

    for old, new in COLOR_MAP.items():
        content = content.replace(old, new)

    with open(path, "w") as f:
        f.write(content)

    print(f"✅ {path} converted ({len(COLOR_MAP)} color swaps)")

if __name__ == "__main__":
    base = "/home/huanghe/myweb/diliy.cn/ssh/internal/tui/assets"
    convert_logo(f"{base}/geoyuan_logo_braille.txt")
    convert_logo(f"{base}/geoyuan_logo_halfblock.txt")
    print("\n🌍 Geoyuan Earth-tone palette:")
    print("  ■ green-500  (地形/植被)")
    print("  ■ orange-500 (地质/土壤)")
    print("  ■ yellow-500 (矿物/金)")
    print("  ■ emerald-500(森林)")
    print("  ■ cyan-500   (河流)")
    print("  ■ blue-500   (海洋)")
    print("  ■ blue-700   (深海)")
