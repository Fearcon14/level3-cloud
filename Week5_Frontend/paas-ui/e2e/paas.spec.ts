import { test, expect } from '@playwright/test';
import { login, baseURL } from './auth';

const instanceName = 'playwright-instance';
const cacheKey = 'playwright-test';
const cacheValue = 'Success';

// Time to wait for instance to become fully operational (status "running").
const INSTANCE_READY_TIMEOUT_MS = 100_000;
const POLL_INTERVAL_MS = 5_000;

// Reload the page until the instance card shows "running" or timeout (page doesn't auto-refresh)
async function waitForInstanceRunning(page, name, timeoutMs) {
  const deadline = Date.now() + timeoutMs;
  const waitAfterReloadMs = 15_000; // give the page time to fetch and render after each reload
  while (Date.now() < deadline) {
    await page.reload();
    await page.waitForLoadState('load');
    const card = page.locator('.card').filter({ has: page.getByText(name, { exact: true }) });
    try {
      await expect(card).toBeVisible({ timeout: waitAfterReloadMs });
      await expect(card.getByText(/running/i)).toBeVisible({ timeout: 5_000 });
      return;
    } catch {
      // card not there yet or still not running, wait and reload again
    }
    await new Promise((r) => setTimeout(r, POLL_INTERVAL_MS));
  }
  await page.reload();
  const card = page.locator('.card').filter({ has: page.getByText(name, { exact: true }) });
  await expect(card.getByText(/running/i)).toBeVisible();
}

test.describe.serial('paas', () => {
  test('has title', async ({ page }) => {
    await page.goto(baseURL);
    await expect(page).toHaveTitle(/Kevin's PaaS/);
  });

  test('login', async ({ page }) => {
    await login(page);
    await expect(page.getByRole('heading', { name: 'Redis Instances' })).toBeVisible();
    await expect(page.getByText('Hello, kevin')).toBeVisible();
  });

  test('create instance', async ({ page }) => {
  await login(page);

  await page.getByRole('button', { name: '+ Create Instance' }).click();
  await page.getByRole('textbox', { name: 'Lowercase alphanumeric' }).click();
  await page.getByRole('textbox', { name: 'Lowercase alphanumeric' }).fill('playwright-instance');
  await page.getByRole('spinbutton').first().click();
  await page.getByRole('spinbutton').first().fill('1');
  await page.getByRole('spinbutton').nth(1).click();
  await page.getByRole('spinbutton').nth(1).fill('1');
  await page.getByRole('button', { name: 'Create', exact: true }).click();

  await expect(page.getByText(instanceName)).toBeVisible();
  });

  // test('use cache', async ({ page }) => {
  //   await login(page);

  //   const card = page.locator('.card').filter({ has: page.getByText(instanceName, { exact: true }) });
  //   await expect(card).toBeVisible();
  //   await waitForInstanceRunning(page, instanceName, INSTANCE_READY_TIMEOUT_MS);

  //   await card.getByRole('button', { name: 'Manage' }).click();
  //   await page.getByRole('textbox', { name: 'Key', exact: true }).click();
  //   await page.getByRole('textbox', { name: 'Key', exact: true }).fill('playwright-test');
  //   await page.getByRole('textbox', { name: 'Value' }).click();
  //   await page.getByRole('textbox', { name: 'Value' }).fill(cacheValue);
  //   await page.getByPlaceholder('TTL (seconds, 0 = no expiry)').click();
  //   await page.getByPlaceholder('TTL (seconds, 0 = no expiry)').fill('30');
  //   await page.getByRole('button', { name: 'POST' }).click();
  //   await page.getByRole('textbox', { name: 'Cache key' }).click();
  //   await page.getByRole('textbox', { name: 'Cache key' }).fill(cacheKey);
  //   await page.getByRole('button', { name: 'GET' }).click();

  //   await expect(page.getByText(`Value:${cacheValue}`)).toBeVisible();
  // });

  // test('modify instance', async ({ page }) => {
  //   await login(page);

  //   const card = page.locator('.card').filter({ has: page.getByText(instanceName, { exact: true }) });
  //   await expect(card).toBeVisible();
  //   // await waitForInstanceRunning(page, instanceName, INSTANCE_READY_TIMEOUT_MS);

  //   await card.getByRole('button', { name: 'Manage' }).click();
  //   await page.getByTitle('Modify').click();
  //   await page.getByRole('spinbutton').nth(1).click();
  //   await page.getByRole('spinbutton').nth(1).fill('3');
  //   await page.getByRole('spinbutton').nth(2).click();
  //   await page.getByRole('spinbutton').nth(2).fill('3');
  //   await page.getByRole('button', { name: 'Save changes' }).click();

  //   await expect(page.getByText('Redis Replicas3')).toBeVisible();
  //   await expect(page.getByText('Sentinel Replicas3')).toBeVisible();
  // });

  test('delete instance', async ({ page }) => {
    await login(page);

    const card = page.locator('.card').filter({ has: page.getByText(instanceName, { exact: true }) });
    await expect(card).toBeVisible();

    page.once('dialog', (dialog) => dialog.accept());
    await card.getByRole('button', { name: /Delete/ }).click();

    await expect(card).toBeHidden();
  });
}); // end describe.serial
