import { expect, test } from '@playwright/test';

import { connected, disableAs, disabled, enableAs, openOperatorAssociations, operatorPage, problemOf, reasonText, scheduleMeeting, suspendedText } from './association-support';
import { newFiling } from './support';

test('US-06: the оператор disables with a reason, the member is stopped, enabling lets them change a record again (AC-21, AC-23)', async ({ browser }) => {
  const c = await connected(browser, newFiling());
  expect((await scheduleMeeting(c.head, c.associationId)).status(), 'a connected association takes a change').toBe(201);

  const operator = await disabled(browser, c);

  // AC-21: on the next action the member sees the reason.
  await c.head.goto(`/o/${c.associationId}/association`);
  await expect(c.head.getByText(suspendedText)).toBeVisible();
  const refused = await problemOf(await scheduleMeeting(c.head, c.associationId));
  expect(refused).toMatchObject({ status: 403, code: 'identity.association_disabled', detail: reasonText });

  // AC-23: enabling restores the association as it was.
  const row = await enableAs(operator, c.f.name);
  await expect(row.getByText('Підключене')).toBeVisible();
  await c.head.goto(`/o/${c.associationId}/association`);
  await expect(c.head.getByRole('heading', { name: c.f.name })).toBeVisible();
  expect((await scheduleMeeting(c.head, c.associationId)).status()).toBe(201);
});

test('AC-22: disabling without a reason is refused and the association stays підключене', async ({ browser }) => {
  const c = await connected(browser, newFiling());
  const operator = await operatorPage(browser);
  const row = await disableAs(operator, c.f.name, '');
  await expect(row.getByText('Вкажіть причину — її побачать члени правління')).toBeVisible();

  const again = await openOperatorAssociations(operator, c.f.name);
  await expect(again.getByText('Підключене')).toBeVisible();
  expect((await scheduleMeeting(c.head, c.associationId)).status(), 'the member keeps working').toBe(201);
});

test('AC-24: disabling a disabled and enabling an enabled association is refused, nothing changes', async ({ browser }) => {
  const c = await connected(browser, newFiling());
  const first = await operatorPage(browser);
  const second = await operatorPage(browser);
  const staleDisable = await openOperatorAssociations(second, c.f.name);

  await disableAs(first, c.f.name, reasonText);
  await staleDisable.getByRole('button', { name: 'Вимкнути' }).click();
  await staleDisable.getByLabel('Причина').fill('інша причина');
  await staleDisable.getByRole('button', { name: 'Вимкнути' }).click();
  await expect(staleDisable.getByText('ОСББ уже в цьому стані')).toBeVisible();

  // The recorded reason is the first one.
  await c.head.goto(`/o/${c.associationId}/association`);
  await expect(c.head.getByText(suspendedText)).toBeVisible();

  const staleEnable = await openOperatorAssociations(second, c.f.name);
  await enableAs(first, c.f.name);
  await staleEnable.getByRole('button', { name: 'Увімкнути' }).click();
  await expect(staleEnable.getByText('ОСББ уже в цьому стані')).toBeVisible();
  expect((await scheduleMeeting(c.head, c.associationId)).status()).toBe(201);
});
