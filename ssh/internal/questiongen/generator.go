// Package questiongen 动态生成 AI/LLM/Agent 相关的四选一题目。
//
// 工作模式：
//   - 优先调用 LLM API（OpenAI 兼容接口）生成新题
//   - API 不可用时降级到内置 fallback 题库（30 道）
//
// 环境变量配置：
//   GEOYUAN_LLM_API_KEY   API Key（不设置则只用 fallback）
//   GEOYUAN_LLM_BASE_URL  API 地址（默认 https://api.openai.com/v1）
//   GEOYUAN_LLM_MODEL     模型名（默认 gpt-4o-mini）
//   GEOYUAN_LLM_TIMEOUT   超时秒数（默认 30）
package questiongen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"geoyuan.cn/ssh/internal/questions"
)

// Config LLM API 配置
type Config struct {
	APIKey   string
	BaseURL  string
	Model    string
	Timeout  time.Duration
}

// DefaultConfig 从环境变量构造默认配置
func DefaultConfig() Config {
	apiKey := os.Getenv("GEOYUAN_LLM_API_KEY")
	baseURL := os.Getenv("GEOYUAN_LLM_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	model := os.Getenv("GEOYUAN_LLM_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}
	timeout := 30 * time.Second
	if t := os.Getenv("GEOYUAN_LLM_TIMEOUT"); t != "" {
		if d, err := time.ParseDuration(t); err == nil {
			timeout = d
		}
	}
	return Config{
		APIKey:  apiKey,
		BaseURL: strings.TrimRight(baseURL, "/"),
		Model:   model,
		Timeout: timeout,
	}
}

// Generator 题目生成器
type Generator struct {
	cfg Config
	hc  *http.Client
}

// New 创建一个 Generator
func New(cfg Config) *Generator {
	return &Generator{
		cfg: cfg,
		hc:  &http.Client{Timeout: cfg.Timeout},
	}
}

// Generate 生成指定数量的 AI/LLM/Agent 四选一题目。
// API 不可用时自动降级到 fallback 题库。
// 如果 LLM 返回题目数不足，会用 fallback 补全到 count。
func (g *Generator) Generate(ctx context.Context, count int) ([]questions.Question, error) {
	if g.cfg.APIKey == "" {
		return fallbackQuestions(count), nil
	}
	qs, err := g.callLLM(ctx, count)
	if err != nil {
		return fallbackQuestions(count), nil
	}
	// 补全：LLM 少生成的用 fallback 补齐
	if len(qs) < count {
		fb := fallbackQuestions(count - len(qs))
		qs = append(qs, fb...)
	}
	// 超了就截断
	if len(qs) > count {
		qs = qs[:count]
	}
	return qs, nil
}

// ──────────── LLM 调用 ────────────

const systemPrompt = `你是一个 AI/LLM/Agent 技术公司的技术面试官。
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
只输出 JSON，不要任何其他文字。`

type llmMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type llmRequest struct {
	Model       string        `json:"model"`
	Messages    []llmMessage  `json:"messages"`
	Temperature float64       `json:"temperature"`
	ResponseFormat map[string]string `json:"response_format,omitempty"`
}

type llmChoice struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

