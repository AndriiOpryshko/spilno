import { readFile } from 'node:fs/promises';

import { expect, test } from '@playwright/test';

import {
  api,
  attachDocument,
  base,
  connected,
  connectStorage,
  disabled,
  enableAs,
  fileIdsOf,
  meetingTitles,
  memberIdOf,
  openNewCase,
  problemOf,
  reasonText,
  scheduleMeeting,
  seedSignedProtocol,
  suspendedText,
} from './association-support';
import { daysAgo, newEmail, newFiling, openCard } from './support';

const pdf = { name: 'estimate.pdf', mimeType: 'application/pdf', buffer: Buffer.from('%PDF-1.4 test estimate') };
const scan = { name: 'protocol-signed-scan.pdf', mimeType: 'application/pdf', buffer: Buffer.from('%PDF-1.4 test signed scan') };
const printCopy = { name: 'protocol-signing-copy.pdf', mimeType: 'application/pdf', buffer: Buffer.from('%PDF-1.4 test signing copy') };
const draft = { name: 'draft-estimate.pdf', mimeType: 'application/pdf', buffer: Buffer.from('%PDF-1.4 test draft') };
const subject = 'Затвердити кошторис ремонту даху';
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
  await openCard(c.head);
  await expect(c.head.getByText(suspendedText)).toBeVisible();

  // Nothing was saved: after enabling, only the meeting from before exists.
  const row = await enableAs(operator, c.f.name);
  await expect(row.getByText('Підключене')).toBeVisible(); // the enabling has landed before the member reads
  expect(await meetingTitles(c.head, c.associationId)).toEqual(['До вимкнення']);
});

test('AC-26: a member of a вимкнене ОСББ opens the реєстр and downloads the signed copy, byte for byte', async ({ browser }) => {
  const c = await connected(browser, newFiling());
  await connectStorage(c.head, newEmail('osbb-storage'));
  await expect(c.head.getByText(/^Сховище документів:/)).toBeVisible();

  // The files reach the storage the way a member puts them there; the signed protocol over them is seeded.
  const caseId = await openNewCase(c.head, c.associationId, 'Ремонт даху');
  for (const file of [scan, printCopy, draft]) {
    await attachDocument(c.head, file);
    await expect(c.head.getByRole('link', { name: file.name })).toBeVisible();
  }
  const files = await fileIdsOf(c.head, c.associationId, caseId);
  seedSignedProtocol({
    association: c.associationId,
    member: await memberIdOf(c.head, c.associationId),
    case_id: caseId,
    scan_file: files[scan.name]!,
    print_file: files[printCopy.name]!,
    number: 1,
    meeting_date: daysAgo(10),
    subject,
  });

  await disabled(browser, c);

  // The реєстр and the protocol open under the notice.
  await c.head.getByRole('link', { name: 'Реєстр протоколів' }).click();
  await expect(c.head.getByText(suspendedText)).toBeVisible();
  await c.head.getByRole('link', { name: subject }).click();
  await expect(c.head.getByRole('heading', { name: 'Протокол № 1' })).toBeVisible();
  await expect(c.head.getByText(suspendedText)).toBeVisible();

  const downloading = c.head.waitForEvent('download');
  await c.head.getByRole('link', { name: `Завантажити ${scan.name}` }).click();
  const download = await downloading;
  expect(download.suggestedFilename()).toBe(scan.name);
  expect(await readFile(await download.path())).toEqual(scan.buffer);

  // The search answers, and a file no signed protocol pins stays closed (ADR-0007).
  expect((await api(c.head).get(`${base(c.associationId)}/registry/decisions`)).status()).toBe(200);
  const unpinned = await problemOf(await api(c.head).get(`${base(c.associationId)}/files/${files[draft.name]}`));
  expect(unpinned).toMatchObject({ status: 403, code: 'identity.association_disabled', detail: reasonText });
});

test('AC-28: attaching a document to a справа names the missing storage and leads to connecting it', async ({ browser }) => {
  const c = await connected(browser, newFiling());
  await openNewCase(c.head, c.associationId, 'Заміна ліфта');
  await attachDocument(c.head, pdf, 409); // no storage is connected yet

  const refusal = c.head.getByRole('alert');
  await expect(refusal.getByText('Спершу підключіть сховище документів ОСББ')).toBeVisible();
  await expect(refusal.getByText('Підключити сховище може будь-який член правління')).toBeVisible();
  await expect(c.head.getByText('У справі ще немає документів. Додайте перший файл нижче.')).toBeVisible();

  await refusal.getByRole('link', { name: 'Підключити сховище' }).click();
  await expect(c.head).toHaveURL(new RegExp(`/o/${c.associationId}/association`));
  await expect(c.head.getByText('сховище не підключене')).toBeVisible();
});
