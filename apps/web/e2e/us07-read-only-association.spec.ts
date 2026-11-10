import { expect, test } from '@playwright/test';

import { api, base, connected, disabled, enableAs, meetingTitles, problemOf, reasonText, scheduleMeeting, suspendedText } from './association-support';
import { newFiling } from './support';

const pdf = { name: 'estimate.pdf', mimeType: 'application/pdf', buffer: Buffer.from('%PDF-1.4 test estimate') };
const someCase = '0198a0a0-0000-7000-8000-00000000ca5e';

test('US-07: in a disabled association every change is refused with the reason and nothing changes (AC-25)', async ({ browser }) => {
  const c = await connected(browser, newFiling());
  expect((await scheduleMeeting(c.head, c.associationId, 'До вимкнення')).status()).toBe(201);

  const operator = await disabled(browser, c);
  const refused = { status: 403, code: 'identity.association_disabled', detail: reasonText };

  // A meeting, the roster, the settings, a draft, a document and the storage connection are all refused.
  expect(await problemOf(await scheduleMeeting(c.head, c.associationId, 'Під час вимкнення'))).toMatchObject(refused);
  expect(await problemOf(await api(c.head).put(`${base(c.associationId)}/settings`, { data: { meeting_form: 'x', signers: 'y', register_start_number: 1 } }))).toMatchObject(refused);
  expect(
    await problemOf(await api(c.head).post(`${base(c.associationId)}/board-members`, { data: { full_name: 'Test User', term_start: '2025-01-01' } })),
  ).toMatchObject(refused);
  expect(await problemOf(await api(c.head).post(`${base(c.associationId)}/cases`, { data: { title: 'Нова справа' } }))).toMatchObject(refused);
  expect(
    await problemOf(await api(c.head).post(`${base(c.associationId)}/cases/${someCase}/documents`, { multipart: { file: pdf } })),
  ).toMatchObject(refused);
  expect(
    await problemOf(await api(c.head).get(`/api/identity/associations/${c.associationId}/storage-connection/authorization`, { maxRedirects: 0 })),
  ).toMatchObject(refused);

  // The member sees the notice on the screen.
  await c.head.goto(`/o/${c.associationId}/association`);
  await expect(c.head.getByText(suspendedText)).toBeVisible();

  // Nothing was saved: after enabling, only the meeting from before exists.
  const row = await enableAs(operator, c.f.name);
  await expect(row.getByText('Підключене')).toBeVisible(); // the enabling has landed before the member reads
  expect(await meetingTitles(c.head, c.associationId)).toEqual(['До вимкнення']);
});

test('AC-26: the registry of a disabled association is still served to its member', async ({ browser }) => {
  const c = await connected(browser, newFiling());
  await disabled(browser, c);
  const res = await api(c.head).get(`${base(c.associationId)}/registry/decisions`);
  const problem = await problemOf(res);
  expect(problem.code, 'a registry read is not refused for the suspension').not.toBe('identity.association_disabled');
  expect(res.status()).not.toBe(403);
});

// The signed copy is served only by the board-protocol registry screens.
test.fixme('AC-26: a member downloads a signed protocol while disabled (needs board-protocol T29, T31, T43)', async () => {});
test.fixme('AC-28: attaching a case document in the screen names the missing storage (needs board-protocol T21, T36)', async () => {});
