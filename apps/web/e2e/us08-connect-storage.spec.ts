import { expect, test } from '@playwright/test';

import { api, base, connected, problemOf } from './association-support';
import { newEmail, newFiling } from './support';
import { approveTransfer, pendingTransfer } from './transfer-support';

const someCase = '0198a0a0-0000-7000-8000-00000000ca5e';
const pdf = { name: 'estimate.pdf', mimeType: 'application/pdf', buffer: Buffer.from('%PDF-1.4 test estimate') };

// The fake consent (SPILNO_DEV_STORAGE) returns to the web callback with a code that is the account address;
// the spec rewrites the code of the redirect to choose the account that "gave the consent".
async function connectWith(page: import('@playwright/test').Page, associationId: string, account: string) {
  // A redirect hop is not offered to page.route, so the rewrite happens on the answer of the authorization call.
  await page.route('**/storage-connection/authorization', async (route) => {
    const response = await route.fetch({ maxRedirects: 0 });
    const url = new URL(response.headers()['location']!);
    url.searchParams.set('code', account);
    return route.fulfill({ response, headers: { ...response.headers(), location: url.toString() } });
  });
  await page.goto(`/o/${associationId}/association`);
  await expect(page.getByText('сховище не підключене')).toBeVisible();
  await page.getByRole('link', { name: 'Підключити сховище' }).click();
}

test('US-08: a член правління connects the association account and every член sees its address (AC-27)', async ({ browser }) => {
  const c = await connected(browser, newFiling());
  const account = newEmail('osbb-storage');
  await connectWith(c.head, c.associationId, account);
  await expect(c.head.getByText(`Сховище документів: ${account}`)).toBeVisible();
  await expect(c.head).toHaveURL(new RegExp(`/o/${c.associationId}/association`));

  // Every член правління reads it: a new голова, seated by a transfer, sees the same line.
  const newHead = 'Новий Голова Правлінович';
  const { page: incoming, id } = await pendingTransfer(browser, c.f, newEmail('newhead'), newHead);
  await approveTransfer(browser, c.f, id, newHead);
  await incoming.goto(`/o/${c.associationId}/association`);
  await expect(incoming.getByText(`Сховище документів: ${account}`)).toBeVisible();
});

test('AC-28: a document is refused until the storage is connected, and any член can connect it', async ({ browser }) => {
  const c = await connected(browser, newFiling());
  const refused = await problemOf(await api(c.head).post(`${base(c.associationId)}/cases/${someCase}/documents`, { multipart: { file: pdf } }));
  expect(refused).toMatchObject({ status: 409, code: 'governance.drive_not_connected', detail: expect.stringContaining('Спершу підключіть сховище документів ОСББ') });
});

test('AC-28b: the sign-in address of a член правління is refused as the storage account, nothing is connected', async ({ browser }) => {
  const c = await connected(browser, newFiling());
  await connectWith(c.head, c.associationId, c.email);
  await expect(c.head.getByText('Це особистий акаунт члена правління — підключіть акаунт ОСББ')).toBeVisible();
  await expect(c.head.getByText('сховище не підключене')).toBeVisible();

  await c.head.goto(`/o/${c.associationId}/association`);
  await expect(c.head.getByText('сховище не підключене')).toBeVisible();
  await expect(c.head.getByText(/Сховище документів:/)).toHaveCount(0);
});
