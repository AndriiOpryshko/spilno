import { expect, test } from '@playwright/test';

import {
  applicationIdOf,
  checkFor,
  connectAssociation,
  daysAgo,
  fileAndConfirm,
  fileConnection,
  newEmail,
  newFiling,
  openReview,
  operatorEmail,
  recordCheck,
  sessionFor,
} from './support';

const approve = (page: import('@playwright/test').Page) => page.getByRole('button', { name: 'Схвалити' }).click();

async function pendingFiling(browser: import('@playwright/test').Browser, email = newEmail('head')) {
  const f = newFiling();
  const applicant = await sessionFor(browser, email);
  await fileAndConfirm(applicant, f);
  const id = await applicationIdOf(applicant, f.name);
  return { f, applicant, id };
}

test.describe('filing refusals (US-01)', () => {
  test('AC-02: one submit marks every broken field and files nothing', async ({ browser }) => {
    const page = await sessionFor(browser, newEmail('head'));
    await page.goto('/applications/new?kind=connection');
    await page.getByRole('button', { name: 'Подати заявку' }).click();

    for (const field of ['edrpou', 'association_name', 'association_address', 'applicant_full_name', 'term_start', 'extract']) {
      await expect(page.locator(`#filing-${field}-error`), field).toBeVisible();
    }
    await page.goto('/applications');
    await expect(page.getByText('У вас ще немає заявок')).toBeVisible();
  });

  test('AC-30: a term that already ended is refused and nothing is recorded', async ({ browser }) => {
    const page = await sessionFor(browser, newEmail('head'));
    await fileConnection(page, { ...newFiling(), termEnd: daysAgo(1) });
    await expect(page.getByText('Строк повноважень голови вже завершився — подайте витяг після переобрання')).toBeVisible();
    await page.goto('/applications');
    await expect(page.getByText('У вас ще немає заявок')).toBeVisible();
  });

  test('AC-04b: a second заявка for the same code is refused and the open one is shown', async ({ browser }) => {
    const { f, applicant } = await pendingFiling(browser);
    await fileConnection(applicant, f);
    await expect(applicant.getByText('У вас уже є заявка на це ОСББ')).toBeVisible();
    await applicant.getByRole('link', { name: 'Мої заявки' }).click();
    await expect(applicant.getByRole('link', { name: f.name, exact: true })).toHaveCount(1);
  });

  test('AC-04: a connected code is refused and the запит на передачу form is offered', async ({ browser }) => {
    const f = newFiling();
    await connectAssociation(browser, f, newEmail('head'));
    const other = await sessionFor(browser, newEmail('head'));
    await fileConnection(other, f);
    await expect(
      other.getByText('Це ОСББ вже підключене — зверніться до голови вашого ОСББ або подайте запит на передачу'),
    ).toBeVisible();
    await other.getByRole('link', { name: 'Подати запит на передачу' }).click();
    await expect(other.getByRole('heading', { name: 'Запит на передачу' })).toBeVisible();
  });
});

