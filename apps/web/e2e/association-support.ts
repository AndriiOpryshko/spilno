import { expect, type APIResponse, type Browser, type Page } from '@playwright/test';

import { connectAssociation, newEmail, operatorEmail, sessionFor, type Filing } from './support';
import { associationIdOf } from './transfer-support';

// Helpers of the disabling, read-only, storage and access specs (US-06, US-07, US-08, AC-04d, AC-07, AC-16, AC-16b).

export const reasonText = 'спір щодо складу правління';
export const suspendedText = `Доступ призупинено: ${reasonText}`;

export interface Problem {
  status: number;
  code?: string;
  detail?: string;
  title?: string;
}

// The association API as the signed-in person of this page sees it (the session cookie rides along).
export const api = (page: Page) => page.context().request;
export const base = (associationId: string) => `/api/governance/associations/${associationId}`;

export async function problemOf(res: APIResponse): Promise<Problem> {
  const body = (await res.json().catch(() => ({}))) as Omit<Problem, 'status'>;
  return { status: res.status(), code: body.code, detail: body.detail, title: body.title };
}

// The record a member changes: a meeting is the one governance write that is built today.
export const scheduleMeeting = (page: Page, associationId: string, title = 'Засідання правління') =>
  api(page).post(`${base(associationId)}/meetings`, { data: { title, held_at: '2026-12-01T10:00:00Z' } });

export const meetingTitles = async (page: Page, associationId: string) => {
  const body = (await (await api(page).get(`${base(associationId)}/meetings`)).json()) as { items: { title: string }[] };
  return body.items.map((m) => m.title);
};

export interface Connected {
  f: Filing;
  email: string;
  head: Page;
  associationId: string;
}

// A connected association with its голова signed in, entered through `/`.
export async function connected(browser: Browser, f: Filing): Promise<Connected> {
  const email = newEmail('head');
  const { applicant: head } = await connectAssociation(browser, f, email);
  const associationId = await associationIdOf(head);
  return { f, email, head, associationId };
}

const rowOf = (page: Page, name: string) => page.getByRole('listitem').filter({ hasText: name });

export async function openOperatorAssociations(page: Page, name: string) {
  await page.goto('/operator/associations');
  const row = rowOf(page, name);
  await expect(row).toBeVisible();
  return row;
}

export async function disableAs(operator: Page, name: string, reason: string) {
  const row = await openOperatorAssociations(operator, name);
  await row.getByRole('button', { name: 'Вимкнути' }).click();
  if (reason) await row.getByLabel('Причина').fill(reason);
  await row.getByRole('button', { name: 'Вимкнути' }).click();
  return row;
}

export async function enableAs(operator: Page, name: string) {
  const row = await openOperatorAssociations(operator, name);
  await row.getByRole('button', { name: 'Увімкнути' }).click();
  return row;
}

export async function operatorPage(browser: Browser) {
  return sessionFor(browser, operatorEmail);
}

export async function disabled(browser: Browser, c: Connected, reason = reasonText) {
  const operator = await operatorPage(browser);
  const row = await disableAs(operator, c.f.name, reason);
  await expect(row.getByText('Вимкнене')).toBeVisible();
  return operator;
}
