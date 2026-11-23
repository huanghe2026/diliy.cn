import { defineConfig } from 'astro/config';

export default defineConfig({
  outDir: './dist',
  site: 'https://diliy.cn',
  // Use noop image service to avoid sharp native binary crashes
  // (sandbox CPU doesn't support required instructions).
  // Images referenced via <img src="/..."> in public/ still work fine.
  image: {
    service: { entrypoint: 'astro/assets/services/noop' },
  },
});
