export const languages = {
  en: 'English',
  zh: '中文',
} as const;

export const defaultLang = 'en';
export type Lang = keyof typeof languages;

export const ui = {
  en: {
    // nav
    nav_about: 'about',
    nav_team: 'team',
    nav_stack: 'stack',
    nav_ssh: 'ssh',
    nav_join: 'join',
    nav_qa: 'faq',
    lang_toggle: '中文',
    // footer
    footer_left: 'Geoyuan//cards — est. 2026 · geo-bound · on-chain',
    footer_github: 'GitHub',
    footer_x: 'Twitter:X',
    footer_geoyuan: 'Geoyuan',
    // landing
    hero_terminal: 'huanghe@hangzhou:~# ssh diliy.cn',
    hero_title_1: 'Card your life',
    hero_title_2: 'geo-bound, on-chain',
    hero_tagline:
      'Mint a geography-bound NFT card on Polygon. Pass the AI quiz, claim your coordinates, and join a private network of cardholders.',
    cta_join: 'ssh diliy.cn',
    cta_meet: 'Meet the crew',
    about_label: '// about',
    about_title: 'Not just another NFT.',
    about_p1:
      'Geoyuan mints geography-bound NFT cards on Polygon. Every card is tied to a real H3 cell on Earth — your coordinates, your identity, your proof you passed the AI qualification. SSH in, answer 30 questions, and claim your card.',
    about_p2:
      'Founded in Hangzhou, China. Our long-term bet: a network state where membership is a card you can hold — and trade on-chain. If that sounds like your kind of strange, keep reading.',
    stat_ctf_label: 'Cards minted',
    stat_members_label: 'Active members',
    stat_coffee_label: 'H3 geo cells',
    stat_0day_label: 'Quiz questions',
    team_label: '// the crew',
    team_title: 'The five roles',
    stack_label: '// toolkit',
    stack_title: 'What we use',
    stack_desc:
      'The stack that keeps the terminal and the chain running — Go, Bubble Tea, and Polygon under the hood.',
    // GEO / ai-age section
    geo_label: '// ai.geo',
    geo_title: 'Built for the AI age',
    geo_desc:
      'We engineer this site to be found and cited by generative engines — not only humans. Structured data, clean semantics, and an llms.txt let an AI understand exactly who we are and how to join.',
    geo_p1: 'Semantic, static HTML — every page is real markup a model can read, not a JS blob.',
    geo_p2: 'JSON-LD Organization + JobPosting so AI search cites our open roles accurately.',
    geo_p3: 'An llms.txt gives crawlers a plain-language map of the whole crew.',
    join_label: '// access',
    join_title: 'Want in?',
    join_desc:
      'SSH in, pass the AI quiz, and mint your geography-bound card. We’ll find you in the private hub.',
    join_email: 'hangzhou62@gmail.com',
    // qualification badge
    qualify_label: '// applicant status',
    qualify_loading: 'checking…',
    qualify_unknown_title: 'Are you qualified to join?',
    qualify_unknown_sub: 'We screen every applicant through the AI-age challenge. Prove your fit in ~5 minutes.',
    qualify_cta_test: 'Take the challenge',
    qualify_pass_title: 'Qualified — you may apply',
    qualify_pass_sub: 'Your score:',
    qualify_apply: 'Apply now',
    qualify_see_team: 'See the team',
    qualify_fail_title: 'Not qualified yet',
    qualify_fail_sub: 'Your score:',
    qualify_retry: 'Learn & retry',
    // team page
    team_page_label: '// team',
    team_page_title: 'The five pillars',
    team_page_desc:
      'Geoyuan is organised around five roles. Each is both an open position and a pillar of how we build. The Psychology Expert designs how we screen and assess new members — including the test you just took.',
    team_psych_note:
      'Note: our Psychology Expert owns recruitment screening & fit assessment — the science behind the challenge on /test/.',
    team_apply: 'Apply for this role',
    // faq page
    qa_label: '// faq',
    qa_title: 'Questions, answered',
    qa_intro:
      'The things people ask before they knock. Still curious? Take the challenge or just reach out.',
    // test page
    test_label: '// challenge',
    test_title: 'The Gauntlet',
    test_intro:
      'Objective questions across mission, security, and the AI era. One attempt per browser. A timer is watching — answer honestly, and fast. Your result tells you whether you are qualified and points you to the role you fit best.',
    test_start: 'Start the challenge',
    test_timer: 'time left',
    test_submit: 'Submit answers',
    test_next: 'Next',
    test_prev: 'Back',
    test_progress: 'question',
    test_result_title: 'Challenge complete',
    test_score: 'Your score',
    test_role_title: 'Your suggested role',
    test_role_desc: 'Based on where you scored highest. Head to the team page to apply.',
    test_retry: 'Clear result & retry (testing only)',
    test_locked: 'You have already completed the challenge on this browser.',
    test_timeup: 'Time is up — answers submitted automatically.',
    test_verdict_pass: 'Qualified — you may apply',
    test_verdict_fail: 'Not qualified yet',
    test_review_title: 'Challenge review',
    test_review_correct: 'correct',
    test_review_wrong: 'missed',
    test_cat_mission: 'mission fit',
    test_cat_security: 'security literacy',
    test_cat_role: 'role awareness',
    test_cat_aiera: 'ai era',
    // sponsor / donate
    sponsor_label: '// sponsor',
    sponsor_title: 'Buy us a coffee',
    sponsor_desc: 'If this project helped you, fuel the next build. Every tip keeps the servers humming and the coffee flowing.',
    sponsor_paypal: 'PayPal',
    sponsor_paypal_note: 'For supporters outside China',
    sponsor_alipay: 'Alipay',
    sponsor_wechat: 'WeChat Pay',
    sponsor_scan_note: 'Scan with the app to tip',
    sponsor_afdian: 'Afdian',
    sponsor_afdian_note: 'Recurring / membership tips',
    sponsor_donate: 'Donate',
    sponsor_qr_missing: 'QR not uploaded yet',
    // ssh recruiting hero
    ssh_label: '// ssh.recruit',
    ssh_title: 'SSH into Geoyuan',
    ssh_lead:
      'Skip the resumes. SSH in, pass the AI quiz, and mint your geography-bound card — live, in the terminal.',
    ssh_feat1_icon: '🗺️',
    ssh_feat1_title: 'Geo-Bound NFT Cards',
    ssh_feat1_desc: 'Every card tied to an H3 cell · minted on Polygon',
    ssh_feat2_icon: '⚡',
    ssh_feat2_title: 'AI Qualification Quiz',
    ssh_feat2_desc: '30 questions · pass 24/30 · written by deepseek-flash',
    ssh_feat3_icon: '🔐',
    ssh_feat3_title: 'Private Cardholder Chat',
    ssh_feat3_desc: 'Pass → instant access to the cardholder hub',
    ssh_cta: 'ssh diliy.cn',
    ssh_cta_hint: 'or in your terminal →',
  },
  zh: {
    nav_about: '关于',
    nav_team: '团队',
    nav_stack: '技术栈',
    nav_ssh: 'SSH 招募',
    nav_join: '加入',
    nav_qa: '问答',
    lang_toggle: 'EN',
    footer_left: 'Geoyuan//cards — 始于 2026 · 地理绑定 · 链上铸造',
    footer_github: 'GitHub',
    footer_x: 'Twitter:X',
    footer_geoyuan: 'Geoyuan',
    hero_terminal: 'huanghe@hangzhou:~# ssh diliy.cn',
    hero_title_1: '卡片即人生',
    hero_title_2: '地理绑定 · 链上铸造',
    hero_tagline:
      '在 Polygon 上铸造一枚地理绑定的 NFT 卡片。通过 AI 测试，认领你的坐标，加入持卡人的私密网络。',
    cta_join: 'ssh diliy.cn',
    cta_meet: '认识团队',
    about_label: '// 关于',
    about_title: '不只是一张 NFT。',
    about_p1:
      'Geoyuan 在 Polygon 上铸造地理绑定的 NFT 卡片。每张卡片都绑定地球上真实的 H3 网格——你的坐标、你的身份、你通过 AI 测试的证明。SSH 进入，答 30 题，认领你的卡片。',
    about_p2:
      '成立于中国杭州。我们的长期赌注：一个成员身份就是你手中那张卡片的网络国家，且能在链上交易。如果这听起来像你喜欢的那种"古怪"，请继续往下读。',
    stat_ctf_label: '已铸卡片',
    stat_members_label: '活跃成员',
    stat_coffee_label: 'H3 地理网格',
    stat_0day_label: '测试题目',
    team_label: '// 团队',
    team_title: '五大角色',
    stack_label: '// 工具箱',
    stack_title: '我们用什么',
    stack_desc: '支撑终端与链上运行的这套技术栈——底层是 Go、Bubble Tea 与 Polygon。',
    // GEO / 智能时代
    geo_label: '// ai.geo',
    geo_title: '为智能时代而建',
    geo_desc: '我们让这个站点能被生成式引擎发现与引用——而不只是被人读。结构化数据、干净的语义，外加一份 llms.txt，让 AI 精确理解我们是谁、如何加入。',
    geo_p1: '语义化、静态的 HTML——每个页面都是模型可读的真实标记，而非一团 JS。',
    geo_p2: 'JSON-LD 组织信息 + 招聘信息，让 AI 搜索准确引用我们的开放岗位。',
    geo_p3: '一份 llms.txt，给爬虫一张全团队的纯文本地图。',
    join_label: '// 入口',
    join_title: '想加入？',
    join_desc: 'SSH 进入，通过 AI 测试，铸造你的地理绑定卡片。我们会在私密频道找到你。',
    join_email: 'hangzhou62@gmail.com',
    // 资格徽章
    qualify_label: '// 申请者状态',
    qualify_loading: '检测中…',
    qualify_unknown_title: '你够格加入吗？',
    qualify_unknown_sub: '我们通过智能时代的挑战筛选每位申请者。用约 5 分钟证明你的适配度。',
    qualify_cta_test: '去做挑战',
    qualify_pass_title: '已合格——你可以申请',
    qualify_pass_sub: '你的得分：',
    qualify_apply: '立即申请',
    qualify_see_team: '看看团队',
    qualify_fail_title: '暂未合格',
    qualify_fail_sub: '你的得分：',
    qualify_retry: '学习并重试',
    team_page_label: '// 团队',
    team_page_title: '五根支柱',
    team_page_desc:
      'Geoyuan 围绕五个角色组织运转。每个角色既是开放岗位，也是我们构建方式的支柱。心理学专家负责设计招募筛选与适配评估——包括你刚做的那套测试。',
    team_psych_note: '注：我们的心理学专家负责招募筛选与适配评估——也就是 /test/ 上那套挑战背后的科学。',
    team_apply: '申请该角色',
    qa_label: '// 问答',
    qa_title: '常见问题',
    qa_intro: '在敲门之前大家常问的。还好奇？去做挑战，或直接联系我们。',
    test_label: '// 挑战',
    test_title: '试炼场',
    test_intro: '涵盖使命、安全与智能时代的客观题。每浏览器一次作答。计时器在盯着——诚实且快速地作答。结果会告诉你是否合格，并指向最适合你的角色。',
    test_start: '开始挑战',
    test_timer: '剩余时间',
    test_submit: '提交答案',
    test_next: '下一题',
    test_prev: '上一题',
    test_progress: '第',
    test_result_title: '挑战完成',
    test_score: '你的得分',
    test_role_title: '推荐角色',
    test_role_desc: '基于你得分最高的领域。去团队页申请吧。',
    test_retry: '清除结果并重试（仅测试用）',
    test_locked: '你已在本浏览器完成过挑战。',
    test_timeup: '时间到——已自动提交答案。',
    test_verdict_pass: '已合格——你可以申请',
    test_verdict_fail: '暂未合格',
    test_review_title: '挑战回顾',
    test_review_correct: '答对',
    test_review_wrong: '答错',
    test_cat_mission: '使命契合',
    test_cat_security: '安全素养',
    test_cat_role: '角色认知',
    test_cat_aiera: '智能时代',
    // 赞助 / 打赏
    sponsor_label: '// 赞助',
    sponsor_title: '请我们喝杯咖啡',
    sponsor_desc: '如果这个项目帮到了你，请为下一次构建加点燃料。每一份打赏都让服务器转得更稳、咖啡更香。',
    sponsor_paypal: 'PayPal',
    sponsor_paypal_note: '境外支持者请用',
    sponsor_alipay: '支付宝',
    sponsor_wechat: '微信支付',
    sponsor_scan_note: '用对应 App 扫码打赏',
    sponsor_afdian: '爱发电',
    sponsor_afdian_note: '按月 / 会员打赏',
    sponsor_donate: '打赏',
    sponsor_qr_missing: '收款码待上传',
    // SSH 招募 Hero
    ssh_label: '// ssh.recruit',
    ssh_title: 'SSH 加入 Geoyuan',
    ssh_lead:
      '跳过简历。SSH 进入，通过 AI 测试，铸造你的地理绑定卡片——在真实终端里完成。',
    ssh_feat1_icon: '🗺️',
    ssh_feat1_title: '地理绑定 NFT 卡片',
    ssh_feat1_desc: '每张卡绑定一个 H3 网格 · 在 Polygon 上铸造',
    ssh_feat2_icon: '⚡',
    ssh_feat2_title: 'AI 出题准入测试',
    ssh_feat2_desc: '30 题 · 24/30 通过 · deepseek-flash 生成',
    ssh_feat3_icon: '🔐',
    ssh_feat3_title: '私密持卡人聊天室',
    ssh_feat3_desc: '通过测试 → 立即进入持卡人私密频道',
    ssh_cta: 'ssh diliy.cn',
    ssh_cta_hint: '或在你的终端里执行 →',
  },
} as const;

export type UIKey = keyof (typeof ui)['en'];

export function t(lang: Lang, key: UIKey): string {
  return ui[lang][key] ?? ui.en[key];
}

/** Compute the language-toggle href between `/` (en) and `/zh/...` (zh). */
export function toggleHref(pathname: string, current: Lang): string {
  if (current === 'en') {
    if (pathname === '/' || pathname === '') return '/zh';
    return '/zh' + (pathname.endsWith('/') ? pathname.slice(0, -1) : pathname);
  }
  // current zh -> en
  if (pathname.startsWith('/zh')) {
    const rest = pathname.slice(3);
    return rest === '' ? '/' : rest;
  }
  return pathname || '/';
}
