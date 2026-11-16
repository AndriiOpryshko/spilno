import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createMemoryHistory, createRootRoute, createRouter, RouterProvider } from '@tanstack/react-router';
import { cleanup, render, screen, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { afterAll, afterEach, beforeAll, describe, expect, it, vi } from 'vitest';

import type { AssociationCard as Card } from '@/api/gen/model';

import { ApplicationDetail } from './applications/application-detail';
import { AssociationCard } from './association/association-card';
import { ReviewPage } from './operator/review-page';
import { ExtractSource } from './operator/source-label';

// The same decision, made at 2026-11-10T22:30Z, is 11.11.2026 in Kyiv on every screen (AC-05, AC-29).
const DECIDED = '2026-11-10T22:30:00Z';
const APP = '00000000-0000-7000-8000-0000000000a1';
const FILING = '00000000-0000-7000-8000-0000000000f1';

const originalTz = process.env.TZ;
beforeAll(() => {
  process.env.TZ = 'America/Los_Angeles';
});
afterAll(() => {
  if (originalTz === undefined) delete process.env.TZ;
  else process.env.TZ = originalTz;
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

const json = (body: unknown) =>
  new Response(JSON.stringify(body), { status: 200, headers: { 'Content-Type': 'application/json' } });

const application = {
  id: APP,
  kind: 'connection',
  edrpou: '12345678',
  association_id: null,
  association_name: 'ОСББ «Тест»',
  state: 'approved',
  filed_at: '2026-11-10',
  rejection_reason: null,
  created_at: '2026-11-10T10:00:00Z',
};
const filing = {
  id: FILING,
  association_name: 'ОСББ «Тест»',
  association_address: 'Київ, вул. Тестова, 1',
  applicant_full_name: 'Тестовий Голова',
  position: 'head',
  term_start: '2025-01-01',
  term_end: null,
  extract: { id: 'x1', original_name: 'extract.pdf', media_type: 'application/pdf', size_bytes: 8, sha256: 'a'.repeat(64) },
  created_at: '2026-11-10T10:00:00Z',
};
const decision = {
  id: '00000000-0000-7000-8000-0000000000d1',
  filing_id: FILING,
  outcome: 'approved',
  reason: null,
  relied_on_document: null,
  association_id: null,
  decided_at: DECIDED,
};

async function inRouter(ui: ReactNode) {
  const router = createRouter({
    routeTree: createRootRoute({
      component: () => (
        <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>{ui}</QueryClientProvider>
      ),
    }),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  });
  render(<RouterProvider router={router} />);
}

describe('decision at 2026-11-10T22:30Z shows 11.11.2026', () => {
  it("on the applicant's page", async () => {
    vi.stubGlobal('fetch', vi.fn(async () => json({ application, filings: [filing], decisions: [decision] })));
    await inRouter(<ApplicationDetail applicationId={APP} />);
    expect(await screen.findByText(/Рішення 11\.11\.2026/)).toBeInTheDocument();
  });

  it('in the operator panel', async () => {
    const review = { application, applicant_email: 'head@example.test', filings: [filing], checks: [], decisions: [decision], warnings: [], competing_application_ids: [], storage_account_email: null };
    vi.stubGlobal(
      'fetch',
      vi.fn(async (input: RequestInfo | URL) => {
        const p = new URL(String(input), 'http://localhost').pathname;
        return p.endsWith('/identity/session') ? json({ account_id: 'a', email: 'operator@example.test', is_operator: true }) : json(review);
      }),
    );
    await inRouter(<ReviewPage applicationId={APP} />);
    expect(await screen.findByText(/Заявка схвалена 11.11.2026/)).toBeInTheDocument();
  });

  it('on the association card', async () => {
    const card: Card = {
      id: '00000000-0000-7000-8000-000000000002',
      name: 'ОСББ «Тестовий Будинок»',
      edrpou: '12345678',
      address: 'м. Київ, вул. Тестова, 1',
      state: 'connected',
      head_full_name: 'Тестовий Голова',
      connected_by_hand: false,
      history: [
        {
          kind: 'connected',
          occurred_at: DECIDED,
          decision_id: '00000000-0000-7000-8000-00000000f002',
          filed_at: '2026-11-10T09:00:00Z',
          reason: null,
          head_full_name: 'Тестовий Голова',
          ended_head_full_name: null,
          signer_full_name: 'Підписант Тестовий',
          extract_id: 'e1',
          extract_dated_on: '2026-11-08',
        },
      ],
    };
    await inRouter(<AssociationCard card={card} />);
    await waitFor(() => expect(document.body.textContent).toContain('11.11.2026'));
    expect(document.body.textContent).toContain('08.11.2026');
  });

  it('in the extract source label of the operator panel', () => {
    render(<ExtractSource checkedAt={DECIDED} />);
    expect(screen.getByText(/11\.11\.2026/)).toBeInTheDocument();
  });
});
