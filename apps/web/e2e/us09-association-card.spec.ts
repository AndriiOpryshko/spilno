import { expect, test } from '@playwright/test';

import { connectAssociation, newEmail, newFiling, openCard, today } from './support';
import { approveTransfer, pendingTransfer } from './transfer-support';

const newHead = 'Новий Голова Правлінович';

test('US-09: a член правління reads the confirmation chain and downloads the extract (AC-29, AC-20)', async ({ browser }) => {
  const f = newFiling();
  const { applicant: previous } = await connectAssociation(browser, f, newEmail('oldhead'));

  // Connected: the card, opened from the sidebar, shows the голова, the signer, the extract date and the decision date.
  await openCard(previous);
  await expect(previous.getByRole('heading', { name: f.name })).toBeVisible();
  await expect(previous.getByText(f.fullName).first()).toBeVisible();
  await expect(previous.getByText(`Підключено · рішення від ${today()}`)).toBeVisible();
  await expect(previous.getByText(`підписант: ${f.fullName}`)).toBeVisible();
  await expect(previous.getByText(/^витяг від \d{2}\.\d{2}\.\d{4}/)).toBeVisible();
  await expect(previous.getByText('підключено вручну, без витягу')).toHaveCount(0);

  const download = previous.waitForEvent('download');
  await previous.getByRole('link', { name: 'Завантажити витяг' }).click();
  expect((await download).suggestedFilename()).toBeTruthy();

  // After a transfer the chain has the запит, and the roster notes of AC-20 name both dates.
  const { page: incoming, id } = await pendingTransfer(browser, f, newEmail('newhead'), newHead);
  await approveTransfer(browser, f, id, newHead);
  await openCard(incoming);
  await expect(incoming.getByText(`Передачу схвалено · рішення від ${today()}`)).toBeVisible();
  await expect(incoming.getByText(`Підключено · рішення від ${today()}`)).toBeVisible();
  await expect(incoming.getByText(`підписант: ${newHead}`)).toBeVisible();
  await expect(incoming.getByText(`призначено: ${newHead} за запитом на передачу від ${today()}, схвалено оператором ${today()}`)).toBeVisible();
  await expect(incoming.getByText(`повноваження завершено: ${f.fullName} за запитом на передачу від ${today()}, схвалено оператором ${today()}`)).toBeVisible();
  await expect(incoming.getByRole('link', { name: 'Завантажити витяг' })).toHaveCount(2);
});
