import { test } from '@playwright/test';
import { fileURLToPath } from 'url';
import path from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const SCREENSHOT_DIR = path.resolve(__dirname, '../screenshots');

async function loginAndSaveToken(page: any): Promise<string> {
  const response = await page.request.post('http://localhost:8080/api/v1/auth/login', {
    data: { email: 'w446709@gmail.com', password: 'admin123' },
  });
  const body = await response.json();
  return body.access || body.token || body.data?.access || body.data?.token || body.data?.tokens?.accessToken || '';
}

async function setupAuth(page: any, token: string) {
  await page.goto('/admin');
  await page.evaluate((t) => {
    localStorage.setItem('nd.access', t);
    localStorage.setItem('nd.user', JSON.stringify({ role: 'admin', username: 'admin' }));
  }, token);
}

test('capture login page', async ({ page }) => {
  await page.goto('/login');
  await page.waitForTimeout(2000);
  await page.screenshot({ path: path.join(SCREENSHOT_DIR, 'login.png'), fullPage: true });
});

test('capture users page', async ({ page }) => {
  const token = await loginAndSaveToken(page);
  await setupAuth(page, token);
  await page.goto('/admin/users');
  await page.waitForTimeout(3000);
  await page.screenshot({ path: path.join(SCREENSHOT_DIR, 'users.png'), fullPage: true });
});

test('capture files page', async ({ page }) => {
  const token = await loginAndSaveToken(page);
  await setupAuth(page, token);
  await page.goto('/admin/files/user-files');
  await page.waitForTimeout(3000);
  await page.screenshot({ path: path.join(SCREENSHOT_DIR, 'files.png'), fullPage: true });
});

test('capture physical files page', async ({ page }) => {
  const token = await loginAndSaveToken(page);
  await setupAuth(page, token);
  await page.goto('/admin/files/physical');
  await page.waitForTimeout(3000);
  await page.screenshot({ path: path.join(SCREENSHOT_DIR, 'physical-files.png'), fullPage: true });
});

test('capture dashboard page', async ({ page }) => {
  const token = await loginAndSaveToken(page);
  await setupAuth(page, token);
  await page.goto('/admin');
  await page.waitForTimeout(3000);
  await page.screenshot({ path: path.join(SCREENSHOT_DIR, 'dashboard.png'), fullPage: true });
});

test('capture storage page', async ({ page }) => {
  const token = await loginAndSaveToken(page);
  await setupAuth(page, token);
  await page.goto('/admin/storage');
  await page.waitForTimeout(3000);
  await page.screenshot({ path: path.join(SCREENSHOT_DIR, 'storage.png'), fullPage: true });
});