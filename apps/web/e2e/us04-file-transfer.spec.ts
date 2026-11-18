import { expect, test } from '@playwright/test';

import { addBoardMember, connected, openRoster } from './association-support';
import { connectAssociation, newEmail, newFiling, sessionFor } from './support';
import { fileTransfer, fileTransferAndConfirm, openTransferFormFromCard, submitTransfer } from './transfer-support';

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

test('US-04: a newly elected голова who is already on the roster files a запит from the card (AC-17)', async ({ browser }) => {
  const f = newFiling();
  const fullName = 'Новий Голова Правлінович';
  const c = await connected(browser, f);

  // The голова seats the person on the roster first: «whether or not they are already on its roster».
  const email = newEmail('newhead');
  await openRoster(c.head);
  await addBoardMember(c.head, { fullName, email, termStart: '2025-02-01' });

  // The person is a current член of this one association, so `/` opens it: the card is their way to the form.
  const member = await sessionFor(browser, email);
  await member.goto('/');
  await expect(member).toHaveURL(new RegExp(`/o/${c.associationId}/`));
  await openTransferFormFromCard(member);
  await submitTransfer(member, f, fullName);

  // Nothing of the current roster is revealed.
  await expect(member.getByText('Запит на передачу подано, очікує розгляду оператором')).toBeVisible();
  await expect(member.getByText(f.fullName)).toHaveCount(0);
});

test('US-04: a запит for a code that is not connected is refused and the заявка form offered (AC-18)', async ({ browser }) => {
  // The запит form opens from the refusal of a connected code; the запит then names a code that is not connected.
  const via = newFiling();
  await connectAssociation(browser, via, newEmail('head'));
  const page = await sessionFor(browser, newEmail('head'));
  await fileTransfer(page, newFiling(), 'Новий Голова Правлінович', via);
  await expect(page.getByText('Це ОСББ ще не підключене — подайте заявку на підключення')).toBeVisible();
  await page.getByRole('link', { name: 'Подати заявку на підключення' }).click();
  await expect(page.getByRole('heading', { name: 'Заявка на підключення' })).toBeVisible();
  await page.goto('/applications');
  await expect(page.getByText('У вас ще немає заявок')).toBeVisible();
});
