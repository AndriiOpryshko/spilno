import { expect, test } from '@playwright/test';

import { fileAndConfirm, newEmail, newFiling, operatorEmail, sessionFor, today } from './support';

test('US-01: a голова files a заявка and finds it pending in their list (AC-01)', async ({ browser }) => {
  const f = newFiling();
  const page = await sessionFor(browser, newEmail('head'));

  await fileAndConfirm(page, f);

  await page.goto('/applications');
  await expect(page.getByRole('heading', { name: 'Мої заявки' })).toBeVisible();
  await expect(page.getByRole('link', { name: f.name, exact: true })).toBeVisible();
  await expect(page.getByText(`очікує розгляду з ${today()}`)).toBeVisible();
  await expect(page.getByText(`Питання щодо заявки: ${operatorEmail}`)).toBeVisible();
});
