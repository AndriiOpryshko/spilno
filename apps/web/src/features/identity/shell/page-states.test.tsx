import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createMemoryHistory, createRouter, RouterProvider } from '@tanstack/react-router';
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { routeTree } from '@/routeTree.gen';

import { uk } from '@/i18n/uk';

import { storageFlowKey } from '../storage/storage-line';

// Every page that reads the API through the real route tree: a skeleton while the read is out,
// sign-in on 401, «Немає доступу» on the access codes, and one alert with «Спробувати ще раз» on a
// 5xx or a network error (AC-05, AC-07, review W2). The fake answers like the server; one endpoint at
// a time is made to misbehave.

const lypky = '00000000-0000-7000-8000-000000000001';
const appId = '00000000-0000-7000-8000-0000000000a1';
const principal = { account_id: '00000000-0000-7000-8000-00000000a001', email: 'user-1@example.test', is_operator: false };
const retryName = 'Спробувати ще раз';

const json = (status: number, body: unknown) =>
  new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': status >= 400 ? 'application/problem+json' : 'application/json' },
  });
const problem = (status: number, code?: string) => json(status, { type: 'about:blank', title: 'Problem', status, detail: 'detail', ...(code ? { code } : {}) });

const membership = {
  association_id: lypky,
  association_name: 'ОСББ «Липки»',
  association_state: 'connected',
  board_member_id: '00000000-0000-7000-8000-00000000b001',
  position: 'member',
};
const notConnected = { connected: false, google_account_email: null, connected_by_member_id: null, connected_at: null };
const connected = { ...notConnected, connected: true, google_account_email: 'osbb-storage@example.test', connected_at: '2026-11-10T10:00:00Z' };
const application = {
  id: appId,
  kind: 'connection',
  edrpou: '12345678',
  association_id: null,
  association_name: 'ОСББ «Тест»',
  state: 'pending',
  filed_at: '2026-11-10',
  rejection_reason: null,
  created_at: '2026-11-10T10:00:00Z',
};

type Answer = Response | Promise<Response>;
const otherId = '00000000-0000-7000-8000-0000000000a2';
const review = {
  application: { ...application, edrpou: '12345678' },
  filings: [],
  checks: [],
  decisions: [],
  warnings: [],
  competing_application_ids: [otherId],
  applicant_email: 'applicant@example.test',
  storage_account_email: null,
};
let asOperator = false;
let rosters: unknown[];
let applications: unknown[];
let broken: { match: (path: string, method: string) => boolean; answer: () => Answer } | null;
let calls: { path: string; method: string }[];

function fakeApi(input: RequestInfo | URL, init?: RequestInit): Answer {
  const path = new URL(String(input), 'http://localhost').pathname.replace(/^\/api/, '');
  const method = init?.method ?? 'GET';
  calls.push({ path, method });
  if (broken?.match(path, method)) return broken.answer();
  if (path === '/identity/session') return json(200, { ...principal, is_operator: asOperator });
  if (path === '/identity/operator/applications') return json(200, { items: [{ ...application, competing: true }], next_cursor: null });
  if (path === `/identity/operator/applications/${appId}`) return json(200, review);
  if (path === `/identity/operator/applications/${otherId}`) {
    return json(200, { ...review, application: { ...application, id: otherId, association_name: 'ОСББ «Конкурент»' }, competing_application_ids: [appId] });
  }
  if (path === '/identity/operator/associations') {
    return json(200, { items: [{ id: lypky, name: 'ОСББ «Липки»', edrpou: '12345678', state: 'connected' }], next_cursor: null });
  }
  if (path === '/identity/memberships') return json(200, { items: rosters, next_cursor: null });
  if (path === '/identity/applications') return json(200, { items: applications, next_cursor: null });
  if (path === `/identity/applications/${appId}`) {
    return json(200, { application, filings: [], decisions: [] });
  }
  if (path === `/identity/associations/${lypky}`) {
    return json(200, { id: lypky, name: 'ОСББ «Липки»', edrpou: '12345678', address: 'м. Київ, вул. Тестова, 1', state: 'connected', head_full_name: 'Test User', connected_by_hand: false, history: [] });
  }
  if (path === `/identity/associations/${lypky}/storage-connection`) return json(method === 'POST' ? 201 : 200, method === 'POST' ? connected : notConnected);
  return json(404, {});
}

