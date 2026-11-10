import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createMemoryHistory, createRootRoute, createRouter, RouterProvider } from '@tanstack/react-router';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { ApplicationDetail } from './application-detail';

const APP = '00000000-0000-7000-8000-0000000000a1';

const filing = (id: string, createdAt: string) => ({
  id,
  association_name: 'ОСББ «Тест»',
  association_address: 'Київ, вул. Тестова, 1',
  applicant_full_name: 'Тестовий Голова',
  position: 'head',
  term_start: '2025-01-01',
  term_end: null,
  extract: { id: `${id}-x`, original_name: 'extract.pdf', media_type: 'application/pdf', size_bytes: 8, sha256: 'a'.repeat(64) },
  created_at: createdAt,
});

const application = (state: string, filedAt: string, reason: string | null) => ({
  id: APP,
  kind: 'connection',
  edrpou: '12345678',
  association_id: null,
  association_name: 'ОСББ «Тест»',
  state,
  filed_at: filedAt,
  rejection_reason: reason,
  created_at: '2026-11-10T10:00:00Z',
});

const rejection = {
  id: '00000000-0000-7000-8000-0000000000d1',
  filing_id: '00000000-0000-7000-8000-0000000000f1',
  outcome: 'rejected',
  reason: 'Підпис КЕП недійсний',
  relied_on_document: null,
  association_id: null,
  decided_at: '2026-11-12T09:00:00Z',
};

const rejectedDetail = {
  application: application('rejected', '2026-11-10', 'Підпис КЕП недійсний'),
  filings: [filing('00000000-0000-7000-8000-0000000000f1', '2026-11-10T10:00:00Z')],
  decisions: [rejection],
};

const refiledDetail = {
  application: application('pending', '2026-11-14', null),
  filings: [
    filing('00000000-0000-7000-8000-0000000000f2', '2026-11-14T10:00:00Z'),
    filing('00000000-0000-7000-8000-0000000000f1', '2026-11-10T10:00:00Z'),
  ],
  decisions: [rejection],
};

const json = (body: unknown, status = 200) =>
  new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });

function stubApi(detail: () => unknown) {
  const fetchMock = vi.fn(async (_url: string | URL | Request, init?: RequestInit) => {
    const method = init?.method ?? 'GET';
    if (method === 'POST') return json(refiledDetail.application, 201);
    return json(detail());
  });
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}

async function renderDetail() {
  const router = createRouter({
    routeTree: createRootRoute({
      component: () => (
        <QueryClientProvider
          client={new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } })}
        >
          <ApplicationDetail applicationId={APP} />
        </QueryClientProvider>
      ),
    }),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  });
  render(<RouterProvider router={router} />);
}

afterEach(() => {
  vi.unstubAllGlobals();
  vi.unstubAllEnvs();
});

describe('ApplicationDetail', () => {
  it('shows a pending заявка with the date, the contact address and no association (AC-05)', async () => {
    vi.stubEnv('VITE_OPERATOR_CONTACT_EMAIL', 'operator@example.test');
    stubApi(() => ({
      application: application('pending', '2026-11-10', null),
      filings: [filing('00000000-0000-7000-8000-0000000000f1', '2026-11-10T10:00:00Z')],
      decisions: [],
    }));
    await renderDetail();
    expect(await screen.findByText(/очікує розгляду з 10\.11\.2026/)).toBeInTheDocument();
    expect(screen.getByText(/operator@example\.test/)).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /Виправити і подати знову/ })).not.toBeInTheDocument();
  });

  it('shows the rejection reason, the decision date and the history (AC-14)', async () => {
    stubApi(() => rejectedDetail);
    await renderDetail();
    expect(await screen.findByText(/відхилена/i)).toBeInTheDocument();
    expect(screen.getAllByText(/Підпис КЕП недійсний/).length).toBeGreaterThan(0);
    expect(screen.getByText(/10\.11\.2026/)).toBeInTheDocument();
    expect(screen.getByText(/12\.11\.2026/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Виправити і подати знову/ })).toBeInTheDocument();
  });

  it('opens the form filled from the latest filing and refiles into pending, keeping the rejection in the history (AC-06)', async () => {
    let refiled = false;
    const fetchMock = stubApi(() => (refiled ? refiledDetail : rejectedDetail));
    fetchMock.mockImplementation(async (_url, init) => {
      if ((init?.method ?? 'GET') === 'POST') {
        refiled = true;
        return json(refiledDetail.application, 201);
      }
      return json(refiled ? refiledDetail : rejectedDetail);
    });
    await renderDetail();
    fireEvent.click(await screen.findByRole('button', { name: /Виправити і подати знову/ }));

    expect(await screen.findByLabelText(/Назва ОСББ/)).toHaveValue('ОСББ «Тест»');
    expect(screen.getByLabelText(/ПІБ/)).toHaveValue('Тестовий Голова');
    const file = new File(['%PDF-1.7'], 'new.pdf', { type: 'application/pdf' });
    fireEvent.change(screen.getByLabelText(/Витяг/), { target: { files: [file] } });
    fireEvent.click(screen.getByRole('button', { name: /^Подати/ }));

    await waitFor(() => {
      const post = fetchMock.mock.calls.find(([, init]) => init?.method === 'POST');
      expect(String(post?.[0])).toBe(`/api/identity/applications/${APP}/filings`);
    });
    expect(await screen.findByText(/очікує розгляду з 14\.11\.2026/)).toBeInTheDocument();
    const history = screen.getByRole('list', { name: /Історія/ });
    expect(within(history).getByText(/Підпис КЕП недійсний/)).toBeInTheDocument();
    expect(within(history).getAllByRole('listitem').length).toBeGreaterThanOrEqual(2);
  });
});
