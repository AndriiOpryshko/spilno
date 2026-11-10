import { expect, test } from '@playwright/test';

test('status page shows live service health from the core API', async ({ page }) => {
  await page.goto('/status');
  await expect(page.getByText('Service health')).toBeVisible();
  await expect(page.getByText('ok', { exact: true })).toBeVisible();
  await expect(page.getByText(/^up \(Postgres /)).toBeVisible();
});
