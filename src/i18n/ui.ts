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
    nav_join: 'join',
    nav_blog: 'blog',
    nav_qa: 'faq',
    lang_toggle: '中文',
    // footer
    footer_left: 'Geoyuan//team — est. 2026 · no logs · no traces',
    footer_github: 'GitHub',
    footer_x: 'Twitter:X',
    footer_geoyuan: 'Geoyuan',
    // landing
    hero_terminal: 'huanghe@hangzhou:~# ./apply.sh',
    hero_title_1: 'We build Geoyuan',
    hero_title_2: 'to a Network state',
    hero_tagline:
      'A distributed, borderless crew building toward a network state. We are recruiting builders from all over the world — prove you belong.',
    cta_join: 'Join the diliy.cn',
    cta_meet: 'Meet the crew',
    about_label: '// about',
    about_title: 'Not just another group.',
    about_p1:
      'Geoyuan is a small, globally distributed team. We do not care about passports or time zones — we care about people who ship, think, and stay curious. This is a site in building for our crews and for those who want in.',
    about_p2:
      'Founded in Hangzhou, China. Our long-term bet: a sci-fi company and a network state, built in the open. If that sounds like your kind of strange, keep reading.',
    stat_ctf_label: 'CTF wins',
    stat_members_label: 'Active members',
    stat_coffee_label: 'Coffee consumed',
    stat_0day_label: '0-days found',
    team_label: '// the crew',
    team_title: 'Operators',
    stack_label: '// toolkit',
    stack_title: 'What we use',
    stack_desc:
      'The usual suspects. We don’t care about the shiny new framework — we care about what actually works at 3 AM during a CTF.',
    join_label: '// access',
    join_title: 'Want in?',
    join_desc:
      'Prove yourself in our test, then drop a Geoyuan card with the text "KAI MEN". We’ll find you.',
    join_email: 'hangzhou62@gmail.com',
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
      '12 objective questions. One attempt per browser. A timer is watching — answer honestly, and fast. Your result points you to the role you fit best.',
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
    test_cat_mission: 'mission fit',
    test_cat_security: 'security literacy',
    test_cat_role: 'role awareness',
  },
  zh: {
    nav_about: '关于',
    nav_team: '团队',
    nav_stack: '技术栈',
    nav_join: '加入',
    nav_blog: '博客',
    nav_qa: '问答',
    lang_toggle: 'EN',
    footer_left: 'Geoyuan//team — 始于 2026 · 无日志 · 无痕迹',
    footer_github: 'GitHub',
    footer_x: 'Twitter:X',
    footer_geoyuan: 'Geoyuan',
    hero_terminal: 'huanghe@hangzhou:~# ./apply.sh',
    hero_title_1: '我们正在构建 Geoyuan',
    hero_title_2: '迈向网络国家',
    hero_tagline:
      '一支去中心化、无国界的团队，正向网络国家进发。我们正在全球招募志同道合的构建者——证明你属于这里。',
    cta_join: '加入 diliy.cn',
    cta_meet: '认识团队',
    about_label: '// 关于',
    about_title: '不只是一群人。',
    about_p1:
      'Geoyuan 是一支极小、分布全球的团队。我们不在乎护照或时区——只在乎那些能交付、会思考、保持好奇的人。这个站点既是为我们自己，也为想加入的人而建。',
    about_p2:
      '成立于中国杭州。我们的长期赌注：一家科幻公司，一个网络国家，公开共建。如果这听起来像你喜欢的那种"古怪"，请继续往下读。',
    stat_ctf_label: 'CTF 胜场',
    stat_members_label: '活跃成员',
    stat_coffee_label: '咖啡消耗',
    stat_0day_label: '发现 0day',
    team_label: '// 团队',
    team_title: '成员',
    stack_label: '// 工具箱',
    stack_title: '我们用什么',
    stack_desc: '老面孔。我们不在乎花哨的新框架——只在乎 CTF 凌晨三点真正管用的是什么。',
    join_label: '// 入口',
    join_title: '想加入？',
    join_desc: '先在测试中证明自己，再扔一张写着"开门"的 Geoyuan 卡片。我们会找到你。',
    join_email: 'hangzhou62@gmail.com',
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
    test_intro: '12 道客观题。每浏览器一次作答。计时器在盯着——诚实且快速地作答。结果会指向最适合你的角色。',
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
    test_cat_mission: '使命契合',
    test_cat_security: '安全素养',
    test_cat_role: '角色认知',
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
