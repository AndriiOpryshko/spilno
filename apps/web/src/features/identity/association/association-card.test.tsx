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
      signer_full_name: 'Підписант Тестовий',
      extract_id: extractId,
      extract_dated_on: '2026-11-08',
    },
  ],
};
const byHand: Card = { ...withExtract, edrpou: null, head_full_name: null, connected_by_hand: true, history: [] };

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

  it('does not offer the запит link once the extract is there', async () => {
    await renderInRouter(<AssociationCard card={withExtract} />);
    expect(screen.queryByRole('link', { name: 'Подати запит на передачу' })).toBeNull();
  });
});
