import { expect, test } from '@playwright/test';

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

// board-protocol T10 (the roster screen) is not merged yet: the step is kept, named and skipped.
test.fixme('US-03: the approved голова adds a член правління (AC-09, board-protocol T10 roster screen)', async () => {});
