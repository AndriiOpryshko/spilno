import { expect, test } from '@playwright/test';

import { api, base, connected, connectStorage, problemOf } from './association-support';
import { newEmail, newFiling, openCard } from './support';
import { approveTransfer, pendingTransfer } from './transfer-support';

const someCase = '0198a0a0-0000-7000-8000-00000000ca5e';
const pdf = { name: 'estimate.pdf', mimeType: 'application/pdf', buffer: Buffer.from('%PDF-1.4 test estimate') };

test('US-08: a член правління connects the association account and every член sees its address (AC-27)', async ({ browser }) => {
  const c = await connected(browser, newFiling());
  const account = newEmail('osbb-storage');
  await connectStorage(c.head, account);
  await expect(c.head.getByText(`Сховище документів: ${account}`)).toBeVisible();
  await expect(c.head).toHaveURL(new RegExp(`/o/${c.associationId}/association`));

  // Every член правління reads it: a new голова, seated by a transfer, sees the same line.
  const newHead = 'Новий Голова Правлінович';
  const { page: incoming, id } = await pendingTransfer(browser, c.f, newEmail('newhead'), newHead);
  await approveTransfer(browser, c.f, id, newHead);
  await openCard(incoming);
  await expect(incoming.getByText(`Сховище документів: ${account}`)).toBeVisible();
});

test('AC-28: a document is refused until the storage is connected, and any член can connect it', async ({ browser }) => {
  const c = await connected(browser, newFiling());
  const refused = await problemOf(await api(c.head).post(`${base(c.associationId)}/cases/${someCase}/documents`, { multipart: { file: pdf } }));
  expect(refused).toMatchObject({ status: 409, code: 'governance.drive_not_connected', detail: expect.stringContaining('Спершу підключіть сховище документів ОСББ') });
});

test('AC-28b: the sign-in address of a член правління is refused as the storage account, nothing is connected', async ({ browser }) => {
  const c = await connected(browser, newFiling());
  await connectStorage(c.head, c.email);
  await expect(c.head.getByText('Це особистий акаунт члена правління — підключіть акаунт ОСББ')).toBeVisible();
  await expect(c.head.getByText('сховище не підключене')).toBeVisible();

  // The refusal is shown on the callback page; the card, opened again, shows the same: nothing was connected.
  await openCard(c.head);
  await expect(c.head.getByText('сховище не підключене')).toBeVisible();
  await expect(c.head.getByText(/Сховище документів:/)).toHaveCount(0);
});
