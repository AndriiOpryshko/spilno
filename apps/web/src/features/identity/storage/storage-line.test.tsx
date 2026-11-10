import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createMemoryHistory, createRootRoute, createRoute, createRouter, Outlet, RouterProvider } from '@tanstack/react-router';
import { render, screen } from '@testing-library/react';
import type { ReactElement } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { StorageCallback, StorageLine } from './storage-line';

const associationId = '00000000-0000-7000-8000-000000000002';
const address = 'osbb-storage@example.test';

const json = (status: number, body: unknown, type = 'application/json') =>
  new Response(JSON.stringify(body), { status, headers: { 'Content-Type': type } });

const connected = { connected: true, google_account_email: address, connected_by_member_id: '00000000-0000-7000-8000-0000000000b1', connected_at: '2026-11-10T10:00:00Z' };
const notConnected = { connected: false, google_account_email: null, connected_by_member_id: null, connected_at: null };

// The core API may still answer 501: every test runs against a stubbed fetch.
function stubApi(handler: (url: string, init?: RequestInit) => Response) {
  const fetchMock = vi.fn(async (url: string, init?: RequestInit) => handler(String(url), init));
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}

function renderRoutes(initial: string, callback: () => ReactElement) {
  const root = createRootRoute({
    component: () => (
      <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } })}>
        <Outlet />
      </QueryClientProvider>
    ),
  });
  const card = createRoute({ getParentRoute: () => root, path: '/o/$associationId/association', component: () => <StorageLine associationId={associationId} /> });
  const cb = createRoute({ getParentRoute: () => root, path: '/storage/callback', component: callback });
  const router = createRouter({ routeTree: root.addChildren([card, cb]), history: createMemoryHistory({ initialEntries: [initial] }) });
  render(<RouterProvider router={router} />);
  return router;
}

afterEach(() => vi.unstubAllGlobals());

describe('StorageLine (AC-27, AC-09)', () => {
  it('shows the account address when the storage is connected', async () => {
    stubApi(() => json(200, connected));
    renderRoutes(`/o/${associationId}/association`, () => <></>);
    expect(await screen.findByText(`Сховище документів: ${address}`)).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: 'Підключити сховище' })).not.toBeInTheDocument();
  });

  it('says the storage is not connected and offers to connect it through the start endpoint (AC-09)', async () => {
    stubApi(() => json(200, notConnected));
    renderRoutes(`/o/${associationId}/association`, () => <></>);
    expect(await screen.findByText('сховище не підключене')).toBeInTheDocument();
    const connect = screen.getByRole('link', { name: 'Підключити сховище' });
    expect(connect.getAttribute('href')).toBe(`/api/identity/associations/${associationId}/storage-connection/authorization`);
  });
});

describe('StorageCallback (AC-27, AC-28b)', () => {
  it('posts the code and state, returns to the card and shows the account address', async () => {
    let done = false;
    const fetchMock = stubApi((_url, init) => {
      if (init?.method === 'POST') {
        done = true;
        return json(201, connected);
      }
      return json(200, done ? connected : notConnected);
    });
    const router = renderRoutes('/storage/callback', () => <StorageCallback associationId={associationId} code="auth-code" state="auth-state" />);
    expect(await screen.findByText(`Сховище документів: ${address}`)).toBeInTheDocument();
    expect(router.state.location.pathname).toBe(`/o/${associationId}/association`);
    const post = (fetchMock.mock.calls as unknown as [string, RequestInit][]).find(([, init]) => init?.method === 'POST');
    expect(post?.[0]).toBe(`/api/identity/associations/${associationId}/storage-connection`);
    expect(JSON.parse(post![1].body as string)).toEqual({ code: 'auth-code', state: 'auth-state' });
  });

  it('shows the personal-account refusal and leaves the storage not connected', async () => {
    stubApi((_url, init) =>
      init?.method === 'POST'
        ? json(409, { type: 'about:blank', title: 'x', status: 409, code: 'identity.storage_account_personal', detail: 'personal' }, 'application/problem+json')
        : json(200, notConnected),
    );
    renderRoutes('/storage/callback', () => <StorageCallback associationId={associationId} code="auth-code" state="auth-state" />);
    expect(await screen.findByText('Це особистий акаунт члена правління — підключіть акаунт ОСББ')).toBeInTheDocument();
    expect(await screen.findByText('сховище не підключене')).toBeInTheDocument();
    expect(screen.queryByText(/Сховище документів:/)).not.toBeInTheDocument();
  });
});
