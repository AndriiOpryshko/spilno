import { createMemoryHistory, createRootRoute, createRouter, RouterProvider } from '@tanstack/react-router';
import { render, screen, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { describe, expect, it } from 'vitest';

import type { AssociationEvent } from '@/api/gen/model';

import { RosterDecisionMark } from './roster-decision-mark';

const associationId = '00000000-0000-7000-8000-000000000002';
const decisionId = '00000000-0000-7000-8000-00000000f002';
const extractId = '00000000-0000-7000-8000-00000000e001';

const approval: AssociationEvent = {
  kind: 'transfer_approved',
  occurred_at: '2026-11-11T08:05:00Z',
  decision_id: decisionId,
  filed_at: '2026-11-10T09:00:00Z',
  reason: null,
  head_full_name: 'Тестовий Голова Правління',
  ended_head_full_name: null,
  signer_full_name: 'Тестовий Голова Правління',
  extract_id: extractId,
  extract_dated_on: '2026-11-08',
};
const filedOn = /10\.11\.2026|10 листопада 2026/;
const approvedOn = /11\.11\.2026|11 листопада 2026/;

async function renderInRouter(ui: ReactNode) {
  const router = createRouter({
    routeTree: createRootRoute({ component: () => <>{ui}</> }),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  });
  render(<RouterProvider router={router} />);
  await waitFor(() => expect(document.body.textContent).not.toBe(''));
}

describe('RosterDecisionMark (AC-20)', () => {
  it('shows the appointed text with the filing and the decision dates and the extract link', async () => {
    await renderInRouter(<RosterDecisionMark kind="appointed" decisionId={decisionId} history={[approval]} associationId={associationId} />);
    const text = document.body.textContent ?? '';
    expect(text).toContain('призначено за запитом на передачу від');
    expect(text).toMatch(filedOn);
    expect(text).toContain('схвалено оператором');
    expect(text).toMatch(approvedOn);
    expect(screen.getByRole('link').getAttribute('href')).toContain(extractId);
  });

  it('shows the ended text with both dates', async () => {
    await renderInRouter(<RosterDecisionMark kind="ended" decisionId={decisionId} history={[approval]} associationId={associationId} />);
    const text = document.body.textContent ?? '';
    expect(text).toContain('повноваження завершено за запитом на передачу від');
    expect(text).toMatch(filedOn);
    expect(text).toMatch(approvedOn);
  });

  it('renders nothing when the decision is not in the history', async () => {
    const { container } = render(<RosterDecisionMark kind="appointed" decisionId="00000000-0000-7000-8000-0000000000ff" history={[approval]} associationId={associationId} />);
    expect(container.textContent).toBe('');
  });

  it('words the mark of a connection by the заявка, not by a запит на передачу', async () => {
    const connection: AssociationEvent = { ...approval, kind: 'connected' };
    await renderInRouter(<RosterDecisionMark kind="appointed" decisionId={decisionId} history={[connection]} associationId={associationId} />);
    const text = document.body.textContent ?? '';
    expect(text).toContain('призначено за заявкою на підключення від');
    expect(text).toMatch(filedOn);
    expect(text).not.toContain('запитом на передачу');
    expect(screen.getByRole('link').getAttribute('href')).toContain(extractId);
  });

  it('words the mark of a transfer by the запит на передачу, for both effects', async () => {
    await renderInRouter(<RosterDecisionMark kind="appointed" decisionId={decisionId} history={[approval]} associationId={associationId} />);
    expect(document.body.textContent).toContain('призначено за запитом на передачу від');
    expect(document.body.textContent).not.toContain('заявкою на підключення');
  });
});

