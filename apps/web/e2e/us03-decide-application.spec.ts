import { expect, test } from '@playwright/test';

import { addBoardMember, connected, openRoster, rosterEntry } from './association-support';
import { checkFor, fileAndConfirm, newEmail, newFiling, openReviewFromQueue, recordCheck, sessionFor, signInByButton } from './support';

test('US-03: the оператор records the check and approves, the applicant opens the new association (AC-08, AC-09, AC-10)', async ({
  browser,
}) => {
  const f = newFiling();
  const applicant = await sessionFor(browser, newEmail('head'));
  await fileAndConfirm(applicant, f);

  // The оператор signs in with the button and reaches the заявка through the queue.
  const operator = await (await browser.newContext()).newPage();
  await signInByButton(operator);
  await openReviewFromQueue(operator, f.name);

  // AC-10: before the check every checked value says so; after it, the source and date.
  await expect(operator.getByText('не перевірено').first()).toBeVisible();
  await expect(operator.getByText('введено заявником').first()).toBeVisible();
  await recordCheck(operator, checkFor(f));
  await expect(operator.getByText(/^з витягу з ЄДР, внесено оператором/).first()).toBeVisible();
  await expect(operator.getByText('не перевірено')).toHaveCount(0);

  // AC-08: approve.
  await operator.getByRole('button', { name: 'Схвалити' }).click();
  await expect(operator.getByText(/^Заявка схвалена \d{2}\.\d{2}\.\d{4}/)).toBeVisible();

  // AC-09: the applicant opens the new association.
  await applicant.goto('/applications');
  await expect(applicant.getByText('схвалена', { exact: true })).toBeVisible();
  await applicant.getByRole('link', { name: `Перейти до ОСББ ${f.name}` }).click();
  await expect(applicant.getByText(f.name).first()).toBeVisible();
  await expect(applicant.getByText(f.fullName).first()).toBeVisible();
  await expect(applicant.getByText('сховище не підключене')).toBeVisible();
});

const registryEmpty = 'У реєстрі ще немає підписаних протоколів';

test('US-03: the approved голова adds a член правління, who then opens the association (AC-09)', async ({ browser }) => {
  const c = await connected(browser, newFiling());
  const member = { fullName: 'Член Правління Тестовий', email: newEmail('member'), termStart: '2025-02-01' };

  // The new голова opens the реєстр протоколів, still empty.
  await c.head.getByRole('link', { name: 'Реєстр протоколів' }).click();
  await expect(c.head.getByRole('heading', { name: 'Реєстр протоколів' })).toBeVisible();
  await expect(c.head.getByText(registryEmpty)).toBeVisible();

  // The голова is the one entry of the new roster, a member from the term start of the заявка.
  await openRoster(c.head);
  await expect(rosterEntry(c.head, c.f.fullName).getByText('голова правління', { exact: true })).toBeVisible();
  await expect(rosterEntry(c.head, c.f.fullName).getByText('з 01.01.2025')).toBeVisible();
  await expect(c.head.getByText('член правління', { exact: true })).toHaveCount(0);

  await addBoardMember(c.head, member);
  await expect(rosterEntry(c.head, member.fullName).getByText('з 01.02.2025')).toBeVisible();

  // The new член правління signs in and lands in the association, with the same roster.
  const added = await sessionFor(browser, member.email);
  await added.goto('/');
  await expect(added).toHaveURL(new RegExp(`/o/${c.associationId}/`));
  await openRoster(added);
  await expect(rosterEntry(added, member.fullName).getByText('член правління', { exact: true })).toBeVisible();
  await expect(rosterEntry(added, c.f.fullName).getByText('голова правління', { exact: true })).toBeVisible();
});
