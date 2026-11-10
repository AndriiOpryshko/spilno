import { defineConfig, devices } from '@playwright/test';

// e2e runs against the real stack: the core service must be up on :8080
// (mprocs or CI starts it) with SPILNO_DEV_SIGN_IN=true and SPILNO_OPERATOR_EMAIL; Playwright starts the
// web dev server itself. E2E_WEB_PORT and CORE_API_URL move the web port and the API address.
const webPort = process.env.E2E_WEB_PORT ?? '5180';
const webUrl = `http://localhost:${webPort}`;

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? 'github' : 'list',
  use: {
    baseURL: webUrl,
    trace: 'on-first-retry',
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
  webServer: {
    command: `pnpm dev --port ${webPort}`,
    url: webUrl,
    // The оператор's contact address shown to an applicant (AC-05); CORE_API_URL is read by vite's proxy.
    env: { VITE_OPERATOR_CONTACT_EMAIL: process.env.E2E_OPERATOR_EMAIL ?? 'operator@example.test' },
    reuseExistingServer: !process.env.CI,
  },
});
