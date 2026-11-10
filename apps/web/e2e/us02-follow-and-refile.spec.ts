import { expect, test } from '@playwright/test';

import {
  applicationIdOf,
  extractFile,
  fileAndConfirm,
  fillFiling,
  newEmail,
  newFiling,
  openReview,
  operatorEmail,
  sessionFor,
  today,
} from './support';

test('US-02: the applicant reads the status and the reason, fixes and refiles (AC-05, AC-06, AC-14)', async ({ browser }) => {
  const f = newFiling();
  const applicant = await sessionFor(browser, newEmail('head'));
  await fileAndConfirm(applicant, f);

  // AC-05: the status and the contact address, nothing of an association.
  const id = await applicationIdOf(applicant, f.name);
  await applicant.getByRole('link', { name: f.name, exact: true }).click();
  await expect(applicant.getByText(`очікує розгляду з ${today()}`)).toBeVisible();
  await expect(applicant.getByText(`Питання щодо заявки: ${operatorEmail}`)).toBeVisible();

  // AC-14: the оператор rejects with a reason.
  const operator = await sessionFor(browser, operatorEmail);
  await openReview(operator, id);
  await operator.getByRole('button', { name: 'Відхилити' }).click();
  await operator.getByLabel('Причина').fill('Витяг нечитабельний');
  await operator.getByRole('dialog').getByRole('button', { name: 'Відхилити' }).click();
  await expect(operator.getByText(/^Заявка відхилена/)).toBeVisible();

  // The applicant reads the reason after signing in.
  await applicant.goto(`/applications/${id}`);
  await expect(applicant.getByText('відхилена', { exact: true })).toBeVisible();
  await expect(applicant.getByText('Причина: Витяг нечитабельний')).toBeVisible();

  // AC-06: fix and refile with a new extract.
  await applicant.getByRole('button', { name: 'Виправити і подати знову' }).click();
  await fillFiling(applicant, { fullName: 'Тест Тестенко Тестович' });
  await applicant.getByLabel('Витяг (до 5 МБ)').setInputFiles({ ...extractFile, name: 'extract-2.pdf' });
  await applicant.getByRole('button', { name: 'Подати заявку' }).click();

  // After a refile the same заявка is pending again, with the rejection kept in its history.
  await expect(applicant.getByText(`очікує розгляду з ${today()}`)).toBeVisible();
  const history = applicant.getByRole('list', { name: 'Історія заявки' });
  await expect(history.getByText(/відхилено — Витяг нечитабельний/)).toBeVisible();
});
