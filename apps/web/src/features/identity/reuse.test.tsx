import { readFileSync } from 'node:fs';
import path from 'node:path';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createMemoryHistory, createRootRoute, createRoute, createRouter, Outlet, RouterProvider } from '@tanstack/react-router';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { getGetSessionQueryKey, getGetStorageConnectionQueryKey, getListMembershipsQueryKey } from '@/api/gen/endpoints/identity/identity';
import type { Application, AssociationCard as Card } from '@/api/gen/model';
import { routeTree } from '@/routeTree.gen';

import { AssociationCard } from './association/association-card';
import { StatusLine } from './applications/status-line';
import { StorageCallback, StorageLine } from './storage/storage-line';

// T77 (review W5, AC-13, AC-21): the identity screens reuse InlineAlert, the generated hooks with their
// keys and the router's Link, and the label and dialog primitives are the shadcn ones or recorded
// deviations. The first block is the Definition of Done's grep, run over the sources; the second block
// is what a person sees and can do.

const sources = import.meta.glob(
  ['/src/features/**/*.{ts,tsx}', '/src/routes/**/*.tsx', '!/**/*.test.{ts,tsx}', '!/**/*.stories.tsx', '!/src/routeTree.gen.ts'],
  { query: '?raw', import: 'default', eager: true },
) as Record<string, string>;

const offenders = (pattern: RegExp) =>
  Object.entries(sources)
    .filter(([, text]) => pattern.test(text))
    .map(([file]) => file)
    .sort();

