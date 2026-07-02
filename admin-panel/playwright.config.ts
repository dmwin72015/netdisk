import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  timeout: 30000,
  use: {
    baseURL: 'http://localhost:9173',
    screenshot: 'on',
    viewport: { width: 1920, height: 1080 },
    ignoreHTTPSErrors: true,
  },
});
