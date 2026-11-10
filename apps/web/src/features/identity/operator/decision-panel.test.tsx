import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createMemoryHistory, createRootRoute, createRouter, RouterProvider } from '@tanstack/react-router';
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import type { ApplicationReview } from '@/api/gen/model';

import { ReviewPage } from './review-page';

const appId = '00000000-0000-7000-8000-0000000000a1';
const filingId = '00000000-0000-7000-8000-0000000000f1';

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

const pending: ApplicationReview = {
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

const operatorEmail = 'operator@example.test';

let review: ApplicationReview;
let sent: { path: string; body: unknown }[];
let refusal: { status: number; code: string } | null;

const json = (status: number, body: unknown) =>
  new Response(JSON.stringify(body), { status, headers: { 'Content-Type': status >= 400 ? 'application/problem+json' : 'application/json' } });

function decided(outcome: 'approved' | 'rejected', body: { reason?: string; relied_on_document?: string }): ApplicationReview {
  return {
    ...review,
    application: {
      ...review.application,
      state: outcome,
      rejection_reason: body.reason ?? null,
    } as ApplicationReview['application'],
    decisions: [
      {
        id: 'd1',
        filing_id: filingId,
        outcome,
        reason: body.reason ?? null,
        relied_on_document: body.relied_on_document ?? null,
        association_id: null,
        decided_at: '2026-11-12T08:00:00Z',
      },
    ],
  };
}

function fakeApi(input: RequestInfo | URL, init?: RequestInit) {
  const path = new URL(String(input), 'http://localhost').pathname;
  if (path.endsWith('/identity/session')) {
    return Promise.resolve(json(200, { account_id: 'acc-1', email: operatorEmail, is_operator: true }));
  }
  if (init?.method === 'POST') {
    const body = JSON.parse(String(init.body ?? '{}'));
    sent.push({ path, body });
    if (refusal) return Promise.resolve(json(refusal.status, { type: 'about:blank', title: 'x', status: refusal.status, code: refusal.code }));
    review = decided(path.endsWith('/approval') ? 'approved' : 'rejected', body);
    return Promise.resolve(json(200, review.decisions[0]));
  }
  if (path.endsWith(`/identity/operator/applications/${appId}`)) return Promise.resolve(json(200, review));
  return Promise.resolve(json(404, {}));
}

async function renderPage() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  const router = createRouter({
    routeTree: createRootRoute({
      component: () => (
        <QueryClientProvider client={client}>
          <ReviewPage applicationId={appId} />
        </QueryClientProvider>
      ),
    }),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  });
  render(<RouterProvider router={router} />);
  await screen.findByText(/введено заявником/);
}

const type = (el: HTMLElement, value: string) => fireEvent.change(el, { target: { value } });
const documentField = () => screen.getByLabelText(/документ, на якому ґрунтується рішення/i);

async function openReject() {
  fireEvent.click(await screen.findByRole('button', { name: 'Відхилити' }));
  const dialog = await screen.findByRole('dialog');
  return { dialog, confirm: within(dialog).getByRole('button', { name: /Відхилити/ }), reason: within(dialog).getByLabelText(/Причина/) };
}

