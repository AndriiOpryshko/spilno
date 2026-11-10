import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createMemoryHistory, createRootRoute, createRouter, RouterProvider } from '@tanstack/react-router';
import { cleanup, render, screen } from '@testing-library/react';
import type { ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import type { ApplicationReview } from '@/api/gen/model';

import { OperatorQueue } from './queue';
import { ReviewPage } from './review-page';

const appId = '00000000-0000-7000-8000-0000000000a1';
const filingId = '00000000-0000-7000-8000-0000000000f1';
const otherId = '00000000-0000-7000-8000-0000000000a2';

const application = {
  id: appId,
  kind: 'connection',
  edrpou: '12345678',
  association_id: null,
  association_name: 'ОСББ «Тестовий Будинок»',
  state: 'pending',
  filed_at: '2026-11-10T09:00:00Z',
  rejection_reason: null,
  created_at: '2026-11-10T09:00:00Z',
} as ApplicationReview['application'];

const base: ApplicationReview = {
  application,
  applicant_email: 'head@example.test',
  filings: [
    {
      id: filingId,
      association_name: 'ОСББ «Тестовий Будинок»',
      association_address: 'м. Київ, вул. Тестова, 1',
      applicant_full_name: 'Тестовий Голова Правління',
      position: 'head',
      term_start: '2026-01-01',
      term_end: null,
      extract: { id: 'e1', original_name: 'vytyah.pdf', media_type: 'application/pdf', size_bytes: 10, sha256: 'a'.repeat(64) },
      created_at: '2026-11-10T09:00:00Z',
    },
  ],
  checks: [],
  decisions: [],
  warnings: [],
  competing_application_ids: [],
  storage_account_email: null,
};

const checked: ApplicationReview = {
  ...base,
  checks: [
    {
      id: 'c1',
      filing_id: filingId,
      signature_valid: true,
      signer_full_name: 'Тестовий Голова Правління',
      signed_on: '2026-11-08',
      head_full_name: 'Тестовий Голова Правління',
      extract_edrpou: '12345678',
      extract_dated_on: '2026-11-05',
      checked_at: '2026-11-11T08:05:00Z',
    },
  ],
};

let review: ApplicationReview;
let queueStatus: number;
let urls: string[];

const json = (status: number, body: unknown) =>
  new Response(JSON.stringify(body), { status, headers: { 'Content-Type': status >= 400 ? 'application/problem+json' : 'application/json' } });

function fakeApi(input: RequestInfo | URL) {
  const path = new URL(String(input), 'http://localhost').pathname;
  urls.push(path);
  if (path.endsWith('/identity/operator/applications')) {
    if (queueStatus !== 200) return Promise.resolve(json(queueStatus, { type: 'about:blank', title: 'x', status: queueStatus, code: 'identity.not_operator' }));
    return Promise.resolve(json(200, { items: [review.application], next_cursor: null }));
  }
  if (path.endsWith(`/identity/operator/applications/${appId}`)) return Promise.resolve(json(200, review));
  return Promise.resolve(json(404, {}));
}

async function renderInRouter(ui: ReactNode) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  const router = createRouter({
    routeTree: createRootRoute({ component: () => <QueryClientProvider client={client}>{ui}</QueryClientProvider> }),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  });
  render(<RouterProvider router={router} />);
}

beforeEach(() => {
  review = base;
  queueStatus = 200;
  urls = [];
  vi.stubGlobal('fetch', vi.fn(fakeApi));
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe('ReviewPage', () => {
  it('marks typed values as entered by the applicant and checked ones as «не перевірено» before the check (AC-10)', async () => {
    await renderInRouter(<ReviewPage applicationId={appId} />);
    expect(await screen.findByText('Тестовий Голова Правління')).toBeInTheDocument();
    expect(screen.getAllByText(/введено заявником/).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/не перевірено/).length).toBeGreaterThan(0);
    expect(screen.queryByText(/перевірено оператором/)).not.toBeInTheDocument();
  });

  it('shows the three source labels with the check date after the check (AC-10)', async () => {
    review = checked;
    await renderInRouter(<ReviewPage applicationId={appId} />);
    await screen.findByText(/введено заявником/);
    expect(screen.getAllByText(/з витягу з ЄДР, внесено оператором (11\.11\.2026|11 листопада 2026)/).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/з КЕП, перевірено оператором (11\.11\.2026|11 листопада 2026)/).length).toBeGreaterThan(0);
    expect(screen.queryByText(/не перевірено/)).not.toBeInTheDocument();
  });

  it('names the differing values in the name warning and keeps «Схвалити» available (AC-11)', async () => {
    review = { ...checked, warnings: [{ kind: 'name_mismatch', detail: 'Іваненко І. І. ≠ Петренко П. П.' }] };
    await renderInRouter(<ReviewPage applicationId={appId} />);
    expect(await screen.findByText(/Іваненко І\. І\. ≠ Петренко П\. П\./)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Схвалити/ })).toBeEnabled();
  });

  it('warns that the extract is older than 30 days and keeps «Схвалити» available (AC-11b)', async () => {
    review = { ...checked, warnings: [{ kind: 'extract_age', detail: '2026-09-01' }] };
    await renderInRouter(<ReviewPage applicationId={appId} />);
    expect(await screen.findByText(/старший за 30 днів/)).toBeInTheDocument();
    expect(screen.getByText(/Витяг сформовано/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Схвалити/ })).toBeEnabled();
  });

  it('marks a competing заявка (AC-04c)', async () => {
    review = { ...base, competing_application_ids: [otherId] };
    await renderInRouter(<ReviewPage applicationId={appId} />);
    expect(await screen.findByText(/конкуруюча заявка/i)).toBeInTheDocument();
  });

  it('marks a competing запит (AC-19b)', async () => {
    review = { ...base, application: { ...application, kind: 'transfer' }, competing_application_ids: [otherId] };
    await renderInRouter(<ReviewPage applicationId={appId} />);
    expect(await screen.findByText(/конкуруючий запит/i)).toBeInTheDocument();
  });

  it('shows the storage account that holds the files of a запит (AC-28b)', async () => {
    review = { ...base, application: { ...application, kind: 'transfer' }, storage_account_email: 'files@example.test' };
    await renderInRouter(<ReviewPage applicationId={appId} />);
    expect(await screen.findByText(/files@example\.test/)).toBeInTheDocument();
  });

  it('offers the check form with the fields of the check', async () => {
    await renderInRouter(<ReviewPage applicationId={appId} />);
    await screen.findByText(/введено заявником/);
    expect(screen.getByLabelText(/Підписант/)).toBeInTheDocument();
    expect(screen.getByLabelText(/Керівник/)).toBeInTheDocument();
    expect(screen.getByLabelText(/ЄДРПОУ/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Записати перевірку/ })).toBeInTheDocument();
  });
});

describe('OperatorQueue', () => {
  it('lists the applications for the оператор (AC-16)', async () => {
    await renderInRouter(<OperatorQueue />);
    expect(await screen.findByText(/Тестовий Будинок/)).toBeInTheDocument();
  });

  it('shows the no-access page and no queue to a person who is not the оператор (AC-16)', async () => {
    queueStatus = 403;
    await renderInRouter(<OperatorQueue />);
    expect(await screen.findByText('Немає доступу')).toBeInTheDocument();
    expect(screen.queryByText(/Тестовий Будинок/)).not.toBeInTheDocument();
  });
});