function open(initial: string) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  const router = createRouter({ routeTree, context: { queryClient }, history: createMemoryHistory({ initialEntries: [initial] }) });
  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
  return router;
}

const skeleton = () => document.querySelector('[data-slot="skeleton"]');
const forever = () => new Promise<Response>(() => {});
const breakPath = (path: string, method: string, answer: () => Answer) => {
  broken = { match: (p, m) => p === path && m === method, answer };
};
const countOf = (path: string, method: string) => calls.filter((c) => c.path === path && c.method === method).length;

beforeEach(() => {
  rosters = [membership];
  applications = [application];
  broken = null;
  asOperator = false;
  calls = [];
  sessionStorage.clear();
  vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => fakeApi(input, init)));
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

interface Page {
  name: string;
  url: string;
  // The one read the page cannot do without.
  path: string;
  // What to wait for on a healthy page, so a failure is not mistaken for a slow render.
  settled: () => Promise<unknown>;
  // The page is in the оператор's section: the session is an оператор's.
  operator?: boolean;
}

const pages: Page[] = [
  { name: 'the entry page: the session', url: '/', path: '/identity/session', settled: () => screen.findByRole('heading', { name: 'Справи' }) },
  { name: 'the entry page: the memberships', url: '/', path: '/identity/memberships', settled: () => screen.findByRole('heading', { name: 'Справи' }) },
  { name: 'the association layout: the session', url: `/o/${lypky}/cases`, path: '/identity/session', settled: () => screen.findByRole('heading', { name: 'Справи' }) },
  { name: 'the association layout: the memberships', url: `/o/${lypky}/cases`, path: '/identity/memberships', settled: () => screen.findByRole('heading', { name: 'Справи' }) },
  { name: 'the association card', url: `/o/${lypky}/association`, path: `/identity/associations/${lypky}`, settled: () => screen.findByRole('heading', { level: 1, name: 'ОСББ «Липки»' }) },
  { name: 'my applications', url: '/applications/', path: '/identity/applications', settled: () => screen.findByText('ОСББ «Тест»') },
  { name: 'an application', url: `/applications/${appId}`, path: `/identity/applications/${appId}`, settled: () => screen.findByRole('heading', { name: 'ОСББ «Тест»' }) },
  { name: "the оператор's review page", url: `/operator/applications/${appId}`, path: `/identity/operator/applications/${appId}`, operator: true, settled: () => screen.findByRole('heading', { level: 1, name: 'ОСББ «Тест»' }) },
  { name: "the оператор's competing item", url: `/operator/applications/${appId}`, path: `/identity/operator/applications/${otherId}`, operator: true, settled: () => screen.findByRole('link', { name: 'ОСББ «Конкурент»' }) },
  { name: "the оператор's associations list", url: '/operator/associations', path: '/identity/operator/associations', operator: true, settled: () => screen.findByRole('heading', { level: 1, name: uk.operator.associationsTitle }) },
];

