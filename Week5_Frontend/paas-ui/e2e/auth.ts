import { Page, expect } from '@playwright/test';

const baseURL = process.env.BASE_URL || 'https://kevin-sinn.runs.onstackit.cloud/';
const username = process.env.E2E_USERNAME || 'kevin';
const password = process.env.E2E_PASSWORD || 'KevinsPassword';

export { baseURL };

/**
 * Logs in to the PaaS app. Call this at the start of tests that require authentication.
 * Waits for redirect to the dashboard before resolving.
 */
export async function login(page: Page): Promise<void> {
  const loginURL = baseURL.replace(/\/?$/, '') + '/login';

  await page.goto(loginURL);
  await page.getByRole('textbox', { name: 'Username' }).fill(username);
  await page.getByRole('textbox', { name: 'Password' }).fill(password);
  await page.getByRole('button', { name: 'Login' }).click();

  await expect(page).toHaveURL(baseURL);
  await expect(page.getByText('PaaS Dashboard')).toBeVisible();
}
