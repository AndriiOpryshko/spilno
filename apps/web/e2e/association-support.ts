import { execFileSync } from 'node:child_process';
import path from 'node:path';

import { expect, type APIResponse, type Browser, type Page } from '@playwright/test';

import { connectAssociation, newEmail, openCard, openQueue, operatorEmail, sessionFor, type Filing } from './support';
import { associationIdOf } from './transfer-support';

// Helpers of the disabling, read-only, storage, roster and access specs (US-03, US-06, US-07, US-08, AC-04d, AC-07, AC-16, AC-16b).

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

// The оператор's list of associations, reached from the queue by the section's own navigation.
export async function openOperatorAssociations(page: Page, name: string) {
  await openQueue(page);
  await page.getByRole('link', { name: 'ОСББ', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'ОСББ', exact: true })).toBeVisible();
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

// The fake consent (SPILNO_DEV_STORAGE) returns to the web callback with a code that is the account address;
// the rewrite of the redirect chooses the account that "gave the consent".
export async function connectStorage(page: Page, account: string) {
  // A redirect hop is not offered to page.route, so the rewrite happens on the answer of the authorization call.
  await page.route('**/storage-connection/authorization', async (route) => {
    const response = await route.fetch({ maxRedirects: 0 });
    const url = new URL(response.headers()['location']!);
    url.searchParams.set('code', account);
    return route.fulfill({ response, headers: { ...response.headers(), location: url.toString() } });
  });
  await openCard(page);
  await expect(page.getByText('сховище не підключене')).toBeVisible();
  await page.getByRole('link', { name: 'Підключити сховище' }).click();
}

export async function openRoster(page: Page) {
  await page.getByRole('link', { name: 'Правління', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'Правління', level: 1 })).toBeVisible();
}

export const rosterEntry = (page: Page, fullName: string) => page.getByRole('listitem').filter({ hasText: fullName });

export async function addBoardMember(page: Page, member: { fullName: string; email: string; termStart: string }) {
  await page.getByRole('button', { name: 'Додати члена правління' }).click();
  const form = page.getByRole('dialog');
  await form.getByLabel('ПІБ').fill(member.fullName);
  await form.getByLabel('Електронна пошта').fill(member.email);
  await form.getByLabel('Початок строку').fill(member.termStart);
  await form.getByRole('button', { name: 'Зберегти' }).click();
  await expect(rosterEntry(page, member.fullName).getByText('член правління', { exact: true })).toBeVisible();
}

// No screen creates a справа yet, so the spec makes one through the API and opens it from the list.
export async function openNewCase(page: Page, associationId: string, title: string) {
  const res = await api(page).post(`${base(associationId)}/cases`, { data: { title } });
  expect(res.status(), 'the справа is created').toBe(201);
  const { id } = (await res.json()) as { id: string };
  // A справи list that is already on the screen does not re-read on a click on its own link.
  await page.reload();
  await page.getByRole('link', { name: 'Справи' }).click();
  await page.getByRole('link', { name: title }).click();
  await expect(page.getByRole('heading', { name: title })).toBeVisible();
  return id;
}

export interface Upload {
  name: string;
  mimeType: string;
  buffer: Buffer;
}

export async function attachDocument(page: Page, file: Upload) {
  await page.getByLabel('Файл (PDF, JPEG або PNG, до 20 МБ)').setInputFiles(file);
  await page.getByRole('button', { name: 'Додати документ' }).click();
}

// The ids of the files behind the documents of a справа, by file name.
export async function fileIdsOf(page: Page, associationId: string, caseId: string) {
  const body = (await (await api(page).get(`${base(associationId)}/cases/${caseId}`)).json()) as {
    documents: { file: { id: string; original_name: string } }[];
  };
  return Object.fromEntries(body.documents.map((d) => [d.file.original_name, d.file.id]));
}

export interface SignedSeed {
  association: string;
  member: string;
  case_id: string;
  scan_file: string;
  print_file: string;
  number: number;
  meeting_date: string;
  subject: string;
}

// A signed protocol written over the owner connection (seed/signed-protocol.sql). The address comes only from
// E2E_OWNER_DATABASE_URL, and the script itself refuses a database that is not marked as the e2e one.
export function seedSignedProtocol(seed: SignedSeed) {
  const url = process.env.E2E_OWNER_DATABASE_URL;
  expect(url, 'E2E_OWNER_DATABASE_URL: the owner connection of the e2e database (see the e2e job in ci.yml)').toBeTruthy();
  const vars = Object.entries(seed).flatMap(([name, value]) => ['-v', `${name}=${value}`]);
  execFileSync('psql', [url!, '-q', '-v', 'ON_ERROR_STOP=1', ...vars, '-f', path.join(import.meta.dirname, 'seed/signed-protocol.sql')], {
    stdio: ['ignore', 'ignore', 'pipe'],
  });
}

// The board member id of the signed-in person in the association, from their memberships.
export async function memberIdOf(page: Page, associationId: string) {
  const body = (await (await api(page).get('/api/identity/memberships')).json()) as { items: { association_id: string; board_member_id: string }[] };
  return body.items.find((m) => m.association_id === associationId)?.board_member_id ?? '';
}
