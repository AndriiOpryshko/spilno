import { expect, test } from '@playwright/test';

import { api, base, connected, problemOf } from './association-support';
import { applicationIdOf, connectAssociation, fileAndConfirm, newEmail, newFiling, operatorEmail, sessionFor } from './support';

const noAccess = 'Немає доступу';
const missingId = '0198a0a0-0000-7000-8000-00000000dead';

test('AC-04d: two windows of one person hold two associations and keep them after reloads', async ({ browser }) => {
  const a = await connected(browser, newFiling());
  const personEmail = a.email;
  const fb = newFiling();
  // The same person's заявка for a second association is approved too.
  const { applicant } = await connectAssociation(browser, fb, personEmail);
  const memberships = (await (await api(applicant).get('/api/identity/memberships')).json()) as { items: { association_id: string }[] };
  expect(memberships.items).toHaveLength(2);
  const b = { associationId: memberships.items.map((m) => m.association_id).find((id) => id !== a.associationId) ?? '' };

  const first = await sessionFor(browser, personEmail);
  const second = await sessionFor(browser, personEmail);
  await first.goto('/');
  await expect(first.getByText('Оберіть ОСББ')).toBeVisible();
  await expect(first.getByRole('link', { name: a.f.name })).toBeVisible();
  await expect(first.getByRole('link', { name: fb.name })).toBeVisible();

  await first.goto(`/o/${a.associationId}/association`);
  await second.goto(`/o/${b.associationId}/association`);
  for (let i = 0; i < 2; i++) {
    await first.reload();
    await second.reload();
    await expect(first.getByRole('heading', { name: a.f.name })).toBeVisible();
    await expect(second.getByRole('heading', { name: fb.name })).toBeVisible();
    await expect(first).toHaveURL(new RegExp(`/o/${a.associationId}/`));
    await expect(second).toHaveURL(new RegExp(`/o/${b.associationId}/`));
  }
  await expect(first.getByRole('heading', { name: fb.name })).toHaveCount(0);
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
  await c.head.goto(`/o/${c.associationId}/association`);
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
  await operator.goto('/operator/applications');
  await expect(operator.getByRole('heading', { name: 'Заявки та запити' })).toBeVisible();

});

// An оператор seated on a roster cannot be made through the built surface: the roster write is a board-protocol stub
// and the оператор cannot approve a заявка they filed (AC-16c).
test.fixme('AC-16b: an оператор who is a член правління works in their own association only (needs board-protocol roster)', async () => {});