describe.each(pages)('$name', ({ url, path, settled, operator }) => {
  beforeEach(() => {
    asOperator = operator === true;
  });

  it('shows a skeleton while the read is out, not a blank page or an empty list', async () => {
    breakPath(path, 'GET', forever);
    open(url);
    await waitFor(() => expect(skeleton()).not.toBeNull());
    expect(screen.queryByRole('alert')).toBeNull();
    expect(screen.queryByText('У вас ще немає заявок')).toBeNull();
  });

  it('sends a 401 to sign-in', async () => {
    breakPath(path, 'GET', () => problem(401, 'identity.session_required'));
    const router = open(url);
    await waitFor(() => expect(router.state.location.pathname).toBe('/sign-in'));
  });

  it.each(['identity.not_current_member', 'identity.not_operator'])('sends a 403 %s to the no-access page', async (code) => {
    breakPath(path, 'GET', () => problem(403, code));
    const router = open(url);
    await waitFor(() => expect(router.state.location.pathname).toBe('/no-access'));
    expect(await screen.findByRole('heading', { name: 'Немає доступу' })).toBeInTheDocument();
  });

  it.each([
    ['a 404', () => problem(404)],
    ['a 500', () => problem(500)],
    ['a network error', () => Promise.reject(new TypeError('Failed to fetch'))],
  ])('shows one alert with «Спробувати ще раз» on %s, stays on the page and reads again on retry', async (_what, answer) => {
    breakPath(path, 'GET', answer);
    const router = open(url);
    const alert = await screen.findByRole('alert');
    expect(router.state.location.pathname).not.toBe('/sign-in');
    expect(router.state.location.pathname).not.toBe('/no-access');
    // Stays on the page it was opened at (so the entry page is not sent on to /applications/, while
    // "my applications" is itself served there); a trailing slash is not a different page.
    const bare = (p: string) => p.replace(/\/+$/, '') || '/';
    expect(bare(router.state.location.pathname)).toBe(bare(url));
    const retry = within(alert).getByRole('button', { name: retryName });

    broken = null;
    const before = countOf(path, 'GET');
    fireEvent.click(retry);
    await settled();
    expect(countOf(path, 'GET')).toBeGreaterThan(before);
    expect(screen.queryByRole('alert')).toBeNull();
  });
});

// The оператор's section: only a 401 means «sign in»; a 5xx or a network error on the session read
// is a failed request, not an expired session (T67, AC-05, AC-07).
describe('the operator layout: the session', () => {
  const session = '/identity/session';

  it.each(['/operator/applications', '/operator/associations'])('sends a 401 on %s to sign-in', async (url) => {
    breakPath(session, 'GET', () => problem(401, 'identity.session_required'));
    const router = open(url);
    await waitFor(() => expect(router.state.location.pathname).toBe('/sign-in'));
  });

  it.each([
    ['a 500', () => problem(500)],
    ['a network error', () => Promise.reject(new TypeError('Failed to fetch'))],
  ])('shows one alert with «Спробувати ще раз» on %s, does not send the person to sign-in and reads again on retry', async (_what, answer) => {
    breakPath(session, 'GET', answer);
    const router = open('/operator/applications');
    const alert = await screen.findByRole('alert');
    expect(router.state.location.pathname).not.toBe('/sign-in');
    expect(router.state.location.pathname).toBe('/operator/applications');
    expect(screen.queryByRole('link', { name: 'Увійти' })).toBeNull();
    const retry = within(alert).getByRole('button', { name: retryName });

    broken = null;
    const before = countOf(session, 'GET');
    fireEvent.click(retry);
    await waitFor(() => expect(countOf(session, 'GET')).toBeGreaterThan(before));
    // The fake knows this person as no оператор: the retried read is answered, and the guard routes on it.
    await waitFor(() => expect(router.state.location.pathname).toBe('/no-access'));
  });
});

describe('the entry page', () => {
  it('sends a person on no roster to their applications through the router, not by reloading the page', async () => {
    rosters = [];
    const router = open('/');
    await waitFor(() => expect(router.state.location.pathname).toMatch(/^\/applications\/?$/));
    expect(await screen.findByRole('heading', { name: 'Мої заявки' })).toBeInTheDocument();
  });

  it('does not send a member to /applications/ when the memberships fail with a 5xx', async () => {
    breakPath('/identity/memberships', 'GET', () => problem(503));
    const router = open('/');
    expect(await screen.findByRole('alert')).toBeInTheDocument();
    expect(router.state.location.pathname).toBe('/');
  });
});

