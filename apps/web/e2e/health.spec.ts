import { expect, test } from '@playwright/test';

test('home page shows live service health from the core API', async ({ page }) => {
  await page.goto('/');
  await expect(page.getByText('Service health')).toBeVisible();
  await expect(page.getByText('ok', { exact: true })).toBeVisible();
  await expect(page.getByText(/^up \(Postgres /)).toBeVisible();
});
