import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createMemoryHistory, createRootRoute, createRouter, RouterProvider } from '@tanstack/react-router';
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import type { ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { getDownloadExtractForReviewUrl } from '@/api/gen/endpoints/identity/identity';
import type { ApplicationReview } from '@/api/gen/model';

import { OperatorQueue } from './queue';
import { ReviewPage } from './review-page';

const appId = '00000000-0000-7000-8000-0000000000a1';
const filingId = '00000000-0000-7000-8000-0000000000f1';
const otherId = '00000000-0000-7000-8000-0000000000a2';
const refiledId = '00000000-0000-7000-8000-0000000000f2';

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
let sent: { path: string; body: unknown }[];

const json = (status: number, body: unknown) =>
  new Response(JSON.stringify(body), { status, headers: { 'Content-Type': status >= 400 ? 'application/problem+json' : 'application/json' } });

function fakeApi(input: RequestInfo | URL, init?: RequestInit) {
  const path = new URL(String(input), 'http://localhost').pathname;
  urls.push(path);
  if (init?.method === 'POST') {
    sent.push({ path, body: JSON.parse(String(init.body ?? '{}')) });
    return Promise.resolve(json(200, {}));
  }
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
  sent = [];
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

// A refiled заявка as the server returns it: filings newest first, checks oldest first (review.sql).
const oldFiling = {
  ...base.filings[0]!,
  id: filingId,
  applicant_full_name: 'Попередній Заявник',
  extract: { ...base.filings[0]!.extract, id: 'e1', original_name: 'old.pdf' },
  created_at: '2026-11-10T09:00:00Z',
};
const newFiling = {
  ...base.filings[0]!,
  id: refiledId,
  applicant_full_name: 'Новий Заявник',
  extract: { ...base.filings[0]!.extract, id: 'e2', original_name: 'new.pdf' },
  created_at: '2026-11-12T09:00:00Z',
};
const checkOf = (id: string, filing: string, signer: string, checkedAt: string) => ({
  id,
  filing_id: filing,
  signature_valid: true,
  signer_full_name: signer,
  signed_on: '2026-11-08',
  head_full_name: 'Керівник за витягом',
  extract_edrpou: '12345678',
  extract_dated_on: '2026-11-05',
  checked_at: checkedAt,
});
const refiled: ApplicationReview = {
  ...base,
  filings: [newFiling, oldFiling],
  checks: [
    checkOf('c1', filingId, 'Підписант першої заяви', '2026-11-11T08:05:00Z'),
    checkOf('c2', refiledId, 'Перший підписант нової заяви', '2026-11-13T09:00:00Z'),
    checkOf('c3', refiledId, 'Останній підписант нової заяви', '2026-11-14T10:00:00Z'),
  ],
};

describe('ReviewPage after a refiling', () => {
  beforeEach(() => {
    review = refiled;
  });

  it('presents the newest filing with the latest check of that filing (AC-06, AC-15)', async () => {
    await renderInRouter(<ReviewPage applicationId={appId} />);
    expect(await screen.findByText('Новий Заявник')).toBeInTheDocument();
    expect(screen.getByText('Останній підписант нової заяви')).toBeInTheDocument();
    expect(screen.getAllByText(/з КЕП, перевірено оператором (14\.11\.2026|14 листопада 2026)/).length).toBeGreaterThan(0);
    expect(screen.queryByText('Перший підписант нової заяви')).not.toBeInTheDocument();
  });

  it('lists the earlier filing as history and offers the check form once (AC-06)', async () => {
    await renderInRouter(<ReviewPage applicationId={appId} />);
    expect(await screen.findByText('Новий Заявник')).toBeInTheDocument();
    expect(screen.getByText('Попередній Заявник')).toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: /Записати перевірку/ })).toHaveLength(1);
    expect(screen.getAllByRole('button', { name: 'Схвалити' })).toHaveLength(1);
  });

  it('links the extract of each filing through the generated download URL (AC-03, AC-08)', async () => {
    await renderInRouter(<ReviewPage applicationId={appId} />);
    await screen.findByText('Новий Заявник');
    const hrefs = screen.getAllByRole('link', { name: /витяг/i }).map((a) => a.getAttribute('href'));
    expect(hrefs).toHaveLength(2);
    expect(hrefs).toContain(getDownloadExtractForReviewUrl(appId, refiledId));
    expect(hrefs).toContain(getDownloadExtractForReviewUrl(appId, filingId));
  });

  it('links the extract of a заявка with a single filing (AC-03)', async () => {
    review = base;
    await renderInRouter(<ReviewPage applicationId={appId} />);
    await screen.findByText(/введено заявником/);
    const links = screen.getAllByRole('link', { name: /витяг/i });
    expect(links).toHaveLength(1);
    expect(links[0]).toHaveAttribute('href', getDownloadExtractForReviewUrl(appId, filingId));
  });

  it('records the check against the newest filing (AC-03, AC-15)', async () => {
    await renderInRouter(<ReviewPage applicationId={appId} />);
    await screen.findByText('Новий Заявник');
    fireEvent.change(screen.getByLabelText(/Підписант/), { target: { value: 'Підписант' } });
    fireEvent.change(screen.getByLabelText(/Дата підпису/), { target: { value: '2026-11-08' } });
    fireEvent.change(screen.getByLabelText(/Керівник/), { target: { value: 'Керівник' } });
    fireEvent.change(screen.getByLabelText(/ЄДРПОУ/), { target: { value: '12345678' } });
    fireEvent.change(screen.getByLabelText(/Дата витягу/), { target: { value: '2026-11-05' } });
    fireEvent.click(screen.getByRole('button', { name: /Записати перевірку/ }));
    await waitFor(() => expect(sent).toHaveLength(1));
    expect(sent[0]?.path).toMatch(/\/checks$/);
    expect(sent[0]?.body).toMatchObject({ filing_id: refiledId });
  });

  it('approves the newest filing (AC-08)', async () => {
    await renderInRouter(<ReviewPage applicationId={appId} />);
    await screen.findByText('Новий Заявник');
    fireEvent.click(screen.getByRole('button', { name: 'Схвалити' }));
    await waitFor(() => expect(sent).toHaveLength(1));
    expect(sent[0]?.path).toMatch(/\/approval$/);
    expect(sent[0]?.body).toMatchObject({ filing_id: refiledId });
  });

  it('rejects the newest filing (AC-08)', async () => {
    await renderInRouter(<ReviewPage applicationId={appId} />);
    await screen.findByText('Новий Заявник');
    fireEvent.click(screen.getByRole('button', { name: 'Відхилити' }));
    const dialog = await screen.findByRole('dialog');
    fireEvent.change(within(dialog).getByLabelText(/Причина/), { target: { value: 'Витяг нечитабельний' } });
    fireEvent.click(within(dialog).getByRole('button', { name: /Відхилити/ }));
    await waitFor(() => expect(sent).toHaveLength(1));
    expect(sent[0]?.path).toMatch(/\/rejection$/);
    expect(sent[0]?.body).toMatchObject({ filing_id: refiledId });
  });
});

describe('OperatorQueue', () => {
  it('lists the applications for the оператор (AC-16)', async () => {
    await renderInRouter(<OperatorQueue />);
    expect(await screen.findByText(/Тестовий Будинок/)).toBeInTheDocument();
  });

  it('shows no queue to a person who is not the оператор, and leaves the refusal to the section guard (AC-16)', async () => {
    queueStatus = 403;
    await renderInRouter(<OperatorQueue />);
    await waitFor(() => expect(urls.some((path) => path.endsWith('/identity/operator/applications'))).toBe(true));
    expect(screen.queryByText(/Тестовий Будинок/)).not.toBeInTheDocument();
    expect(screen.queryByText('Немає доступу')).not.toBeInTheDocument();
  });
});
