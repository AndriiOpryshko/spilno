import { expect, test } from '@playwright/test';

import { connectAssociation, newEmail, newFiling, sessionFor } from './support';
import { fileTransfer, fileTransferAndConfirm } from './transfer-support';

test('US-04: the new голова files a запит на передачу and follows it (AC-17)', async ({ browser }) => {
  const f = newFiling();
  await connectAssociation(browser, f, newEmail('head'));

  const incoming = await sessionFor(browser, newEmail('newhead'));
  await fileTransferAndConfirm(incoming, f, 'Новий Голова Правлінович');

  // Nothing of the current roster is revealed; the запит is listed as pending.
  await expect(incoming.getByText(f.fullName)).toHaveCount(0);
  await incoming.goto('/applications');
  await expect(incoming.getByRole('link', { name: f.name, exact: true })).toHaveCount(1);
  await expect(incoming.getByText(/^очікує розгляду з/)).toBeVisible();
});

test('US-04: a запит for a code that is not connected is refused and the заявка form offered (AC-18)', async ({ browser }) => {
  const page = await sessionFor(browser, newEmail('head'));
  await fileTransfer(page, newFiling(), 'Новий Голова Правлінович');
  await expect(page.getByText('Це ОСББ ще не підключене — подайте заявку на підключення')).toBeVisible();
  await page.getByRole('link', { name: 'Подати заявку на підключення' }).click();
  await expect(page.getByRole('heading', { name: 'Заявка на підключення' })).toBeVisible();
  await page.goto('/applications');
  await expect(page.getByText('У вас ще немає заявок')).toBeVisible();
});