test.describe('decision refusals (US-03)', () => {
  test('AC-03: approval is refused until the check is recorded', async ({ browser }) => {
    const { applicant, id } = await pendingFiling(browser);
    const operator = await sessionFor(browser, operatorEmail);
    await openReview(operator, id);
    await approve(operator);
    await expect(
      operator.getByText('Спершу перевірте підпис КЕП і внесіть підписанта, керівника, код і дату витягу'),
    ).toBeVisible();
    await applicant.goto(`/applications/${id}`);
    await expect(applicant.getByText(/^очікує розгляду з/)).toBeVisible();
  });

  test('AC-03b: an extract that names another code blocks approval', async ({ browser }) => {
    const { f, id } = await pendingFiling(browser);
    const operator = await sessionFor(browser, operatorEmail);
    await openReview(operator, id);
    await recordCheck(operator, checkFor(f, { edrpou: '87654321' }));
    await approve(operator);
    await expect(
      operator.getByText('Витяг стосується іншого ОСББ — відхиліть заявку або перевірте внесений код'),
    ).toBeVisible();
    await expect(operator.getByRole('button', { name: 'Схвалити' })).toBeVisible();
  });

  test('AC-04c: competing заявки are marked, and the second approval is refused', async ({ browser }) => {
    const f = newFiling();
    const first = await sessionFor(browser, newEmail('head'));
    const second = await sessionFor(browser, newEmail('head'));
    await fileAndConfirm(first, f);
    await fileAndConfirm(second, { ...f, name: `${f.name} bis` });
    const firstId = await applicationIdOf(first, f.name);
    const secondId = await applicationIdOf(second, `${f.name} bis`);

    const operator = await sessionFor(browser, operatorEmail);
    await openReview(operator, secondId);
    await expect(operator.getByText('конкуруюча заявка')).toBeVisible();
    await openReview(operator, firstId);
    await expect(operator.getByText('конкуруюча заявка')).toBeVisible();
    await recordCheck(operator, checkFor(f));
    await approve(operator);
    await expect(operator.getByText(/^Заявка схвалена/)).toBeVisible();

    await openReview(operator, secondId);
    await recordCheck(operator, checkFor(f));
    await approve(operator);
    await expect(
      operator.getByText('Це ОСББ вже підключене — зверніться до голови вашого ОСББ або подайте запит на передачу'),
    ).toBeVisible();

    await operator.getByRole('button', { name: 'Відхилити' }).click();
    await operator.getByLabel('Причина').fill('Це ОСББ уже підключене');
    await operator.getByRole('dialog').getByRole('button', { name: 'Відхилити' }).click();
    await expect(operator.getByText(/^Заявка відхилена/)).toBeVisible();
  });

  test('AC-11, AC-12: a розбіжність імен warns and needs the document to approve', async ({ browser }) => {
    const { f, id } = await pendingFiling(browser);
    const operator = await sessionFor(browser, operatorEmail);
    await openReview(operator, id);
    await recordCheck(operator, checkFor(f, { head: 'Інший Керівник Витягович' }));
    await expect(operator.getByRole('status').first()).toBeVisible();

    await approve(operator);
    await expect(
      operator.getByText('Схвалення з розбіжністю потребує документа: вкажіть, на чому ґрунтується рішення'),
    ).toBeVisible();

    await operator.getByLabel('Документ, на якому ґрунтується рішення').fill('протокол загальних зборів № 3 від 01.03.2026');
    await approve(operator);
    await expect(operator.getByText(/^Заявка схвалена/)).toBeVisible();
  });

  test('AC-11b: an extract older than 30 days warns and still can be approved', async ({ browser }) => {
    const { f, id } = await pendingFiling(browser);
    const operator = await sessionFor(browser, operatorEmail);
    await openReview(operator, id);
    await recordCheck(operator, checkFor(f, { datedOn: daysAgo(45) }));
    await expect(operator.getByText(/Витяг сформовано \S+ — старший за 30 днів/)).toBeVisible();
    await approve(operator);
    await expect(operator.getByText(/^Заявка схвалена/)).toBeVisible();
  });

  test('AC-13: rejecting without a reason is refused', async ({ browser }) => {
    const { applicant, id } = await pendingFiling(browser);
    const operator = await sessionFor(browser, operatorEmail);
    await openReview(operator, id);
    await operator.getByRole('button', { name: 'Відхилити' }).click();
    await operator.getByRole('dialog').getByRole('button', { name: 'Відхилити' }).click();
    await expect(operator.getByText('Вкажіть причину — її побачить заявник')).toBeVisible();
    await applicant.goto(`/applications/${id}`);
    await expect(applicant.getByText(/^очікує розгляду з/)).toBeVisible();
  });

  test('AC-15: a decided заявка cannot be decided again', async ({ browser }) => {
    const { f, id } = await pendingFiling(browser);
    const operator = await sessionFor(browser, operatorEmail);
    const stale = await operator.context().newPage();
    await openReview(operator, id);
    await openReview(stale, id);

    await recordCheck(operator, checkFor(f));
    await approve(operator);
    await expect(operator.getByText(/^Заявка схвалена/)).toBeVisible();

    await approve(stale);
    await expect(stale.getByText('Рішення за цією заявкою вже ухвалено')).toBeVisible();
    await openReview(operator, id);
    await expect(operator.getByText(/^Заявка схвалена/)).toBeVisible();
  });

  test('AC-16c: the оператор decides nothing they filed themselves', async ({ browser }) => {
    const f = newFiling();
    const operator = await sessionFor(browser, operatorEmail);
    await fileAndConfirm(operator, f);
    const id = await applicationIdOf(operator, f.name);
    await openReview(operator, id);
    await expect(operator.getByText('Це ваша власна заявка — оператор не ухвалює рішення щодо себе')).toBeVisible();
    await expect(operator.getByRole('button', { name: 'Схвалити' })).toHaveCount(0);
  });
});
