import { expect, test } from '@playwright/test';

import { checkFor, connectAssociation, newEmail, newFiling, openReview, operatorEmail, recordCheck, sessionFor } from './support';
import { approveTransfer, associationIdOf, pendingTransfer } from './transfer-support';

const newHead = 'Новий Голова Правлінович';

test('US-05: the оператор approves, the new голова takes over and the previous one is refused on the next action (AC-19)', async ({
  browser,
}) => {
  const f = newFiling();
  const oldEmail = newEmail('oldhead');
  const { applicant: previous } = await connectAssociation(browser, f, oldEmail);
  // The previous голова has the association open before the approval.
  const associationId = await associationIdOf(previous);
  await previous.goto(`/o/${associationId}/association`);
  await expect(previous.getByRole('heading', { name: f.name })).toBeVisible();

  const { page: incoming, id } = await pendingTransfer(browser, f, newEmail('newhead'), newHead);
  await approveTransfer(browser, f, id, newHead);

  // The new голова opens the association and sees themselves as the голова.
  await incoming.goto('/applications');
  await incoming.getByRole('link', { name: `Перейти до ОСББ ${f.name}` }).click();
  await expect(incoming.getByText(f.name).first()).toBeVisible();
  await incoming.goto(`/o/${associationId}/association`);
  await expect(incoming.getByRole('heading', { name: f.name })).toBeVisible();
  await expect(incoming.getByText(newHead).first()).toBeVisible();

  // The previous голова's open page: the next action is refused, no new sign-in.
  await previous.getByRole('link', { name: 'Справи' }).click();
  await expect(previous.getByRole('heading', { name: 'Немає доступу' })).toBeVisible();
  await previous.goto(`/o/${associationId}/association`);
  await expect(previous.getByRole('heading', { name: 'Немає доступу' })).toBeVisible();
});

test('US-05: approving a запит without a check, or with an extract of another code, is refused (AC-18b)', async ({ browser }) => {
  const f = newFiling();
  await connectAssociation(browser, f, newEmail('oldhead'));
  const { page: incoming, id } = await pendingTransfer(browser, f, newEmail('newhead'), newHead);

  const operator = await sessionFor(browser, operatorEmail);
  await openReview(operator, id);
  await operator.getByRole('button', { name: 'Схвалити' }).click();
  await expect(
    operator.getByText('Спершу перевірте підпис КЕП і внесіть підписанта, керівника, код і дату витягу'),
  ).toBeVisible();

  await recordCheck(operator, checkFor(f, { signer: newHead, head: newHead, edrpou: '87654321' }));
  await operator.getByRole('button', { name: 'Схвалити' }).click();
  await expect(operator.getByText('Витяг стосується іншого ОСББ — відхиліть заявку або перевірте внесений код')).toBeVisible();

  await incoming.goto(`/applications/${id}`);
  await expect(incoming.getByText(/^очікує розгляду з/)).toBeVisible();
  // The roster is unchanged: the new голова still has no association.
  await incoming.goto('/');
  await expect(incoming).toHaveURL(/\/applications/);
});
