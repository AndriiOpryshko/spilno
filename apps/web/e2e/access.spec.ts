import { expect, test } from '@playwright/test';

import { addBoardMember, api, base, connected, openRoster, problemOf, rosterEntry, scheduleMeeting } from './association-support';
import {
  applicationIdOf,
  approveApplication,
  checkFor,
  fileAndConfirm,
  newEmail,
  newFiling,
  openCard,
  openQueue,
  openReview,
  operatorEmail,
  sessionFor,
  type Filing,
} from './support';

const noAccess = 'Немає доступу';
const missingId = '0198a0a0-0000-7000-8000-00000000dead';

test('AC-04d: two windows of one person hold two associations and keep them after reloads', async ({ browser }) => {
  // The person files both заявки from their list of заявки, before either is approved: once on a single
  // roster, `/` opens that association, where no screen links to the form.
  const fa = newFiling();
  const fb = newFiling();
  const first = await sessionFor(browser, newEmail('head'));
  await fileAndConfirm(first, fa);
  await fileAndConfirm(first, fb);
  for (const f of [fa, fb]) await approveApplication(browser, await applicationIdOf(first, f.name), checkFor(f));
  const memberships = (await (await api(first).get('/api/identity/memberships')).json()) as {
    items: { association_id: string; association_name: string }[];
  };
  expect(memberships.items).toHaveLength(2);
  const idOf = (f: Filing) => memberships.items.find((m) => m.association_name === f.name)?.association_id ?? '';

  // The second window is a page of the same browser context: one sign-in, one session cookie. A choice kept
  // in the session would be shared by both windows.
  const second = await first.context().newPage();
  await first.goto('/');
  await expect(first.getByText('Оберіть ОСББ')).toBeVisible();
  await expect(first.getByRole('link', { name: fa.name })).toBeVisible();
  await expect(first.getByRole('link', { name: fb.name })).toBeVisible();

  // Each window picks its association and opens the card from the sidebar.
  await first.getByRole('link', { name: fa.name }).click();
  await openCard(first);
  await second.goto('/');
  await second.getByRole('link', { name: fb.name }).click();
  await openCard(second);
  for (let i = 0; i < 2; i++) {
    await first.reload();
    await second.reload();
    await expect(first.getByRole('heading', { name: fa.name })).toBeVisible();
    await expect(second.getByRole('heading', { name: fb.name })).toBeVisible();
    await expect(first).toHaveURL(new RegExp(`/o/${idOf(fa)}/`));
    await expect(second).toHaveURL(new RegExp(`/o/${idOf(fb)}/`));
  }
  await expect(first.getByRole('heading', { name: fb.name })).toHaveCount(0);
  await expect(second.getByRole('heading', { name: fa.name })).toHaveCount(0);

  // The next action of the first window, taken after the second one chose, still belongs to its association.
  await openRoster(first);
  await expect(first).toHaveURL(new RegExp(`/o/${idOf(fa)}/roster`));
  await expect(rosterEntry(first, fa.fullName).getByText('голова правління', { exact: true })).toBeVisible();
});

test('AC-07: someone off the roster is denied the same way whether the association exists or not', async ({ browser }) => {
  const c = await connected(browser, newFiling());
  const other = await connected(browser, newFiling());
  const applicant = await sessionFor(browser, newEmail('applicant'));
  await fileAndConfirm(applicant, newFiling());
  const nobody = await sessionFor(browser, newEmail('nobody'));

  for (const [who, page] of [['an applicant', applicant], ['nobody', nobody], ['a member of another association', other.head]] as const) {
    const real = await problemOf(await api(page).get(`${base(c.associationId)}/board-members`));
    const missing = await problemOf(await api(page).get(`${base(missingId)}/board-members`));
    expect(real, who).toMatchObject({ status: 403, code: 'identity.not_current_member' });
    expect(real, `${who}: the same answer for a missing association`).toEqual(missing);

    for (const path of ['cases', 'association']) {
      await page.goto(`/o/${c.associationId}/${path}`);
      await expect(page.getByRole('heading', { name: noAccess }), `${who} ${path}`).toBeVisible();
      await expect(page.getByText(c.f.name)).toHaveCount(0);
    }
  }
});

