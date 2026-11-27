import { expect, test } from '@playwright/test';

import { openRoster, rosterEntry } from './association-support';
import { checkFor, connectAssociation, newEmail, newFiling, openCard, openReview, operatorEmail, recordCheck, sessionFor, today } from './support';
import { approveTransfer, associationIdOf, pendingTransfer } from './transfer-support';

const newHead = 'Новий Голова Правлінович';
const renamed = 'Попередній Голова Виправлений';

test('US-05: the оператор approves, the new голова takes over and the previous one is refused on the next action (AC-19)', async ({
  browser,
}) => {
  const f = newFiling();
  const oldEmail = newEmail('oldhead');
  const { applicant: previous } = await connectAssociation(browser, f, oldEmail);
  // The previous голова has the association open before the approval.
  const associationId = await associationIdOf(previous);
  await openCard(previous);
  await expect(previous.getByRole('heading', { name: f.name })).toBeVisible();

  const { page: incoming, id } = await pendingTransfer(browser, f, newEmail('newhead'), newHead);
  await approveTransfer(browser, f, id, newHead);

  // The new голова opens the association and sees themselves as the голова.
  await incoming.goto('/applications');
  await incoming.getByRole('link', { name: `Перейти до ОСББ ${f.name}` }).click();
  await expect(incoming.getByText(f.name).first()).toBeVisible();
  await openCard(incoming);
  await expect(incoming.getByRole('heading', { name: f.name })).toBeVisible();
  await expect(incoming.getByText(newHead).first()).toBeVisible();

  // The previous голова's open page: the next action is refused, no new sign-in.
  await previous.getByRole('link', { name: 'Справи' }).click();
  await expect(previous.getByRole('heading', { name: 'Немає доступу' })).toBeVisible();
  // The address typed by hand answers the same: the page of no access has no navigation to click.
  await previous.goto(`/o/${associationId}/association`);
  await expect(previous.getByRole('heading', { name: 'Немає доступу' })).toBeVisible();
});

test('US-05: a fix of the previous голова\'s name leaves the marks of the transfer on the same entries (AC-20, AC-29)', async ({ browser }) => {
  const f = newFiling();
  await connectAssociation(browser, f, newEmail('oldhead'));
  const { page: incoming, id } = await pendingTransfer(browser, f, newEmail('newhead'), newHead);
  await approveTransfer(browser, f, id, newHead);

  // The new голова corrects the name on the entry the approval ended; its term is shown, not offered for a change.
  await incoming.goto('/');
  await openRoster(incoming);
  await incoming.getByRole('button', { name: `Змінити ${f.fullName}` }).click();
  const form = incoming.getByRole('dialog');
  await expect(form.getByText('Щоб повернути людину до правління, додайте запис знову')).toBeVisible();
  await expect(form.getByLabel('Кінець строку')).toHaveCount(0);
  await form.getByLabel('ПІБ').fill(renamed);
  await form.getByRole('button', { name: 'Зберегти' }).click();
  await expect(form).toHaveCount(0);

  // The roster: the edited entry is still the ended one, the new голова still the appointed one.
  const basis = `за запитом на передачу від ${today()}, схвалено оператором ${today()}`;
  await expect(incoming.getByText(f.fullName)).toHaveCount(0);
  const ended = rosterEntry(incoming, renamed);
  await expect(ended.getByText(`повноваження завершено ${basis}`)).toBeVisible();
  await expect(ended.getByText('призначено')).toHaveCount(0);
  const appointed = rosterEntry(incoming, newHead);
  await expect(appointed.getByText(`призначено ${basis}`)).toBeVisible();
  await expect(appointed.getByText('повноваження завершено')).toHaveCount(0);

  // The card: the previous голова is named as ended under their corrected name, the new one is the голова.
  await openCard(incoming);
  await expect(incoming.getByText(`повноваження завершено: ${renamed} ${basis}`)).toBeVisible();
  await expect(incoming.getByText(`призначено: ${newHead} ${basis}`)).toBeVisible();
  await expect(incoming.getByText(/^повноваження завершено:/)).toHaveCount(1);
  const current = incoming.locator('section').filter({ has: incoming.getByRole('heading', { name: 'Голова правління' }) });
  await expect(current.getByText(newHead, { exact: true })).toBeVisible();
  await expect(current.getByText(renamed)).toHaveCount(0);
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
