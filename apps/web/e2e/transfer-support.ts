import { expect, type Browser, type Page } from '@playwright/test';

import { applicationIdOf, approveApplication, checkFor, daysAgo, extractFile, fileConnection, fillFiling, openCard, sessionFor, type Filing } from './support';

// Helpers of the transfer flows (US-04, US-05, US-09), used by this task's specs only.

export const transferStart = () => daysAgo(5);

// The extract of a запит differs from the one of the заявка, so a download shows which decision it belongs to.
export const transferExtract = { ...extractFile, name: 'transfer-extract.pdf', buffer: Buffer.from('%PDF-1.4 test extract of a transfer') };

// The way in for a person on the roster (AC-17): the link on the card of an association they are in.
export async function openTransferFormFromCard(page: Page) {
  await openCard(page);
  await page.getByRole('link', { name: 'Подати запит на передачу' }).click();
  await expect(page.getByRole('heading', { name: 'Запит на передачу' })).toBeVisible();
}

// The way in for a person off the roster: the refusal of a заявка на підключення for a
// code that is already connected (AC-04). The person files that заявка with their own name for `via`, a
// connected association, is refused (nothing is recorded) and follows the link.
export async function openTransferForm(page: Page, via: Filing, fullName: string) {
  await fileConnection(page, { code: via.code, name: via.name, address: via.address, fullName, termStart: transferStart() });
  await page.getByRole('link', { name: 'Подати запит на передачу' }).click();
  await expect(page.getByRole('heading', { name: 'Запит на передачу' })).toBeVisible();
}

// Fills the open запит form and submits it.
export async function submitTransfer(page: Page, f: Filing, fullName: string) {
  await fillFiling(page, { code: f.code, name: f.name, address: f.address, fullName, termStart: transferStart() });
  await page.getByLabel('Витяг (до 5 МБ)').setInputFiles(transferExtract);
  await page.getByRole('button', { name: 'Подати запит' }).click();
}

// `via` differs from `f` only when the запит itself names a code that is not connected (AC-18).
export async function fileTransfer(page: Page, f: Filing, fullName: string, via: Filing = f) {
  await openTransferForm(page, via, fullName);
  await submitTransfer(page, f, fullName);
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

export const approveTransfer = (browser: Browser, f: Filing, id: string, fullName: string) =>
  approveApplication(browser, id, checkFor(f, { signer: fullName, head: fullName }));

// `/` sends a person with one association to it; the id is read from the address.
export async function associationIdOf(page: Page) {
  await page.goto('/');
  await page.waitForURL(/\/o\/[^/]+/);
  return /\/o\/([^/]+)/.exec(page.url())?.[1] ?? '';
}