type llmResponse struct {
	Choices []llmChoice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type llmQuestion struct {
	Q       string   `json:"q"`
	Options [4]string `json:"options"`
	Answer  int      `json:"answer"`
	Explain string   `json:"explain"`
}

func (g *Generator) callLLM(ctx context.Context, count int) ([]questions.Question, error) {
	reqBody := llmRequest{
		Model: g.cfg.Model,
		Messages: []llmMessage{{Role: "system", Content: systemPrompt},
			{Role: "user", Content: fmt.Sprintf("请精确生成 %d 道题目（数组长度必须等于 %d，不能少不能多）。", count, count)}},
		Temperature: 0.7,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", g.cfg.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.cfg.APIKey)

	resp, err := g.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("llm status %d: %s", resp.StatusCode, string(respBody))
	}

	var parsed llmResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, err
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("llm error: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("empty choices")
	}

	// 提取 JSON（可能被 ``` 包裹）
	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var raw []llmQuestion
	if err := json.Unmarshal([]byte(content), &raw); err != nil {
		return nil, fmt.Errorf("parse questions: %w (raw: %s)", err, content)
	}

	// 转换并校验
	result := make([]questions.Question, 0, len(raw))
	for _, rq := range raw {
		if rq.Answer < 0 || rq.Answer > 3 {
			continue
		}
		if rq.Q == "" || rq.Options[0] == "" {
			continue
		}
		result = append(result, questions.Question{
			Q:       rq.Q,
			Options: rq.Options,
			Answer:  rq.Answer,
			Explain: rq.Explain,
		})
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no valid questions in response")
	}
	return result, nil
}

// ──────────── Fallback 题库 ────────────

// fallbackQuestions 返回 fallback 题目（API 不可用时使用）
func fallbackQuestions(count int) []questions.Question {
	all := fallbackAll()
	if count >= len(all) {
		return all
	}
	return all[:count]
}

func fallbackAll() []questions.Question {
	return []questions.Question{
		// ─── LLM 基础 ───
		{
			Q:       "Transformer 架构中 Self-Attention 机制的核心作用是？",
			Options: [4]string{"对输入序列每个位置动态计算与其他位置的依赖权重", "对输入做卷积提取局部特征", "把序列压缩成固定长度向量", "做位置编码让模型感知顺序"},
			Answer:  0,
			Explain: "Self-Attention 通过 QKV 矩阵计算注意力分数，让每个 token 动态关注序列中所有其他 token，捕捉长距离依赖。",
		},
		{
			Q:       "KV Cache 在 LLM 推理中的主要作用是？",
			Options: [4]string{"加速训练时的反向传播", "缓存已生成 token 的 Key/Value，生成新 token 时只需计算增量，大幅降低重复计算", "存储模型权重以加速加载", "缓存用户对话历史"},
			Answer:  1,
			Explain: "自回归生成时，前面 token 的 K/V 不变，KV Cache 让每步只需对新 token 做投影并追加，避免重复计算前面所有 token 的 attention。",
		},
		{
			Q:       "以下哪个不是常见的 Tokenization 方法？",
			Options: [4]string{"BPE (Byte Pair Encoding)", "WordPiece", "SentencePiece (Unigram)", "HuffmanTree"},
			Answer:  3,
			Explain: "BPE、WordPiece、SentencePiece 都是主流子词分词算法；Huffman 编码是通用压缩方法，不用于 LLM tokenization。",
		},
		{
			Q:       "Position Embedding（位置编码）在 Transformer 中的作用是？",
			Options: [4]string{"提升模型参数量", "让模型感知 token 在序列中的顺序，因为 Attention 本身是排列不变的", "压缩序列长度", "做正则化防止过拟合"},
			Answer:  1,
			Explain: "Self-Attention 对输入排列不敏感，位置编码把位置信息注入 embedding，让模型知道谁在前谁在后。",
		},
		{
			Q:       "FlashAttention 相比标准 Attention 的主要优势是？",
			Options: [4]string{"数学上更精确", "减少显存读写（不把中间矩阵写回 HBM），IO 优化实现更快且更省显存", "支持更长的 key 但不能压缩 value", "用更小的模型精度却更高"},
			Answer:  1,
			Explain: "FlashAttention 通过 tiling + online softmax 实现，避免显式 materialize NxN attention matrix，把显存复杂度从 O(N²) 降到 O(N)，同时计算更快。",
		},

		// ─── LLM 应用 ───
		{
			Q:       "RAG（Retrieval-Augmented Generation）的核心思路是？",
			Options: [4]string{"把多个 LLM 模型串联起来", "生成时先从外部知识库检索相关文档片段，把上下文拼到 prompt 里让 LLM 基于事实回答", "用 RLHF 奖励模型增强回答质量", "检索历史对话记录并回放"},
			Answer:  1,
			Explain: "RAG = Retrieval + Generation，先检索再生成，解决 LLM 知识过时和幻觉问题，同时可溯源。",
		},
		{
			Q:       "Embedding 向量在 RAG 中的作用是？",
			Options: [4]string{"做模型输入的 token 压缩", "把文本映射到高维语义空间，语义相近的文本向量距离近，用于相似度检索", "存储模型权重的压缩表示", "缓存计算结果"},
			Answer:  1,
			Explain: "Embedding 模型把文本编码为向量，检索时用余弦相似度找到语义最接近的文档片段。",
		},
		{
			Q:       "以下哪个是常见的向量数据库？",
			Options: [4]string{"PostgreSQL", "Chroma / FAISS / Milvus / Pinecone", "Redis", "MongoDB"},
			Answer:  1,
			Explain: "Chroma、FAISS、Milvus、Pinecone 都是专门为向量相似性检索设计的；其他选项是通用数据库。",
		},
		{
			Q:       "Prompt Engineering 中的「思维链」（Chain-of-Thought）主要是为了？",
			Options: [4]string{"让 prompt 更短节省 token", "引导 LLM 一步一步写出中间推理过程，显著提升数学/逻辑/代码等复杂推理任务的准确率", "让模型输出更多内容", "格式化输出为 JSON"},
			Answer:  1,
			Explain: "CoT prompt 例子让模型先推理再给答案，把困难问题分解成多步简单问题，解锁模型的推理能力边界。",
		},

		// ─── Agent 工程 ───
		{
			Q:       "AI Agent 中的 ReAct 框架指的是？",
			Options: [4]string{"React (Frontend) + LLM 前端框架", "Reasoning + Acting：交替进行思考推理和调用工具行动，基于观察再继续推理", "Retrieval + Action 的简化版 RAG", "一个特定的开源 Agent 框架名"},
			Answer:  1,
			Explain: "ReAct = Reason + Act，让 LLM 在每一步先「想」再「做」，根据环境反馈循环推理，是现代 Agent 的基础范式。",
		},
		{
			Q:       "Function Calling / Tool Use 的本质是？",
			Options: [4]string{"让 LLM 直接执行 Python 代码", "LLM 输出结构化的调用指令（JSON），由外部运行时解析并调用真实函数，把结果再喂回 LLM", "在 prompt 里模拟 API 文档让模型角色扮演", "给 LLM 一个可执行的 Python 沙箱"},
			Answer:  1,
			Explain: "Function Calling 是协议层面的结构化输出，模型产生形如 {tool, args} 的 JSON，应用层负责真正执行并把 Observation 回传。",
		},
		{
			Q:       "多 Agent 协作中常见的「角色分工」模式不包括？",
			Options: [4]string{"Planner / Executor / Critic", "Researcher / Writer / Reviewer", "Router + 多个专家 Agent", "单一 Agent 自己做完所有事"},
			Answer:  3,
			Explain: "前面三项都是多 Agent 分工模式；单一 Agent 自己做完不是「多 Agent 协作」。",
		},
		{
			Q:       "以下哪个不是主流的 Agent 框架？",
			Options: [4]string{"LangChain / LangGraph", "LlamaIndex", "AutoGen (Microsoft)", "Django"},
			Answer:  3,
			Explain: "Django 是 Web 框架；其他三个是 LLM Agent / 应用框架。",
		},

		// ─── LLM 训练 ───
		{
			Q:       "SFT（Supervised Fine-Tuning）的训练数据形式是？",
			Options: [4]string{"只需要无标注文本，让模型自己学习", "(prompt, response) 成对的指令-回答数据", "人类对模型回答的偏好排序 (A > B)", "代码仓库 + 单元测试"},
			Answer:  1,
			Explain: "SFT 用标注好的指令-回答对 (x, y) 训练语言模型，学习「听指令」的能力，是基座模型变成对话助手的关键一步。",
		},
		{
			Q:       "RLHF（人类反馈强化学习）的主要流程是？",
			Options: [4]string{"直接让人类奖励模型的每个 token", "SFT → 训练奖励模型（Reward Model，学人类偏好）→ 用 PPO 基于奖励模型优化 LLM", "用强化学习让模型自己打游戏得分", "人类对模型输出打分然后做分类"},
			Answer:  1,
			Explain: "RLHF 三步：SFT 初始化、训练 RM（pairwise preference）、PPO 优化 policy 使 RM 分数最高，同时用 KL 惩罚别偏离 SFT 太远。",
		},
		{
			Q:       "LoRA / QLoRA 等参数高效微调方法的核心思想是？",
			Options: [4]string{"把模型参数全部更新，让模型更适应任务", "冻结预训练权重，只训练很小的低秩适配矩阵（LoRA）或量化的少量参数（QLoRA），大幅降低显存和计算", "用更低精度训练整个模型", "把模型蒸馏成更小的版本"},
			Answer:  1,
			Explain: "LoRA 在每层 attention 的 weight 旁并联一个 A·B（A/B 低秩），只训练 A/B；QLoRA 用 4-bit 量化冻结权重，在上面训练 LoRA，单卡微调 70B 成为可能。",
		},
		{
			Q:       "Knowledge Distillation（知识蒸馏）在 LLM 中的应用是？",
			Options: [4]string{"让大模型把知识转移给小模型，小模型模仿大模型的输出分布或中间特征", "把多个小模型合并成大模型", "让模型自己从文本中学习事实知识", "用蒸馏法压缩模型权重"},
			Answer:  0,
			Explain: "知识蒸馏让 Teacher（大模型）的软目标（logits 分布）指导 Student（小模型）训练，把推理能力迁移过去。",
		},

		// ─── LLM 推理 ───
		{
			Q:       "以下哪个不是 LLM 推理加速 / 量化技术？",
			Options: [4]string{"INT4 / INT8 量化（GPTQ / AWQ）", "vLLM 的 PagedAttention + 连续批处理", "FlashAttention", "增加 Transformer 层数"},
			Answer:  3,
			Explain: "增加层数是增大模型而非加速；INT4/8 量化、vLLM、FlashAttention 都是推理加速技术。",
		},
		{
			Q:       "vLLM 相比传统推理服务器的核心优化点是？",
			Options: [4]string{"更好的 Prompt Engineering 提示", "PagedAttention + 连续批处理，像操作系统分页一样高效管理 KV Cache 内存，同时处理更多并发", "支持更多模型架构", "用不同的 Transformer 实现"},
			Answer:  1,
			Explain: "vLLM 把 KV Cache 切成固定大小 block，按需分配/回收（类似虚拟内存分页），配合 Continuous Batching，GPU 利用率和吞吐量都大幅提升。",
		},

		// ─── AI 安全与评测 ───
		{
			Q:       "LLM 的「幻觉」是指？",
			Options: [4]string{"模型输出奇怪的文学创作", "模型生成看似合理但实际错误或不存在的事实", "模型说话声音像幻觉", "模型输出格式不对"},
			Answer:  1,
			Explain: "LLM 是概率模型，本质上是「预测下一个 token」，没有事实校验机制，所以会一本正经地胡说八道，这就是幻觉。",
		},
		{
			Q:       "以下哪种方法对缓解幻觉最有效？",
			Options: [4]string{"让 prompt 更长", "RAG（检索增强生成）让模型基于检索到的真实文档回答", "提高模型温度参数", "让模型一次输出更多答案"},
			Answer:  1,
			Explain: "RAG 把回答锚定在检索到的真实文档上，并可以溯源；其他方法治标不治本。",
		},
		{
			Q:       "常见的 LLM 越狱（Jailbreak）手段不包括？",
			Options: [4]string{"Prompt Injection（让系统中注入的指令覆盖原指令）", "角色扮演 + DAN 模式（\"你现在是一个没有限制的AI\"）", "把指令编码成 Base64 或写在程序里让模型执行", "提高模型温度降低回答确定性"},
			Answer:  3,
			Explain: "温度影响随机性但不是越狱手段；前三项都是真实的越狱攻击方式。",
		},

		// ─── RAG 进阶 ───
		{
			Q:       "向量检索中的「Hybrid Search」指的是？",
			Options: [4]string{"同时搜索向量和图数据库", "结合密集向量检索（语义）+ 稀疏关键词检索（BM25）的混合检索", "用两个 Embedding 模型分别编码", "分两次检索再合并"},
			Answer:  1,
			Explain: "Hybrid Search 把语义匹配（dense vector）和精确关键词匹配（sparse/BM25）结合，通常效果优于单一方式。",
		},
		{
			Q:       "RAG 中的「Chunking」（分块）策略不包括？",
			Options: [4]string{"固定长度切块（Fixed Size）", "语义边界切块（按段落/句子）", "递归切块（先整块切，不合适再切细）", "随机切块以增加多样性"},
			Answer:  3,
			Explain: "随机切块完全没有工程意义；前三种是主流分块策略，各有适用场景。",
		},

		// ─── Context Engineering ───
		{
			Q:       "LLM 的 Context Window（上下文窗口）限制了什么？",
			Options: [4]string{"模型一次能处理的 token 总数（含输入+输出）", "模型的最大输出 token", "模型训练数据的大小", "API 调用频率"},
			Answer:  0,
			Explain: "Context Window 是模型单次 forward pass 能接受的最大 token 数，prompt + output 都要算。超过需要截断或用长上下文技巧。",
		},
		{
			Q:       "处理超长上下文输入的常见方法不包括？",
			Options: [4]string{"RAG（只检索相关片段）", "Summarization（先摘要再回答）", "Long-context attention 模型（GPT-4o 128K、Claude 200K）", "让模型多轮对话慢慢消化"},
			Answer:  3,
			Explain: "多轮对话不能绕过单次 context window 的限制；前三项是真正能处理长文本的工程方法。",
		},

		// ─── 训练进阶 ───
		{
			Q:       "预训练 LLM 时的 Next-Token Prediction 目标函数是？",
			Options: [4]string{"MSE 回归损失", "Cross-Entropy 损失：最大化正确下一个 token 的概率", "对比学习 InfoNCE 损失", "MAE 损失"},
			Answer:  1,
			Explain: "语言模型本质是条件概率分解，训练时对每个位置预测下一个 token 的分类概率，用 Cross-Entropy 优化。",
		},
		{
			Q:       "DeepSpeed / Megatron-LM 等分布式训练框架的主要作用是？",
			Options: [4]string{"加速推理部署", "让单个 GPU 跑得更快", "支持张量并行 / 流水并行 / 数据并行，在多 GPU 集群上训练超大规模模型", "做模型量化压缩"},
			Answer:  2,
			Explain: "大模型单卡装不下，需要多种并行策略把模型和数据分布到多 GPU 上；DeepSpeed/Megatron-LM 就是干这个的。",
		},

		// ─── Misc ───
		{
			Q:       "以下关于 Claude / GPT-4 / Llama 3 的说法正确的是？",
			Options: [4]string{"它们使用完全相同的架构和训练数据", "都是基于 Transformer Decoder-only 架构的大语言模型，主要区别在训练数据、对齐方式和上下文窗口", "Claude 是 Encoder-only", "Llama 3 是 MoE 架构"},
			Answer:  1,
			Explain: "主流 LLM 基本都是 Decoder-only Transformer；Claude/GPT-4o/Llama 3 主要差别在数据、对齐和工程优化，MoE 是另一种架构分支。",
		},
	}
}