describe('the link to the заявка на підключення form (AC-01)', () => {
  const linkName = 'Подати заявку на підключення';
  const target = '/applications/new?kind=connection';

  it('is on an empty list of applications, next to the empty message', async () => {
    applications = [];
    open('/applications/');
    await screen.findByText('У вас ще немає заявок');
    expect(screen.getByRole('link', { name: linkName })).toHaveAttribute('href', target);
  });

  it('is above a list that has a заявка', async () => {
    open('/applications/');
    const first = await screen.findByRole('link', { name: 'ОСББ «Тест»' });
    const link = screen.getByRole('link', { name: linkName });
    expect(link).toHaveAttribute('href', target);
    expect(link.compareDocumentPosition(first) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });
});

describe('the storage line on the association card', () => {
  const storage = `/identity/associations/${lypky}/storage-connection`;
  const card = () => screen.findByRole('heading', { level: 1, name: 'ОСББ «Липки»' });

  it('shows a skeleton while the storage connection is read', async () => {
    breakPath(storage, 'GET', forever);
    open(`/o/${lypky}/association`);
    await card();
    await waitFor(() => expect(within(screen.getByRole('main')).queryByText('сховище не підключене')).toBeNull());
    await waitFor(() => expect(skeleton()).not.toBeNull());
  });

  it('sends a 401 to sign-in', async () => {
    breakPath(storage, 'GET', () => problem(401, 'identity.session_required'));
    const router = open(`/o/${lypky}/association`);
    await waitFor(() => expect(router.state.location.pathname).toBe('/sign-in'));
  });

  it('sends a 403 not_current_member to the no-access page', async () => {
    breakPath(storage, 'GET', () => problem(403, 'identity.not_current_member'));
    const router = open(`/o/${lypky}/association`);
    await waitFor(() => expect(router.state.location.pathname).toBe('/no-access'));
  });

  it('shows an alert with «Спробувати ще раз» on a 5xx and the line after the retry', async () => {
    breakPath(storage, 'GET', () => problem(500));
    open(`/o/${lypky}/association`);
    await card();
    const alert = await screen.findByRole('alert');
    broken = null;
    fireEvent.click(within(alert).getByRole('button', { name: retryName }));
    expect(await screen.findByText('сховище не підключене')).toBeInTheDocument();
    expect(screen.queryByRole('alert')).toBeNull();
  });
});

describe('the storage callback', () => {
  const storage = `/identity/associations/${lypky}/storage-connection`;
  const callback = '/storage/callback?code=auth-code&state=auth-state';

  beforeEach(() => sessionStorage.setItem(storageFlowKey, lypky));

  it('shows a skeleton while the connection is being completed', async () => {
    breakPath(storage, 'POST', forever);
    open(callback);
    await waitFor(() => expect(skeleton()).not.toBeNull());
    expect(screen.queryByRole('alert')).toBeNull();
  });

  it('sends a 401 to sign-in', async () => {
    breakPath(storage, 'POST', () => problem(401, 'identity.session_required'));
    const router = open(callback);
    await waitFor(() => expect(router.state.location.pathname).toBe('/sign-in'));
  });

  it('sends a 403 not_current_member to the no-access page', async () => {
    breakPath(storage, 'POST', () => problem(403, 'identity.not_current_member'));
    const router = open(callback);
    await waitFor(() => expect(router.state.location.pathname).toBe('/no-access'));
  });

  it.each([
    ['a 500', () => problem(500)],
    ['a network error', () => Promise.reject(new TypeError('Failed to fetch'))],
  ])('shows an alert with «Спробувати ще раз» on %s and posts again on retry', async (_what, answer) => {
    breakPath(storage, 'POST', answer);
    const router = open(callback);
    const alert = await screen.findByRole('alert');
    expect(router.state.location.pathname).toBe('/storage/callback');
    expect(alert).not.toHaveTextContent('особистий акаунт');
    broken = null;
    fireEvent.click(within(alert).getByRole('button', { name: retryName }));
    await waitFor(() => expect(countOf(storage, 'POST')).toBe(2));
  });

  it('keeps the personal-account message for that refusal only', async () => {
    breakPath(storage, 'POST', () => problem(409, 'identity.storage_account_personal'));
    open(callback);
    expect(await screen.findByText('Це особистий акаунт члена правління — підключіть акаунт ОСББ')).toBeInTheDocument();
  });

  // Not completed, and the cause is the request itself (a bad or replayed state, a refused code, a вимкнене
  // ОСББ): reading it again would post the same single-use code, so no retry; the way out is the card (T74).
  it.each([
    ['a 422 for a bad or replayed state', () => json(422, { type: 'about:blank', title: 'Validation failed', status: 422, errors: [{ pointer: '/state', detail: 'unknown, expired or already used' }] })],
    ['a 422 for a code the provider refused', () => json(422, { type: 'about:blank', title: 'Validation failed', status: 422, errors: [{ pointer: '/code', detail: 'refused by the provider' }] })],
    ['a 403 identity.association_disabled', () => problem(403, 'identity.association_disabled')],
    ['a 409 with another code', () => problem(409, 'identity.something_else')],
  ])('says the connection was not completed and links to the card, with no retry, on %s', async (_what, answer) => {
    breakPath(storage, 'POST', answer);
    const router = open(callback);
    const alert = await screen.findByRole('alert');
    expect(router.state.location.pathname).toBe('/storage/callback');
    expect(alert.textContent?.trim().length).toBeGreaterThan(0);
    expect(alert).not.toHaveTextContent('особистий акаунт');
    expect(alert).not.toHaveTextContent('Не вдалося виконати запит');
    expect(screen.queryByRole('button', { name: retryName })).toBeNull();
    expect(within(alert).getByRole('link').getAttribute('href')).toBe(`/o/${lypky}/association`);
    await new Promise((resolve) => setTimeout(resolve, 50));
    expect(countOf(storage, 'POST')).toBe(1);
  });

  it('explains what happened and links back to the card, without its code and state', async () => {
    const router = open('/storage/callback');
    const alert = await screen.findByRole('alert');
    expect(alert.textContent?.trim().length).toBeGreaterThan(0);
    expect(countOf(storage, 'POST')).toBe(0);
    expect(within(alert).getByRole('link').getAttribute('href')).toBe(`/o/${lypky}/association`);
    expect(router.state.location.pathname).toBe('/storage/callback');
  });

  describe('without its sessionStorage key the association is unknown', () => {
    const second = '00000000-0000-7000-8000-000000000002';
    beforeEach(() => {
      sessionStorage.clear();
      rosters = [membership, { ...membership, association_id: second, association_name: 'ОСББ «Сонячний»' }];
    });

    it('explains what happened and offers the card of each association of the person, not the entry page', async () => {
      const router = open(callback);
      const alert = await screen.findByRole('alert');
      expect(alert.textContent?.trim().length).toBeGreaterThan(0);
      expect(countOf(storage, 'POST')).toBe(0);
      await waitFor(() => {
        const hrefs = screen.getAllByRole('link').map((l) => l.getAttribute('href'));
        expect(hrefs).toContain(`/o/${lypky}/association`);
        expect(hrefs).toContain(`/o/${second}/association`);
      });
      expect(screen.getAllByRole('link').map((l) => l.getAttribute('href'))).not.toContain('/');
      expect(router.state.location.pathname).toBe('/storage/callback');
    });

    it('offers the card of the one association it has', async () => {
      rosters = [membership];
      open(callback);
      await screen.findByRole('alert');
      await waitFor(() => expect(screen.getAllByRole('link').map((l) => l.getAttribute('href'))).toContain(`/o/${lypky}/association`));
      expect(screen.getAllByRole('link').map((l) => l.getAttribute('href'))).not.toContain('/');
    });
  });
});
