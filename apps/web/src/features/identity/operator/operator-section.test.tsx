import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createMemoryHistory, createRouter, RouterProvider } from '@tanstack/react-router';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { getStartSignInUrl } from '@/api/gen/endpoints/identity/identity';
import { routeTree } from '@/routeTree.gen';

// The оператор's section through the real route tree: the layout, the index, the navigation and
// the one access guard (AC-16, AC-16b, AC-21). The server stays the rule: the fake answers 403
// `identity.not_operator` on every operator path for anyone it does not know as the оператор.

const appId = '00000000-0000-7000-8000-0000000000a1';
const principal = { account_id: '00000000-0000-7000-8000-00000000a001', email: 'user-1@example.test' };

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
};
const association = {
  id: '00000000-0000-7000-8000-000000000001',
  name: 'ОСББ «Липки»',
  edrpou: '12345678',
  state: 'connected',
  connected_by_hand: false,
  created_at: '2026-01-01T00:00:00Z',
};

type Who = 'operator' | 'person' | 'stale-operator' | 'expired';
let who: Who;
let calls: string[];

const json = (status: number, body: unknown) =>
  new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': status >= 400 ? 'application/problem+json' : 'application/json' },
  });
const problem = (status: number, code: string) => json(status, { type: 'about:blank', title: 'x', status, code });

function fakeApi(input: RequestInfo | URL) {
  const path = new URL(String(input), 'http://localhost').pathname;
  calls.push(path);
  if (path.endsWith('/identity/session')) {
    if (who === 'expired') return Promise.resolve(problem(401, 'identity.session_required'));
    return Promise.resolve(json(200, { ...principal, is_operator: who === 'operator' || who === 'stale-operator' }));
  }
  if (path.endsWith('/identity/memberships')) return Promise.resolve(json(200, { items: [], next_cursor: null }));
  if (path.includes('/identity/operator/')) {
    if (who === 'expired') return Promise.resolve(problem(401, 'identity.session_required'));
    if (who !== 'operator') return Promise.resolve(problem(403, 'identity.not_operator'));
    if (path.endsWith('/identity/operator/applications')) {
      return Promise.resolve(json(200, { items: [application], next_cursor: null }));
    }
    if (path.endsWith('/identity/operator/associations')) {
      return Promise.resolve(json(200, { items: [association], next_cursor: null }));
    }
  }
  return Promise.resolve(json(404, {}));
}

async function open(initial: string) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  const router = createRouter({ routeTree, context: { queryClient }, history: createMemoryHistory({ initialEntries: [initial] }) });
  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
  return router;
}

beforeEach(() => {
  who = 'operator';
  calls = [];
  vi.stubGlobal('fetch', vi.fn(fakeApi));
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe('the оператор\'s section (AC-16)', () => {
  it('opens the queue from «Оператор» on the picker', async () => {
    const router = await open('/');
    fireEvent.click(await screen.findByRole('link', { name: 'Оператор' }));
    expect(await screen.findByRole('heading', { name: 'Заявки та запити' })).toBeInTheDocument();
    expect(await screen.findByRole('link', { name: 'ОСББ «Тестовий Будинок»' })).toBeInTheDocument();
    expect(router.state.location.pathname).toBe('/operator/applications');
  });

  it('sends the section address itself to the queue', async () => {
    const router = await open('/operator');
    expect(await screen.findByRole('heading', { name: 'Заявки та запити' })).toBeInTheDocument();
    expect(router.state.location.pathname).toBe('/operator/applications');
  });

  it('links the queue and the associations list to each other (AC-21)', async () => {
    const router = await open('/operator/applications');
    expect(await screen.findByRole('heading', { name: 'Заявки та запити' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Заявки' })).toHaveAttribute('href', '/operator/applications');

    fireEvent.click(screen.getByRole('link', { name: 'ОСББ' }));
    expect(await screen.findByRole('heading', { name: 'ОСББ' })).toBeInTheDocument();
    expect(await screen.findByText('ОСББ «Липки»')).toBeInTheDocument();
    expect(router.state.location.pathname).toBe('/operator/associations');

    fireEvent.click(screen.getByRole('link', { name: 'Заявки' }));
    expect(await screen.findByRole('heading', { name: 'Заявки та запити' })).toBeInTheDocument();
    expect(router.state.location.pathname).toBe('/operator/applications');
  });

  it('keeps the navigation on the associations list too', async () => {
    await open('/operator/associations');
    expect(await screen.findByRole('heading', { name: 'ОСББ' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Заявки' })).toHaveAttribute('href', '/operator/applications');
    expect(screen.getByRole('link', { name: 'ОСББ' })).toHaveAttribute('href', '/operator/associations');
  });
});

describe('one access guard for every operator page (AC-16, AC-16b)', () => {
  const pages = ['/operator', '/operator/applications/', '/operator/associations', `/operator/applications/${appId}`];

  it.each(pages)('sends a person who is not the оператор from %s to /no-access', async (page) => {
    who = 'person';
    const router = await open(page);
    expect(await screen.findByRole('heading', { name: 'Немає доступу' })).toBeInTheDocument();
    expect(router.state.location.pathname).toBe('/no-access');
    expect(screen.queryByText(/Тестовий Будинок|Липки/)).toBeNull();
    expect(screen.queryByRole('link', { name: 'Заявки' })).toBeNull();
  });

  it.each(pages)('lets the server decide: a 403 not_operator on %s sends even a session marked оператор to /no-access', async (page) => {
    who = 'stale-operator';
    const router = await open(page);
    expect(await screen.findByRole('heading', { name: 'Немає доступу' })).toBeInTheDocument();
    expect(router.state.location.pathname).toBe('/no-access');
  });

  it.each(pages)('sends an expired session from %s to /sign-in', async (page) => {
    who = 'expired';
    const router = await open(page);
    expect(await screen.findByRole('link', { name: 'Увійти' })).toHaveAttribute('href', getStartSignInUrl());
    expect(router.state.location.pathname).toBe('/sign-in');
  });
});

describe('every page of the оператор\'s queue (AC-04c)', () => {
  it('offers «Показати ще» until next_cursor is null and appends the next page', async () => {
    const second = { ...application, id: '00000000-0000-7000-8000-0000000000a2', association_name: 'ОСББ «Друга Сторінка»' };
    const urls: string[] = [];
    vi.stubGlobal(
      'fetch',
      vi.fn((input: RequestInfo | URL) => {
        const u = new URL(String(input), 'http://localhost');
        if (u.pathname.endsWith('/identity/operator/applications')) {
          urls.push(u.search);
          return Promise.resolve(
            u.searchParams.get('cursor') === 'page-2'
              ? json(200, { items: [second], next_cursor: null })
              : json(200, { items: [application], next_cursor: 'page-2' }),
          );
        }
        return fakeApi(input);
      }),
    );
    await open('/operator/applications');
    expect(await screen.findByRole('link', { name: 'ОСББ «Тестовий Будинок»' })).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: 'ОСББ «Друга Сторінка»' })).toBeNull();

    fireEvent.click(await screen.findByRole('button', { name: 'Показати ще' }));

    expect(await screen.findByRole('link', { name: 'ОСББ «Друга Сторінка»' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'ОСББ «Тестовий Будинок»' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Показати ще' })).toBeNull();
    expect(urls.at(-1)).toContain('cursor=page-2');
  });
});
