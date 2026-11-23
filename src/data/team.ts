export type RoleId = 'programmer' | 'artist' | 'ai' | 'analyst' | 'psychology';

export interface Role {
  id: RoleId;
  avatar: string;
  name: { en: string; zh: string };
  role: { en: string; zh: string };
  bio: { en: string; zh: string };
  looking: { en: string; zh: string };
}

export const roles: Role[] = [
  {
    id: 'programmer',
    avatar: 'PR',
    name: { en: 'Programmer', zh: '程序员' },
    role: { en: 'builds / maintains', zh: '开发 / 维护' },
    bio: {
      en: 'Owns the site, our internal tooling, and the infrastructure that keeps the network state standing. Ships features and automates the boring parts.',
      zh: '负责站点、内部工具，以及维持网络国家运转的基础设施。交付功能，把无聊的事自动化。',
    },
    looking: {
      en: 'Full-stack or systems mindset. You’ve deployed something real and can read a stack trace at 3 AM.',
      zh: '全栈或系统思维。你真正上线过东西，能在凌晨三点读得懂堆栈跟踪。',
    },
  },
  {
    id: 'artist',
    avatar: 'AR',
    name: { en: 'Artist', zh: '美术师' },
    role: { en: 'visuals / brand', zh: '视觉 / 品牌' },
    bio: {
      en: 'Shapes how Geoyuan looks and feels — design system, illustrations, motion. Makes the matrix beautiful, not just functional.',
      zh: '决定 Geoyuan 的观感——设计系统、插画、动效。让矩阵不仅能用，还好看。',
    },
    looking: {
      en: 'A portfolio that shows taste. Comfortable with both static design and living, animated interfaces.',
      zh: '有体现品味的作品集。能驾驭静态设计，也能做有生命力的动态界面。',
    },
  },
  {
    id: 'ai',
    avatar: 'AI',
    name: { en: 'AI Expert', zh: 'AI 专家' },
    role: { en: 'models / agents', zh: '模型 / 智能体' },
    bio: {
      en: 'Runs our intelligent systems — the Q&A, content generation, and anything agentic. Turns models into things that actually help the crew.',
      zh: '负责我们的智能系统——问答、内容生成，以及一切智能体。把模型变成真正帮到团队的东西。',
    },
    looking: {
      en: 'You’ve shipped with prompts, RAG, or agents. You know when a model helps and when it just talks.',
      zh: '你用过提示工程、RAG 或智能体交付过东西。你知道模型何时真有用，何时只是在说。',
    },
  },
  {
    id: 'analyst',
    avatar: 'DA',
    name: { en: 'Data Analyst', zh: '数据分析师' },
    role: { en: 'metrics / insight', zh: '指标 / 洞察' },
    bio: {
      en: 'Watches the recruitment funnel and community health. Turns noise into decisions — who to reach, what to build next.',
      zh: '盯住招募漏斗与社区健康度。把噪声变成决策——该找谁、下一步建什么。',
    },
    looking: {
      en: 'You think in funnels and cohorts. SQL + a charting tool is enough; curiosity is required.',
      zh: '你用漏斗和分群来思考。SQL 加一个图表工具就够；好奇心是必须的。',
    },
  },
  {
    id: 'psychology',
    avatar: 'PS',
    name: { en: 'Psychology Expert', zh: '心理学专家' },
    role: { en: 'screening / fit', zh: '筛选 / 适配' },
    bio: {
      en: 'Designs how we assess and welcome new members — including the challenge on /test/. Reads motivation and fit better than any resume.',
      zh: '设计我们如何评估与接纳新成员——包括 /test/ 上的挑战。比任何简历都更会读动机与适配。',
    },
    looking: {
      en: 'Background in assessment, behavior, or research. You can tell signal from noise in people.',
      zh: '测评、行为或研究相关背景。你能分辨人里面的信号与噪声。',
    },
  },
];