beforeEach(() => {
  review = pending;
  sent = [];
  refusal = null;
  vi.stubGlobal('fetch', vi.fn(fakeApi));
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

const approvalRefusals: [string, number, string][] = [
  ['identity.signature_not_checked', 409, 'Спершу перевірте підпис КЕП і внесіть підписанта, керівника, код і дату витягу'],
  ['identity.extract_code_mismatch', 409, 'Витяг стосується іншого ОСББ — відхиліть заявку або перевірте внесений код'],
  ['identity.mismatch_needs_document', 422, 'Схвалення з розбіжністю потребує документа: вкажіть, на чому ґрунтується рішення'],
  ['identity.application_not_pending', 409, 'Рішення за цією заявкою вже ухвалено'],
  ['identity.operator_own_application', 409, 'Це ваша власна заявка — оператор не ухвалює рішення щодо себе'],
  ['identity.association_already_connected', 409, 'Це ОСББ вже підключене — зверніться до голови вашого ОСББ або подайте запит на передачу'],
];

describe('DecisionPanel on the review page', () => {
  it('approves after the document is typed, sends it and shows the new state with its date (AC-08, AC-12)', async () => {
    await renderPage();
    type(await screen.findByLabelText(/документ, на якому ґрунтується рішення/i), 'Протокол зборів № 5');
    fireEvent.click(screen.getByRole('button', { name: 'Схвалити' }));
    expect(await screen.findByText(/схвалена/)).toBeInTheDocument();
    expect(screen.getByText(/12\.11\.2026|12 листопада 2026/)).toBeInTheDocument();
    expect(sent).toHaveLength(1);
    expect(sent[0]?.path).toMatch(/\/approval$/);
    expect(sent[0]?.body).toMatchObject({ filing_id: filingId, relied_on_document: 'Протокол зборів № 5' });
    expect(screen.queryByRole('button', { name: 'Схвалити' })).not.toBeInTheDocument();
  });

  it('limits the document field to 500 characters (AC-12)', async () => {
    await renderPage();
    expect(documentField()).toHaveAttribute('maxlength', '500');
  });

  it('rejects with a reason in a dialog and shows the new state with its date (AC-13)', async () => {
    await renderPage();
    const { confirm, reason } = await openReject();
    type(reason, 'Витяг нечитабельний');
    fireEvent.click(confirm);
    expect(await screen.findByText(/відхилена/)).toBeInTheDocument();
    expect(screen.getByText(/12\.11\.2026|12 листопада 2026/)).toBeInTheDocument();
    expect(sent[0]?.path).toMatch(/\/rejection$/);
    expect(sent[0]?.body).toMatchObject({ filing_id: filingId, reason: 'Витяг нечитабельний' });
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('asks for a reason when the rejection is empty and does not send it (AC-13)', async () => {
    await renderPage();
    const { confirm } = await openReject();
    fireEvent.click(confirm);
    expect(await screen.findByText('Вкажіть причину — її побачить заявник')).toBeInTheDocument();
    expect(sent).toHaveLength(0);
  });

  it('shows the server message when the reason is refused as required (AC-13)', async () => {
    refusal = { status: 422, code: 'identity.reason_required' };
    await renderPage();
    const { confirm, reason } = await openReject();
    type(reason, 'x');
    fireEvent.click(confirm);
    expect(await screen.findByText('Вкажіть причину — її побачить заявник')).toBeInTheDocument();
  });

  it.each(approvalRefusals)('shows the PRD §5.1 message for %s on approval (AC-03, AC-03b, AC-12, AC-15, AC-16c, AC-18b)', async (code, status, message) => {
    refusal = { status, code };
    await renderPage();
    fireEvent.click(await screen.findByRole('button', { name: 'Схвалити' }));
    expect(await screen.findByText(message)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Схвалити' })).toBeInTheDocument();
  });

  it('shows the already-decided message when a rejection hits a decided заявка (AC-15)', async () => {
    refusal = { status: 409, code: 'identity.application_not_pending' };
    await renderPage();
    const { confirm, reason } = await openReject();
    type(reason, 'Причина');
    fireEvent.click(confirm);
    expect(await screen.findByText('Рішення за цією заявкою вже ухвалено')).toBeInTheDocument();
  });

  it('shows no decision buttons on the оператор\'s own filing (AC-16c)', async () => {
    review = { ...pending, applicant_email: operatorEmail };
    await renderPage();
    expect(await screen.findByText('Це ваша власна заявка — оператор не ухвалює рішення щодо себе')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Схвалити' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Відхилити' })).not.toBeInTheDocument();
  });
});