describe('the screens under src/features follow the reuse conventions (W5)', () => {
  it('reads the sources it checks', () => {
    expect(Object.keys(sources).length).toBeGreaterThan(20);
    expect(Object.keys(sources).some((f) => f.includes('/routes/'))).toBe(true);
  });

  it('renders no ad-hoc alert paragraph: every role="alert" comes from InlineAlert', () => {
    // Scoped to identity and the routes: InlineAlert itself and the system health card own their role="alert".
    expect(offenders(/role="alert"/).filter((f) => !f.includes('/features/governance/') && !f.includes('/features/system/'))).toEqual([]);
  });

  it('builds no query key by hand: the generated key helpers name the cache entries', () => {
    expect(offenders(/queryKey:\s*\['/)).toEqual([]);
  });

  it('has no <a href> for in-app navigation: only the generated API URL helpers (redirects, downloads) stay anchors', () => {
    const anchorsWithoutApiUrl = Object.entries(sources).flatMap(([file, text]) =>
      [...text.matchAll(/<a\b[^]*?\shref=(\{[^}]*\}|"[^"]*")/g)]
        .map((m) => m[1] ?? '')
        .filter((href) => !/^\{get\w+Url\(/.test(href))
        .map((href) => `${file}: ${href}`),
    );
    expect(anchorsWithoutApiUrl).toEqual([]);
  });

  it('has no in-app address in a link description either', () => {
    expect(offenders(/href:\s*['"`]\//)).toEqual([]);
  });
});

describe('label and dialog are the shadcn primitives or recorded deviations', () => {
  const read = (relative: string) => readFileSync(path.resolve(import.meta.dirname, relative), 'utf8');
  const files: Record<string, string> = {
    'label.tsx': read('../../components/ui/label.tsx'),
    'dialog.tsx': read('../../components/ui/dialog.tsx'),
  };
  const file = (name: string) => files[name] ?? '';
  const doc = read('../../../../../docs/design/design-system.md');

  // A deviation is a line of the design system document that names the file.
  const listed = (name: string) => doc.split('\n').some((line) => line.includes(name) && /deviat|відхил|trimmed|copy/i.test(line));

  it('can read the files it checks', () => {
    expect(file('label.tsx')).toContain('Label');
    expect(file('dialog.tsx')).toContain('Dialog');
    expect(doc).toContain('Usage rules');
  });

  it('label is the Radix label of shadcn, or listed as a deviation', () => {
    const shadcn = /\bLabel as LabelPrimitive\b/.test(file('label.tsx')) && /from ["']radix-ui["']/.test(file('label.tsx'));
    expect(shadcn || listed('label.tsx'), 'label.tsx is a plain <label> and no deviation line names it in docs/design/design-system.md').toBe(true);
  });

  it('dialog has the whole shadcn set, or is listed as a deviation', () => {
    const text = file('dialog.tsx');
    const shadcn = ['DialogHeader', 'DialogFooter', 'DialogDescription', 'DialogOverlay', 'DialogPortal'].every((part) => text.includes(part));
    expect(shadcn || listed('dialog.tsx'), 'dialog.tsx is a trimmed copy and no deviation line names it in docs/design/design-system.md').toBe(true);
  });
});

const lypky = '00000000-0000-7000-8000-000000000001';
const sonyachnyi = '00000000-0000-7000-8000-000000000002';
const principal = { account_id: '00000000-0000-7000-8000-00000000a001', email: 'user-1@example.test', is_operator: false };
const notConnected = { connected: false, google_account_email: null, connected_by_member_id: null, connected_at: null };

const json = (status: number, body: unknown) =>
  new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': status >= 400 ? 'application/problem+json' : 'application/json' },
  });

const membership = (id: string, name: string) => ({
  association_id: id,
  association_name: name,
  association_state: 'connected',
  board_member_id: '00000000-0000-7000-8000-00000000b001',
  position: 'member',
});

let isOperator = false;
let application: Application;

function fakeApi(input: RequestInfo | URL) {
  const path = new URL(String(input), 'http://localhost').pathname;
  if (path.endsWith('/identity/session')) return Promise.resolve(json(200, { ...principal, is_operator: isOperator }));
  if (path.endsWith('/identity/memberships')) {
    return Promise.resolve(json(200, { items: [membership(lypky, 'ОСББ «Липки»'), membership(sonyachnyi, 'ОСББ «Сонячний»')], next_cursor: null }));
  }
  const card = path.match(/\/identity\/associations\/([^/]+)$/);
  if (card) {
    return Promise.resolve(
      json(200, { id: card[1], name: 'ОСББ «Липки»', edrpou: '12345678', address: 'м. Київ, вул. Тестова, 1', state: 'connected', head_full_name: 'Test User', connected_by_hand: false, history: [] }),
    );
  }
  if (path.endsWith('/storage-connection')) {
    return Promise.resolve(json(200, notConnected));
  }
  if (path.endsWith('/identity/applications')) return Promise.resolve(json(200, { items: [application], next_cursor: null }));
  return Promise.resolve(json(200, { items: [], next_cursor: null }));
}

let client: QueryClient;

function openRoutes(initial: string) {
  client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  const router = createRouter({ routeTree, context: { queryClient: client }, history: createMemoryHistory({ initialEntries: [initial] }) });
  render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
  return router;
}

// A component under test inside a router that accepts any address, so a click lands somewhere.
function renderInRouter(ui: ReactNode) {
  const root = createRootRoute({ component: () => <Outlet /> });
  const home = createRoute({ getParentRoute: () => root, path: '/', component: () => <>{ui}</> });
  const anywhere = createRoute({ getParentRoute: () => root, path: '$', component: () => <p>elsewhere</p> });
  const router = createRouter({ routeTree: root.addChildren([home, anywhere]), history: createMemoryHistory({ initialEntries: ['/'] }) });
  render(<RouterProvider router={router} />);
  return router;
}

beforeEach(() => {
  isOperator = false;
  application = {
    id: '00000000-0000-7000-8000-0000000000a1',
    kind: 'connection',
    edrpou: '12345678',
    association_id: lypky,
    association_name: 'ОСББ «Липки»',
    state: 'approved',
    filed_at: '2026-11-10',
    rejection_reason: null,
    created_at: '2026-11-10T10:00:00Z',
  };
  vi.stubGlobal('fetch', vi.fn(fakeApi));
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe('the session, the roster and the storage line sit under the generated keys', () => {
  const handMade = () => client.getQueryCache().findAll().map((q) => q.queryKey).filter((k) => ['session', 'memberships', 'storage-connection'].includes(String(k[0])));

  it('reads the entry page through the session and memberships keys', async () => {
    openRoutes('/');
    await waitFor(() => expect(client.getQueryData(getListMembershipsQueryKey())).toBeDefined());
    expect(client.getQueryData(getGetSessionQueryKey())).toBeDefined();
    expect(handMade()).toEqual([]);
  });

  // The layout is identity's; the child is too (the association page), so the check does not depend on a
  // governance screen, whose own reads are outside this task.
  it('reads the association layout through the same keys', async () => {
    openRoutes(`/o/${lypky}/association`);
    await waitFor(() => expect(client.getQueryData(getListMembershipsQueryKey())).toBeDefined());
    expect(client.getQueryData(getGetSessionQueryKey())).toBeDefined();
    expect(handMade()).toEqual([]);
  });

  it('reads the оператор section through the session key', async () => {
    isOperator = true;
    openRoutes('/operator/applications');
    await waitFor(() => expect(client.getQueryData(getGetSessionQueryKey())).toBeDefined());
    expect(handMade()).toEqual([]);
  });

  it('reads the storage line under the generated key', async () => {
    openRoutes(`/o/${lypky}/association`);
    expect(await screen.findByText('сховище не підключене')).toBeInTheDocument();
    expect(client.getQueryData(getGetStorageConnectionQueryKey(lypky))).toBeDefined();
    expect(handMade()).toEqual([]);
  });

  it('keeps the storage line under the generated key after the consent finishes', async () => {
    client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
    const root = createRootRoute({
      component: () => (
        <QueryClientProvider client={client}>
          <Outlet />
        </QueryClientProvider>
      ),
    });
    const card = createRoute({ getParentRoute: () => root, path: '/o/$associationId/association', component: () => <StorageLine associationId={lypky} /> });
    const callback = createRoute({ getParentRoute: () => root, path: '/storage/callback', component: () => <StorageCallback associationId={lypky} code="auth-code" state="auth-state" /> });
    const router = createRouter({ routeTree: root.addChildren([card, callback]), history: createMemoryHistory({ initialEntries: ['/storage/callback'] }) });
    render(<RouterProvider router={router} />);
    await waitFor(() => expect(router.state.location.pathname).toBe(`/o/${lypky}/association`));
    expect(await screen.findByText('сховище не підключене')).toBeInTheDocument();
    expect(client.getQueryData(getGetStorageConnectionQueryKey(lypky))).toBeDefined();
    expect(handMade()).toEqual([]);
  });
});

describe('in-app links navigate through the router, not by a page load', () => {
  it('opens a заявка from the list in place', async () => {
    const router = openRoutes('/applications');
    const link = await screen.findByRole('link', { name: 'ОСББ «Липки»' });
    expect(link.getAttribute('href')).toBe(`/applications/${application.id}`);
    fireEvent.click(link);
    await waitFor(() => expect(router.state.location.pathname).toBe(`/applications/${application.id}`));
  });

  it('opens the association from an approved заявка in place', async () => {
    const router = renderInRouter(<StatusLine application={application} />);
    const link = await screen.findByRole('link', { name: 'Перейти до ОСББ ОСББ «Липки»' });
    expect(link.getAttribute('href')).toBe(`/o/${lypky}/association`);
    fireEvent.click(link);
    await waitFor(() => expect(router.state.location.pathname).toBe(`/o/${lypky}/association`));
  });

  it('starts a transfer request from the card of a hand-connected association in place', async () => {
    const card: Card = {
      id: lypky,
      name: 'ОСББ «Липки»',
      edrpou: '12345678',
      address: 'м. Київ, вул. Тестова, 1',
      state: 'connected',
      head_full_name: 'Test User',
      connected_by_hand: true,
      history: [],
    };
    const router = renderInRouter(<AssociationCard card={card} />);
    const link = await screen.findByRole('link', { name: 'Подати запит на передачу' });
    expect(link.getAttribute('href')).toBe(`/applications/new?kind=transfer&association_id=${lypky}`);
    fireEvent.click(link);
    await waitFor(() => expect(router.state.location.pathname).toBe('/applications/new'));
    expect(router.state.location.search).toMatchObject({ kind: 'transfer', association_id: lypky });
  });
});

describe('the refusal of a personal storage account is an InlineAlert', () => {
  it('shows the message with the alert glyph, like every other refusal', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn((_url: string, init?: RequestInit) =>
        Promise.resolve(
          init?.method === 'POST'
            ? json(409, { type: 'about:blank', title: 'x', status: 409, code: 'identity.storage_account_personal', detail: 'personal' })
            : json(200, notConnected),
        ),
      ),
    );
    client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
    const root = createRootRoute({
      component: () => (
        <QueryClientProvider client={client}>
          <StorageCallback associationId={lypky} code="auth-code" state="auth-state" />
        </QueryClientProvider>
      ),
    });
    render(<RouterProvider router={createRouter({ routeTree: root, history: createMemoryHistory({ initialEntries: ['/'] }) })} />);
    const alert = await screen.findByRole('alert');
    expect(alert).toHaveTextContent('Це особистий акаунт члена правління — підключіть акаунт ОСББ');
    expect(alert.querySelector('svg'), 'InlineAlert draws the status glyph inside its alert').not.toBeNull();
  });
});
