import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createMemoryHistory, createRouter, RouterProvider } from '@tanstack/react-router';
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { routeTree } from '@/routeTree.gen';

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
let rosters: unknown[];
let broken: { match: (path: string, method: string) => boolean; answer: () => Answer } | null;
let calls: { path: string; method: string }[];

function fakeApi(input: RequestInfo | URL, init?: RequestInit): Answer {
  const path = new URL(String(input), 'http://localhost').pathname.replace(/^\/api/, '');
  const method = init?.method ?? 'GET';
  calls.push({ path, method });
  if (broken?.match(path, method)) return broken.answer();
  if (path === '/identity/session') return json(200, principal);
  if (path === '/identity/memberships') return json(200, { items: rosters, next_cursor: null });
  if (path === '/identity/applications') return json(200, { items: [application], next_cursor: null });
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
  broken = null;
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
}

const pages: Page[] = [
  { name: 'the entry page: the session', url: '/', path: '/identity/session', settled: () => screen.findByRole('heading', { name: 'Справи' }) },
  { name: 'the entry page: the memberships', url: '/', path: '/identity/memberships', settled: () => screen.findByRole('heading', { name: 'Справи' }) },
  { name: 'the association layout: the session', url: `/o/${lypky}/cases`, path: '/identity/session', settled: () => screen.findByRole('heading', { name: 'Справи' }) },
  { name: 'the association layout: the memberships', url: `/o/${lypky}/cases`, path: '/identity/memberships', settled: () => screen.findByRole('heading', { name: 'Справи' }) },
  { name: 'the association card', url: `/o/${lypky}/association`, path: `/identity/associations/${lypky}`, settled: () => screen.findByRole('heading', { level: 1, name: 'ОСББ «Липки»' }) },
  { name: 'my applications', url: '/applications/', path: '/identity/applications', settled: () => screen.findByText('ОСББ «Тест»') },
  { name: 'an application', url: `/applications/${appId}`, path: `/identity/applications/${appId}`, settled: () => screen.findByRole('heading', { name: 'ОСББ «Тест»' }) },
];

describe.each(pages)('$name', ({ url, path, settled }) => {
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

  it.each([
    ['its sessionStorage key', () => sessionStorage.clear(), callback],
    ['its code and state', () => undefined, '/storage/callback'],
  ])('explains what happened and links back, without %s', async (_what, prepare, url) => {
    prepare();
    const router = open(url);
    const alert = await screen.findByRole('alert');
    expect(alert.textContent?.trim().length).toBeGreaterThan(0);
    expect(countOf(storage, 'POST')).toBe(0);
    // Without the key the association is unknown: the way back is the entry page or the card itself.
    const back = within(alert).getByRole('link');
    expect(back.getAttribute('href')).toMatch(/^\/$|^\/o\/[^/]+\/association$/);
    expect(router.state.location.pathname).toBe('/storage/callback');
  });
});