test('AC-16: a person who is not the оператор sees nothing of the queue and decides nothing', async ({ browser }) => {
  const c = await connected(browser, newFiling());
  const f = newFiling();
  const pending = await sessionFor(browser, newEmail('head'));
  await fileAndConfirm(pending, f);
  const pendingId = await applicationIdOf(pending, f.name);

  for (const [who, page] of [['a член правління', c.head], ['a голова with a pending заявка', pending]] as const) {
    await page.goto('/operator/applications');
    await expect(page.getByText(noAccess), who).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Заявки та запити' }), who).toHaveCount(0);

    const calls = {
      queue: await api(page).get('/api/identity/operator/applications'),
      review: await api(page).get(`/api/identity/operator/applications/${pendingId}`),
      reject: await api(page).post(`/api/identity/operator/applications/${pendingId}/rejection`, { data: { reason: 'спроба' } }),
      disable: await api(page).post(`/api/identity/operator/associations/${c.associationId}/disabling`, { data: { reason: 'спроба' } }),
      enable: await api(page).post(`/api/identity/operator/associations/${c.associationId}/enabling`),
    };
    for (const [name, res] of Object.entries(calls)) {
      expect(await problemOf(res), `${who}: ${name}`).toMatchObject({ status: 403, code: 'identity.not_operator' });
    }
  }

  // Nothing was decided or switched off.
  await pending.goto(`/applications/${pendingId}`);
  await expect(pending.getByText(/^очікує розгляду з/)).toBeVisible();
  await openCard(c.head);
  await expect(c.head.getByRole('heading', { name: c.f.name })).toBeVisible();
});

test('AC-16b: the оператор sees the queue and no association records', async ({ browser }) => {
  const foreign = await connected(browser, newFiling());
  const operator = await sessionFor(browser, operatorEmail);

  // The role alone opens no association: the same denial as for anyone off the roster (AC-07).
  const before = await problemOf(await api(operator).get(`${base(foreign.associationId)}/board-members`));
  expect(before).toMatchObject({ status: 403, code: 'identity.not_current_member' });
  await operator.goto(`/o/${foreign.associationId}/cases`);
  await expect(operator.getByRole('heading', { name: noAccess })).toBeVisible();
  await openQueue(operator);

});

// The оператор files a заявка from `/`, which offers the form on the picker but not inside a single association.
// The second test seats the оператор on one roster, so on a fresh database the filing has to come first.
test.describe('the оператор as an applicant and as a член правління', () => {
  test.describe.configure({ mode: 'serial' });

  test('AC-16c: the оператор decides nothing they filed themselves', async ({ browser }) => {
    const f = newFiling();
    const operator = await sessionFor(browser, operatorEmail);
    await fileAndConfirm(operator, f);
    const id = await applicationIdOf(operator, f.name);
    await openReview(operator, id);
    await expect(operator.getByText('Це ваша власна заявка — оператор не ухвалює рішення щодо себе')).toBeVisible();
    await expect(operator.getByRole('button', { name: 'Схвалити' })).toHaveCount(0);
  });

  test('AC-16b: an оператор who is a член правління works in their own association only', async ({ browser }) => {
    const own = await connected(browser, newFiling());
    const foreign = await connected(browser, newFiling());

    // The голова seats the оператор's address on the roster.
    await openRoster(own.head);
    await addBoardMember(own.head, { fullName: 'Оператор Правлінський Тестович', email: operatorEmail, termStart: '2025-02-01' });

    // `/` lands the оператор in the association, or on the picker when another run seated them elsewhere too.
    const operator = await sessionFor(browser, operatorEmail);
    await operator.goto('/');
    await expect(operator.getByRole('link', { name: 'Оператор' })).toBeVisible();
    const picked = operator.getByRole('link', { name: own.f.name, exact: true });
    if (await picked.isVisible()) await picked.click();
    await expect(operator).toHaveURL(new RegExp(`/o/${own.associationId}/`));

    // There they work as any член правління: they read the roster and change a record.
    await openRoster(operator);
    await expect(rosterEntry(operator, own.f.fullName).getByText('голова правління', { exact: true })).toBeVisible();
    expect((await scheduleMeeting(operator, own.associationId)).status()).toBe(201);

    // Another association stays closed to them, and the role is still theirs.
    const refused = await problemOf(await api(operator).get(`${base(foreign.associationId)}/board-members`));
    expect(refused).toMatchObject({ status: 403, code: 'identity.not_current_member' });
    expect(refused, 'the same answer as for a missing association (AC-07)').toEqual(await problemOf(await api(operator).get(`${base(missingId)}/board-members`)));
    await operator.goto(`/o/${foreign.associationId}/roster`);
    await expect(operator.getByRole('heading', { name: noAccess })).toBeVisible();
    await openQueue(operator);
  });
});
