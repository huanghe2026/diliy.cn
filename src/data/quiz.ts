import type { RoleId } from './team';

export type Category = 'mission' | 'security' | 'role';

export interface QuizItem {
  id: string;
  category: Category;
  /** role this question is most indicative of (used for the final suggestion) */
  role: RoleId;
  q: { en: string; zh: string };
  /** parallel EN/ZH option arrays; `answer` index is the same in both */
  options: { en: string[]; zh: string[] };
  answer: number;
  explain: { en: string; zh: string };
}

// 12 objective questions across three categories. Generated for the Geoyuan
// recruitment challenge. `answer` is the index of the correct option and is
// identical in both language arrays (they are parallel).
export const quiz: QuizItem[] = [
  {
    id: 'q1',
    category: 'mission',
    role: 'psychology',
    q: {
      en: 'In our framing, a “network state” is best described as:',
      zh: '在我们的定义里，"网络国家"最接近：',
    },
    options: {
      en: [
        'A country with clearly defined land borders',
        'A community organised around a shared mission, enabled by the internet rather than geography',
        'Simply a cryptocurrency',
        'A competitive gaming clan',
      ],
      zh: [
        '一个有明确陆地边界的国家',
        '围绕共同使命组织、由互联网而非地理连接的社群',
        '仅仅是一种加密货币',
        '一个电竞战队',
      ],
    },
    answer: 1,
    explain: {
      en: 'We mean a borderless community coordinated by mission and the internet — not territory.',
      zh: '我们指的是由使命与互联网协调、无国界的社群——而非领土。',
    },
  },
  {
    id: 'q2',
    category: 'mission',
    role: 'programmer',
    q: {
      en: 'We prioritise members who:',
      zh: '我们优先看重这样的成员：',
    },
    options: {
      en: [
        'Have the largest social-media following',
        'Live in a specific time zone',
        'Ship, think, and stay curious',
        'Hold a passport from a specific country',
      ],
      zh: [
        '社交媒体粉丝最多',
        '处在某个特定时区',
        '能交付、会思考、保持好奇',
        '持有某个特定国家的护照',
      ],
    },
    answer: 2,
    explain: {
      en: 'Passports and time zones don’t matter to us — output and curiosity do.',
      zh: '护照和时区对我们无关紧要——交付与好奇心才重要。',
    },
  },
  {
    id: 'q3',
    category: 'security',
    role: 'programmer',
    q: {
      en: 'In a CTF, “pwn” usually refers to:',
      zh: '在 CTF 中，"pwn" 通常指：',
    },
    options: {
      en: [
        'Sending phishing emails',
        'Designing brand logos',
        'Exploiting memory-corruption bugs to take control of a program',
        'Writing end-user documentation',
      ],
      zh: [
        '发送钓鱼邮件',
        '设计品牌 logo',
        '利用内存破坏漏洞夺取程序控制权',
        '编写用户文档',
      ],
    },
    answer: 2,
    explain: {
      en: '“pwn” = owning a target, typically via binary/memory exploitation.',
      zh: '"pwn" = 拿下目标，通常靠二进制/内存利用。',
    },
  },
  {
    id: 'q4',
    category: 'security',
    role: 'programmer',
    q: {
      en: 'Cross-site scripting (XSS) is:',
      zh: '跨站脚本（XSS）是：',
    },
    options: {
      en: [
        'A CSS layout framework',
        'Injecting malicious scripts into web pages viewed by others',
        'A relational database engine',
        'A version-control system',
      ],
      zh: [
        '一个 CSS 布局框架',
        '向他人浏览的网页注入恶意脚本',
        '一种关系型数据库引擎',
        '一个版本控制系统',
      ],
    },
    answer: 1,
    explain: {
      en: 'XSS lets an attacker run script in a victim’s browser via a vulnerable page.',
      zh: 'XSS 让攻击者借漏洞页面在受害者浏览器中执行脚本。',
    },
  },
  {
    id: 'q5',
    category: 'security',
    role: 'analyst',
    q: {
      en: 'Why might we deliberately avoid keeping access logs?',
      zh: '我们为何可能刻意不保留访问日志？',
    },
    options: {
      en: [
        'Logs are only expensive to store',
        'We don’t actually run any servers',
        'It is illegal to log anything',
        'Privacy & opsec — less data means less to leak',
      ],
      zh: [
        '日志只贵在存储',
        '我们其实根本不跑服务器',
        '记录任何东西都违法',
        '隐私与运营安全——数据越少，可泄露的越少',
      ],
    },
    answer: 3,
    explain: {
      en: 'A core opsec principle: minimise what could be exposed if a system is compromised.',
      zh: '核心运营安全原则：尽量减少系统被攻破时可能暴露的内容。',
    },
  },
  {
    id: 'q6',
    category: 'security',
    role: 'ai',
    q: {
      en: 'RAG (retrieval-augmented generation) means:',
      zh: 'RAG（检索增强生成）是指：',
    },
    options: {
      en: [
        'Grounding a model’s answers in retrieved documents',
        'Deleting a trained model',
        'A specific type of GPU',
        'A programming language',
      ],
      zh: [
        '用检索到的文档为模型回答提供依据',
        '删除一个训练好的模型',
        '一种特定型号的 GPU',
        '一门编程语言',
      ],
    },
    answer: 0,
    explain: {
      en: 'RAG fetches relevant context so the model answers from real sources, not just weights.',
      zh: 'RAG 先检索相关上下文，让模型基于真实来源作答，而非仅凭权重。',
    },
  },
  {
    id: 'q7',
    category: 'role',
    role: 'artist',
    q: {
      en: 'Good visual identity for a hacker collective should:',
      zh: '对黑客团体而言，好的视觉识别应当：',
    },
    options: {
      en: [
        'Directly copy a big brand’s look',
        'Be legible, consistent, and communicate the right tone',
        'Use as many fonts as possible',
        'Avoid dark mode entirely',
      ],
      zh: [
        '直接照搬大品牌的样式',
        '清晰、一致，并能传达正确的调性',
        '用尽可能多的字体',
        '彻底不用暗色模式',
      ],
    },
    answer: 1,
    explain: {
      en: 'Identity is about clarity and tone, not maximal decoration.',
      zh: '识别度在于清晰与调性，而非堆砌装饰。',
    },
  },
  {
    id: 'q8',
    category: 'role',
    role: 'ai',
    q: {
      en: 'An “agent” in AI usually means:',
      zh: 'AI 中的"智能体（agent）"通常指：',
    },
    options: {
      en: [
        'A human personal assistant',
        'A chatbot that only replies to prompts',
        'A system that plans and acts with tools over multiple steps',
        'A database index',
      ],
      zh: [
        '一个人类私人助理',
        '只会回复提示的聊天机器人',
        '能规划并借助工具分步行动的系统',
        '一个数据库索引',
      ],
    },
    answer: 2,
    explain: {
      en: 'Agents loop: reason → use a tool → observe → repeat until the task is done.',
      zh: '智能体会循环：推理→用工具→观察→重复，直到任务完成。',
    },
  },
  {
    id: 'q9',
    category: 'role',
    role: 'analyst',
    q: {
      en: 'A “conversion funnel” helps you understand:',
      zh: '"转化漏斗"帮你理解：',
    },
    options: {
      en: [
        'The weather forecast',
        'CPU utilisation',
        'How people move from first interest to a real action',
        'Proper font sizing',
      ],
      zh: [
        '天气预报',
        'CPU 使用率',
        '人们如何从初次感兴趣走到真正行动',
        '正确的字号',
      ],
    },
    answer: 2,
    explain: {
      en: 'Funnels show where interested people drop off before they join or act.',
      zh: '漏斗展示感兴趣的人在哪一步流失，未能加入或行动。',
    },
  },
  {
    id: 'q10',
    category: 'role',
    role: 'psychology',
    q: {
      en: 'When screening new members, the most useful signal is usually:',
      zh: '在筛选新成员时，最有用的信号通常是：',
    },
    options: {
      en: [
        'Their exact age',
        'Their physical location',
        'Demonstrated motivation and fit over time',
        'Their self-rated skill score',
      ],
      zh: [
        '他们的确切年龄',
        '他们的物理位置',
        '长期展现出的动机与适配度',
        '他们自评的技能分数',
      ],
    },
    answer: 2,
    explain: {
      en: 'Sustained demonstrated behaviour beats self-reports and demographics.',
      zh: '持续展现出的行为，优于自评与人口统计。',
    },
  },
  {
    id: 'q11',
    category: 'role',
    role: 'programmer',
    q: {
      en: 'Infrastructure as code (IaC) means:',
      zh: '"基础设施即代码"（IaC）是指：',
    },
    options: {
      en: [
        'Drawing servers on a whiteboard',
        'Managing infrastructure through versioned config files',
        'A new data-protection law',
        'A kind of coffee machine',
      ],
      zh: [
        '在白板上画服务器',
        '用纳入版本管理的配置文件来管理基础设施',
        '一部新的数据保护法律',
        '一种咖啡机',
      ],
    },
    answer: 1,
    explain: {
      en: 'IaC makes servers reproducible and reviewable like any other code.',
      zh: 'IaC 让服务器像普通代码一样可复现、可评审。',
    },
  },
  {
    id: 'q12',
    category: 'mission',
    role: 'psychology',
    q: {
      en: 'If you don’t pass a challenge here, you should:',
      zh: '如果你没通过这里的挑战，应当：',
    },
    options: {
      en: [
        'Give up entirely',
        'Create alt accounts to retry',
        'Argue with the graders',
        'Learn, come back sharper, and contribute where you can',
      ],
      zh: [
        '彻底放弃',
        '开小号重来',
        '和评分者争论',
        '学习、变得更强再来，并在能处贡献力量',
      ],
    },
    answer: 3,
    explain: {
      en: 'A challenge is a filter, not a wall. Growth and contribution beat gaming the system.',
      zh: '挑战是过滤器，不是墙。成长与贡献胜过钻空子。',
    },
  },
];
