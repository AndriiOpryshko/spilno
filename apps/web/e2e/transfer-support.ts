import { expect, type Browser, type Page } from '@playwright/test';

import { applicationIdOf, checkFor, daysAgo, extractFile, fillFiling, openReview, operatorEmail, recordCheck, sessionFor, type Filing } from './support';

// Helpers of the transfer flows (US-04, US-05, US-09), used by this task's specs only.

export const transferStart = () => daysAgo(5);

export async function fileTransfer(page: Page, f: Filing, fullName: string) {
  await page.goto('/applications/new?kind=transfer');
  await fillFiling(page, { code: f.code, name: f.name, address: f.address, fullName, termStart: transferStart() });
  await page.getByLabel('Витяг (до 5 МБ)').setInputFiles(extractFile);
  await page.getByRole('button', { name: 'Подати запит' }).click();
}

export async function fileTransferAndConfirm(page: Page, f: Filing, fullName: string) {
  await fileTransfer(page, f, fullName);
  await expect(page.getByText('Запит на передачу подано, очікує розгляду оператором')).toBeVisible();
}

// The new голова files a запит; returns their page and the id of the запит.
export async function pendingTransfer(browser: Browser, f: Filing, email: string, fullName: string) {
  const page = await sessionFor(browser, email);
  await fileTransferAndConfirm(page, f, fullName);
  const id = await applicationIdOf(page, f.name);
  return { page, id };
}

export async function approveTransfer(browser: Browser, f: Filing, id: string, fullName: string) {
  const operator = await sessionFor(browser, operatorEmail);
  await openReview(operator, id);
  await recordCheck(operator, checkFor(f, { signer: fullName, head: fullName }));
  await operator.getByRole('button', { name: 'Схвалити' }).click();
  await expect(operator.getByText(/^Заявка схвалена/)).toBeVisible();
  await operator.context().close();
}

// `/` sends a person with one association to it; the id is read from the address.
export async function associationIdOf(page: Page) {
  await page.goto('/');
  await page.waitForURL(/\/o\/[^/]+/);
  return /\/o\/([^/]+)/.exec(page.url())?.[1] ?? '';
}
