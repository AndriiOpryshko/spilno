import { readFile } from 'node:fs/promises';

import { expect, test } from '@playwright/test';

import {
  checkFor,
  extractFile,
  fileAndConfirm,
  fillFiling,
  newEmail,
  newFiling,
  openReviewFromQueue,
  operatorEmail,
  recordCheck,
  sessionFor,
  today,
} from './support';

const refiledName = 'Виправлений Голова Тестович';
const refiledExtract = { ...extractFile, name: 'extract-2.pdf', buffer: Buffer.from('%PDF-1.4 test extract, refiled') };

test('US-02: the applicant reads the status and the reason, fixes and refiles, the оператор approves the refiled заявка (AC-05, AC-06, AC-08, AC-14)', async ({
  browser,
}) => {
  const f = newFiling();
  const applicant = await sessionFor(browser, newEmail('head'));
  await fileAndConfirm(applicant, f);

  // AC-05: the status and the contact address, nothing of an association.
  await applicant.goto('/');
  await applicant.getByRole('link', { name: f.name, exact: true }).click();
  await expect(applicant.getByText(`очікує розгляду з ${today()}`)).toBeVisible();
  await expect(applicant.getByText(`Питання щодо заявки: ${operatorEmail}`)).toBeVisible();

  // AC-14: the оператор rejects with a reason.
  const operator = await sessionFor(browser, operatorEmail);
  await openReviewFromQueue(operator, f.name);
  await operator.getByRole('button', { name: 'Відхилити' }).click();
  await operator.getByLabel('Причина').fill('Витяг нечитабельний');
  await operator.getByRole('dialog').getByRole('button', { name: 'Відхилити' }).click();
  await expect(operator.getByText(/^Заявка відхилена/)).toBeVisible();

  // The applicant reads the reason on the заявка they have open.
  await applicant.reload();
  await expect(applicant.getByText('відхилена', { exact: true })).toBeVisible();
  await expect(applicant.getByText('Причина: Витяг нечитабельний')).toBeVisible();

  // AC-06: fix and refile with a new extract.
  await applicant.getByRole('button', { name: 'Виправити і подати знову' }).click();
  await fillFiling(applicant, { fullName: refiledName });
  await applicant.getByLabel('Витяг (до 5 МБ)').setInputFiles(refiledExtract);
  await applicant.getByRole('button', { name: 'Подати заявку' }).click();

  // After a refile the same заявка is pending again, with the rejection kept in its history.
  await expect(applicant.getByText(`очікує розгляду з ${today()}`)).toBeVisible();
  const history = applicant.getByRole('list', { name: 'Історія заявки' });
  await expect(history.getByText(/відхилено — Витяг нечитабельний/)).toBeVisible();

  // AC-08: the refiled заявка is back in the queue; the оператор works on the new filing, not the rejected one.
  await openReviewFromQueue(operator, f.name);
  await expect(operator.getByText(refiledName).first()).toBeVisible();
  await expect(operator.getByRole('heading', { name: 'Попередні подання' })).toBeVisible();
  await expect(operator.getByText('не перевірено').first()).toBeVisible();

  // The extract the оператор downloads is the refiled one.
  const downloading = operator.waitForEvent('download');
  await operator.getByRole('link', { name: 'Завантажити витяг' }).first().click();
  const download = await downloading;
  expect(download.suggestedFilename()).toBe(refiledExtract.name);
  expect(await readFile(await download.path())).toEqual(refiledExtract.buffer);

  await recordCheck(operator, checkFor({ ...f, fullName: refiledName }));
  await expect(operator.getByRole('status')).toHaveCount(0);
  await operator.getByRole('button', { name: 'Схвалити' }).click();
  await expect(operator.getByText(/^Заявка схвалена \d{2}\.\d{2}\.\d{4}/)).toBeVisible();

  // The applicant now holds the association, with the refiled name as its голова.
  await applicant.reload();
  await expect(applicant.getByText('схвалена', { exact: true })).toBeVisible();
  await applicant.goto('/');
  await applicant.getByRole('link', { name: 'ОСББ', exact: true }).click();
  await expect(applicant.getByRole('heading', { name: f.name })).toBeVisible();
  await expect(applicant.getByText(refiledName).first()).toBeVisible();
});
