/// <reference types="vitest/config" />
import path from 'node:path';
import tailwindcss from '@tailwindcss/vite';
import { tanstackRouter } from '@tanstack/router-plugin/vite';
import react from '@vitejs/plugin-react';
import { storybookTest } from '@storybook/addon-vitest/vitest-plugin';
import { playwright } from '@vitest/browser-playwright';
import { defineConfig } from 'vite';

const apiProxy = {
  '/api': {
    target: process.env.CORE_API_URL ?? 'http://localhost:8080',
    rewrite: (p: string) => p.replace(/^\/api/, ''),
  },
};

export default defineConfig({
  plugins: [tanstackRouter({ target: 'react', autoCodeSplitting: true }), react(), tailwindcss()],
  resolve: {
    alias: { '@': path.resolve(import.meta.dirname, './src') },
  },
  server: { port: 5180, strictPort: true, proxy: apiProxy },
  preview: { port: 4180, strictPort: true, proxy: apiProxy },
  test: {
    projects: [
      {
        extends: true,
        test: {
          name: 'unit',
          environment: 'jsdom',
          setupFiles: ['./src/test/setup.ts'],
          include: ['src/**/*.test.{ts,tsx}'],
        },
      },
      {
        // Every story renders in Chromium and runs its play function and a11y checks.
        extends: true,
        plugins: [storybookTest({ configDir: path.join(import.meta.dirname, '.storybook') })],
        test: {
          name: 'storybook',
          // Visual regression: `toMatchScreenshot` in the stories. Baselines are committed next to the
          // features, one per platform and architecture (font rasterisation differs between them).
          // Update deliberately: pnpm --filter @spilno/web test:stories -u
          browser: {
            enabled: true,
            headless: true,
            provider: playwright({}),
            instances: [{ browser: 'chromium' }],
            // Pinned so a laptop and the CI image render the same pixels.
            viewport: { width: 1280, height: 800 },
            expect: {
              toMatchScreenshot: {
                resolveScreenshotPath: ({ arg, browserName, ext }) =>
                  path.join(
                    import.meta.dirname,
                    'src/features/identity/__screenshots__',
                    `${arg}-${browserName}-${process.platform}-${process.arch}${ext}`,
                  ),
                comparatorName: 'pixelmatch',
                comparatorOptions: { threshold: 0.1, allowedMismatchedPixelRatio: 0.001 },
              },
            },
          },
        },
      },
    ],
  },
});
