export interface FaqItem {
  q: { en: string; zh: string };
  a: { en: string; zh: string };
}

// AI-generated FAQ for the Geoyuan recruitment community.
export const faq: FaqItem[] = [
  {
    q: {
      en: 'What exactly is Geoyuan?',
      zh: 'Geoyuan 到底是什么？',
    },
    a: {
      en: 'A small, globally distributed crew building toward a network state. We do CTFs, build our own tooling, and run sci-fi experiments in the open. Think of it as a club for people who ship and stay curious.',
      zh: '一支极小、分布全球的团队，正向网络国家进发。我们打 CTF、自研工具，并公开做科幻式实验。可以理解为"为那些能交付、保持好奇的人"准备的俱乐部。',
    },
  },
  {
    q: {
      en: 'How do I actually join?',
      zh: '我到底怎么加入？',
    },
    a: {
      en: 'Read the site, take the challenge at /test/, then reach out by email or drop a Geoyuan card with the text “KAI MEN”. If you’re a fit, we’ll find you.',
      zh: '读完本站，去 /test/ 做挑战，然后用邮件联系我们，或扔一张写着"开门"的 Geoyuan 卡片。合拍的话，我们会找到你。',
    },
  },
  {
    q: {
      en: 'What is the challenge on /test/?',
      zh: '/test/ 上的挑战是什么？',
    },
    a: {
      en: '12 objective questions with standard answers. One attempt per browser, and a timer is watching. It screens fit and points you to the role you match best — it’s not a wall, just a filter.',
      zh: '12 道有标准答案的客观题。每浏览器一次作答，计时器在盯着。它筛选适配度，并指出你最匹配的角色——它不是墙，只是过滤器。',
    },
  },
  {
    q: {
      en: 'Do I need to be a hacker?',
      zh: '我必须是个黑客吗？',
    },
    a: {
      en: 'No. We have five roles: programmer, artist, AI expert, data analyst, and psychology expert. Pick where you fit — we need builders of every kind.',
      zh: '不需要。我们有五个角色：程序员、美术师、AI专家、数据分析师、心理学专家。选你适合的那块——我们需要各种构建者。',
    },
  },
  {
    q: {
      en: 'Is this remote and global?',
      zh: '这是远程、全球化的吗？',
    },
    a: {
      en: 'Yes. We are borderless. Time zones don’t matter to us — output does. You can be anywhere with a connection.',
      zh: '是的。我们无国界。时区对我们无关紧要——交付才重要。只要有网络，你在哪都行。',
    },
  },
  {
    q: {
      en: 'What language should I use?',
      zh: '我该用什么语言？',
    },
    a: {
      en: 'The site is English + 中文, and community comms are a mix. English is the lingua franca; Chinese works too. Write in whatever you think in.',
      zh: '本站是英文 + 中文，社区交流也是混着来。英文是通用语，中文也行。用你思考时所用的语言写就好。',
    },
  },
  {
    q: {
      en: 'How much time is expected?',
      zh: '需要投入多少时间？',
    },
    a: {
      en: 'We care about shipped work, not hours logged. Contribute where you can, when you can. Consistency beats intensity.',
      zh: '我们在乎交付的成果，不在乎打卡的时长。能贡献时就贡献。持续比一时猛干更重要。',
    },
  },
  {
    q: {
      en: 'Is there a hierarchy?',
      zh: '这里有等级制度吗？',
    },
    a: {
      en: 'Flat-ish. Five pillars (roles) coordinate the work, and the Psychology Expert owns how we screen and assess new members.',
      zh: '比较扁平。五根支柱（角色）协同推进工作，心理学专家负责我们如何筛选与评估新成员。',
    },
  },
  {
    q: {
      en: 'Do you keep logs or track me?',
      zh: '你们会记录日志或追踪我吗？',
    },
    a: {
      en: 'We minimise data on purpose — less to leak, less to lose. The challenge runs client-side and stores only your result locally in your browser.',
      zh: '我们刻意在最小范围内收集数据——可泄露的少，损失的也少。挑战在客户端运行，只在你浏览器本地保存你的结果。',
    },
  },
  {
    q: {
      en: 'How do I contact you?',
      zh: '怎么联系你们？',
    },
    a: {
      en: 'Email hangzhou62@gmail.com, or find us on GitHub / Twitter:X via the footer links. A Geoyuan card with “KAI MEN” works too.',
      zh: '邮件 hangzhou62@gmail.com，或通过页脚链接在 GitHub / Twitter:X 找到我们。一张写着"开门"的 Geoyuan 卡片也行。',
    },
  },
];
