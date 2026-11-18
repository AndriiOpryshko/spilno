import { createMemoryHistory, createRootRoute, createRouter, RouterProvider } from '@tanstack/react-router';
import { render, screen, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { describe, expect, it } from 'vitest';

import type { AssociationCard as Card } from '@/api/gen/model';

import { AssociationCard } from './association-card';

const id = '00000000-0000-7000-8000-000000000002';
const extractId = '00000000-0000-7000-8000-00000000e001';

const withExtract: Card = {
  id,
  name: 'ОСББ «Тестовий Будинок»',
  edrpou: '12345678',
  address: 'м. Київ, вул. Тестова, 1',
  state: 'connected',
  head_full_name: 'Тестовий Голова Правління',
  connected_by_hand: false,
  history: [
    {
      kind: 'disabled',
      occurred_at: '2026-12-01T10:00:00Z',
      decision_id: null,
      filed_at: null,
      reason: 'Неоплачений договір',
      head_full_name: null,
      ended_head_full_name: null,
      signer_full_name: null,
      extract_id: null,
      extract_dated_on: null,
    },
    {
      kind: 'connected',
      occurred_at: '2026-11-11T08:05:00Z',
      decision_id: '00000000-0000-7000-8000-00000000f002',
      filed_at: '2026-11-10T09:00:00Z',
      reason: null,
      head_full_name: 'Тестовий Голова Правління',
      ended_head_full_name: null,
      signer_full_name: 'Підписант Тестовий',
      extract_id: extractId,
      extract_dated_on: '2026-11-08',
    },
  ],
};
const byHand: Card = { ...withExtract, edrpou: null, head_full_name: null, connected_by_hand: true, history: [] };

const transferEvent = {
  kind: 'transfer_approved' as const,
  occurred_at: '2026-12-05T08:05:00Z',
  decision_id: '00000000-0000-7000-8000-00000000f003',
  filed_at: '2026-12-03T09:00:00Z',
  reason: null,
  head_full_name: 'Новий Голова Правління',
  signer_full_name: 'Підписант Тестовий',
  extract_id: '00000000-0000-7000-8000-00000000e002',
  extract_dated_on: '2026-12-01',
};
const transferEndedAHead: Card = {
  ...withExtract,
  head_full_name: 'Новий Голова Правління',
  history: [{ ...transferEvent, ended_head_full_name: 'Колишній Голова Правління' }],
};
// AC-19c: the hand-made association had no голова entry, so the transfer ends no term.
const transferEndedNoOne: Card = {
  ...byHand,
  connected_by_hand: false,
  head_full_name: 'Новий Голова Правління',
  history: [{ ...transferEvent, ended_head_full_name: null }],
};

async function renderInRouter(ui: ReactNode) {
  const router = createRouter({
    routeTree: createRootRoute({ component: () => <>{ui}</> }),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  });
  render(<RouterProvider router={router} />);
  await waitFor(() => expect(document.body.textContent).not.toBe(''));
}

describe('AssociationCard (AC-29)', () => {
  it('shows the голова, the extract with its date and signer, the decision date and the history', async () => {
    await renderInRouter(<AssociationCard card={withExtract} />);
    const text = document.body.textContent ?? '';
    expect(text).toContain('Тестовий Голова Правління');
    expect(text).toContain('Підписант Тестовий');
    expect(text).toMatch(/08\.11\.2026|8 листопада 2026/);
    expect(text).toMatch(/11\.11\.2026|11 листопада 2026/);
    expect(text).toContain('Неоплачений договір');
    expect(screen.getAllByRole('link').some((a) => a.getAttribute('href')?.includes(extractId))).toBe(true);
    expect(text).not.toContain('підключено вручну, без витягу');
  });

  it('marks the association set up by hand and links to the запит form with its id (AC-19c)', async () => {
    await renderInRouter(<AssociationCard card={byHand} />);
    expect(document.body.textContent).toContain('підключено вручну, без витягу');
    const link = screen.getByRole('link', { name: 'Подати запит на передачу' });
    const href = link.getAttribute('href') ?? '';
    expect(href).toContain('/applications/new');
    expect(href).toContain('kind=transfer');
    expect(href).toContain(`association_id=${id}`);
  });

  it('offers the запит link on the card of an association connected by a заявка (AC-17)', async () => {
    await renderInRouter(<AssociationCard card={withExtract} />);
    const href = screen.getByRole('link', { name: 'Подати запит на передачу' }).getAttribute('href') ?? '';
    expect(href).toContain('/applications/new');
    expect(href).toContain('kind=transfer');
    expect(href).toContain(`association_id=${id}`);
  });

  it('shows the by-hand badge only on the hand-made association', async () => {
    await renderInRouter(<AssociationCard card={withExtract} />);
    expect(document.body.textContent).not.toContain('підключено вручну, без витягу');
  });
});

describe('AssociationCard history marks (AC-19c, AC-20)', () => {
  it('shows «повноваження завершено» with the name of the голова whose term the transfer ended', async () => {
    await renderInRouter(<AssociationCard card={transferEndedAHead} />);
    const ended = screen.getByText(/повноваження завершено/);
    expect(ended.textContent).toContain('Колишній Голова Правління');
    expect(ended.textContent).toMatch(/03\.12\.2026|3 грудня 2026/);
    expect(ended.textContent).toMatch(/05\.12\.2026|5 грудня 2026/);
  });

  it('shows «призначено» with the name of the new голова next to it', async () => {
    await renderInRouter(<AssociationCard card={transferEndedAHead} />);
    const appointed = screen.getByText(/призначено/);
    expect(appointed.textContent).toContain('Новий Голова Правління');
    expect(appointed.textContent).not.toContain('Колишній Голова Правління');
  });

  it('shows «призначено» only when the transfer ended no term (the association set up by hand)', async () => {
    await renderInRouter(<AssociationCard card={transferEndedNoOne} />);
    expect(screen.getByText(/призначено/).textContent).toContain('Новий Голова Правління');
    expect(document.body.textContent).not.toContain('повноваження завершено');
  });
});
