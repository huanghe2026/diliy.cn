import { defineConfig } from 'astro/config';

// Bilingual routing is handled manually: `/` = English, `/zh/...` = 中文.
// No i18n integration needed; each page renders from src/i18n/ui.ts + data modules.
export default defineConfig({
  outDir: './dist',
  site: 'https://diliy.cn',
});
