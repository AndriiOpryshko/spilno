import { expect, test } from '@playwright/test';

// verify-ui on the first screen found one console error: /favicon.ico answered 404.
test('home page loads without console errors and serves its favicon', async ({ page }) => {
  const errors: string[] = [];
  page.on('console', (m) => {
    if (m.type() === 'error') errors.push(m.text());
  });

  await page.goto('/');
  await expect(page.getByText('Service health')).toBeVisible();

  const href = await page.locator('link[rel="icon"]').getAttribute('href');
  expect(href).toBeTruthy();
  const icon = await page.request.get(href!);
  expect(icon.status()).toBe(200);
  expect(icon.headers()['content-type']).toContain('image/svg+xml');

  expect(errors).toEqual([]);
});
