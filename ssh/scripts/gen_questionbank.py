#!/usr/bin/env python3
"""预生成 geoyuan 准入测试题库：10 套 × 30 题，保存为 questionbank.json。

与 ssh/internal/questiongen/generator.go 使用相同的 system prompt 与方向，
保证预生成题库与实时 AI 出题质量一致。运行：
    GEOYUAN_LLM_API_KEY=sk-xxx python3 gen_questionbank.py
"""
import json
import os
import sys
import time
import urllib.request
from concurrent.futures import ThreadPoolExecutor, as_completed

API_KEY = os.environ.get("GEOYUAN_LLM_API_KEY", "sk-83e8493a70e249cbb538fd1f022becb4")
BASE_URL = "https://api.deepseek.com"
MODEL = "deepseek-flash"
SETS = 10          # 套数
PER_SET = 30       # 每套题数
WORKERS = 5        # 并发数
EFFORT = "low"     # deepseek-flash 支持 low/high/max，low 更快

SYSTEM_PROMPT = """你是一个 AI/LLM/Agent 技术公司的技术面试官。
请生成高质量的四选一选择题，用于筛选对 AI 工程有真实理解的候选人。

题目覆盖方向（按比例）：
- LLM 基础 & 架构（Transformer、Attention、Tokenization、位置编码、KV Cache、FlashAttention）
- LLM 应用（Prompt Engineering、RAG、向量数据库、Embedding、上下文工程）
- Agent 工程（工具调用、Function Calling、多 Agent 协作、ReAct、Planning）
- LLM 训练（预训练、SFT、RLHF、LoRA/QLoRA、蒸馏、对齐）
- LLM 推理（量化、模型压缩、推理优化、vLLM/TGI、GPU 调度）
- AI 安全与评测（幻觉、越狱、偏见、评测指标）

输出严格为 JSON 数组，每个元素：
{
  "q": "题干",
  "options": ["A选项", "B选项", "C选项", "D选项"],
  "answer": 0,
  "explain": "详细解析"
}
answer 为正确选项索引（0=A, 1=B, 2=C, 3=D）。
只输出 JSON，不要任何其他文字。"""


def gen_set(idx):
    """生成一套 30 题，返回题目列表。"""
    body = {
        "model": MODEL,
        "messages": [
            {"role": "system", "content": SYSTEM_PROMPT},
            {"role": "user", "content": f"请精确生成 {PER_SET} 道题目（数组长度必须等于 {PER_SET}，不能少不能多）。"},
        ],
        "temperature": 0.9,   # 提高温度，让 10 套之间尽量不重复
        "effort": EFFORT,
    }
    req = urllib.request.Request(
        BASE_URL + "/chat/completions",
        data=json.dumps(body).encode(),
        headers={
            "Content-Type": "application/json",
            "Authorization": "Bearer " + API_KEY,
        },
    )
    t0 = time.time()
    with urllib.request.urlopen(req, timeout=120) as r:
        resp = json.loads(r.read().decode())
    dt = time.time() - t0

    content = resp["choices"][0]["message"]["content"]
    content = content.strip().strip("```json").strip("```").strip()
    arr = json.loads(content)

    # 校验并规范化
    out = []
    for it in arr:
        q = it.get("q", "")
        opts = it.get("options", [])
        ans = it.get("answer", -1)
        exp = it.get("explain", "")
        if not q or len(opts) < 4 or not isinstance(ans, int) or not (0 <= ans <= 3):
            continue
        out.append({"q": q, "options": opts[:4], "answer": ans, "explain": exp})
        if len(out) >= PER_SET:
            break

    if len(out) < PER_SET:
        raise RuntimeError(f"set {idx}: only {len(out)} valid questions (need {PER_SET})")
    return idx, dt, out


def main():
    results = []
    with ThreadPoolExecutor(max_workers=WORKERS) as ex:
        futs = [ex.submit(gen_set, i) for i in range(SETS)]
        for fut in as_completed(futs):
            idx, dt, qs = fut.result()
            results.append((idx, qs))
            print(f"set {idx+1}/{SETS}: {len(qs)} 题, {dt:.1f}s", flush=True)

    results.sort(key=lambda x: x[0])
    bank = [qs for _, qs in results]

    out_path = os.path.join(os.path.dirname(__file__), "..", "internal", "questions", "questionbank.json")
    with open(out_path, "w", encoding="utf-8") as f:
        json.dump(bank, f, ensure_ascii=False, indent=1)

    total = sum(len(s) for s in bank)
    print(f"\nOK: {len(bank)} 套, 共 {total} 题 → {out_path}")


if __name__ == "__main__":
    main()
