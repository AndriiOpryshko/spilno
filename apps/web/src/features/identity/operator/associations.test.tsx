import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createMemoryHistory, createRootRoute, createRouter, RouterProvider } from '@tanstack/react-router';
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { OperatorAssociations } from './associations';

const lypkyId = '00000000-0000-7000-8000-000000000001';
const sonyachnyiId = '00000000-0000-7000-8000-000000000002';

type State = 'connected' | 'disabled';
type Row = { id: string; name: string; edrpou: string | null; state: State; connected_by_hand: boolean; created_at: string };

let rows: Row[];
let calls: { url: string; body: unknown }[];

const json = (status: number, body: unknown, type = 'application/json') =>
  new Response(JSON.stringify(body), { status, headers: { 'Content-Type': type } });
const problem = (status: number, code: string) =>
  json(status, { type: 'about:blank', title: 'x', status, code }, 'application/problem+json');

// A stateful stand-in for the identity API (the real operations may still answer 501).
function fakeApi(input: RequestInfo | URL, init?: RequestInit) {
  const url = new URL(String(input), 'http://localhost').pathname;
  const body = init?.body ? JSON.parse(String(init.body)) : undefined;
  if ((init?.method ?? 'GET') === 'GET' && url.endsWith('/identity/operator/associations')) {
    return Promise.resolve(json(200, { items: rows, next_cursor: null }));
  }
  const m = url.match(/\/identity\/operator\/associations\/([^/]+)\/(disabling|enabling)$/);
  if (!m) return Promise.resolve(json(404, {}));
  calls.push({ url, body });
  const row = rows.find((r) => r.id === m[1])!;
  const target: State = m[2] === 'disabling' ? 'disabled' : 'connected';
  if (row.state === target) return Promise.resolve(problem(409, 'identity.association_state_unchanged'));
  if (target === 'disabled' && !String(body?.reason ?? '').trim()) {
    return Promise.resolve(problem(422, 'identity.reason_required'));
  }
  row.state = target;
  return Promise.resolve(
    json(200, {
      association_id: row.id,
      state: target,
      reason: target === 'disabled' ? body.reason : null,
      changed_at: '2026-11-10T10:00:00Z',
    }),
  );
}

async function renderPage() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  const router = createRouter({
    routeTree: createRootRoute({
      component: () => (
        <QueryClientProvider client={client}>
          <OperatorAssociations />
        </QueryClientProvider>
      ),
    }),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  });
  render(<RouterProvider router={router} />);
  await screen.findByText(/Липки/);
}

const rowOf = (name: RegExp) => screen.getByText(name).closest('tr, li, article, [data-slot="card"], div')!.parentElement!;
const user = () => ({
  click: async (el: HTMLElement) => fireEvent.click(el),
  type: async (el: HTMLElement, text: string) => fireEvent.change(el, { target: { value: text } }),
});

beforeEach(() => {
  calls = [];
  rows = [
    { id: lypkyId, name: 'ОСББ «Липки»', edrpou: '12345678', state: 'connected', connected_by_hand: false, created_at: '2026-01-01T00:00:00Z' },
    { id: sonyachnyiId, name: 'ОСББ «Сонячний»', edrpou: null, state: 'disabled', connected_by_hand: true, created_at: '2026-01-02T00:00:00Z' },
  ];
  vi.stubGlobal('fetch', vi.fn(fakeApi));
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe('OperatorAssociations', () => {
  it('lists the name, the ЄДРПОУ and the state of each association', async () => {
    await renderPage();
    expect(screen.getByText(/Липки/)).toBeInTheDocument();
    expect(screen.getByText('12345678')).toBeInTheDocument();
    expect(screen.getByText(/Сонячний/)).toBeInTheDocument();
  });

  it('disables a connected association with a reason, and the badge flips to вимкнене (AC-21)', async () => {
    await renderPage();
    const u = user();
    const row = rowOf(/Липки/);
    expect(within(row).getByText(/підключене/i)).toBeInTheDocument();

    await u.click(within(row).getByRole('button', { name: /Вимкнути/ }));
    await u.type(within(row).getByRole('textbox'), 'спір щодо складу правління');
    await u.click(within(row).getAllByRole('button', { name: /Вимкнути/ }).at(-1)!);

    await waitFor(() => expect(within(rowOf(/Липки/)).getByText(/вимкнене/i)).toBeInTheDocument());
    expect(calls).toEqual([
      { url: expect.stringContaining(`/${lypkyId}/disabling`), body: { reason: 'спір щодо складу правління' } },
    ]);
  });

  it('refuses an empty reason with the members message and keeps the association підключене (AC-22)', async () => {
    await renderPage();
    const u = user();
    const row = rowOf(/Липки/);
    await u.click(within(row).getByRole('button', { name: /Вимкнути/ }));
    await u.click(within(row).getAllByRole('button', { name: /Вимкнути/ }).at(-1)!);

    expect(await screen.findByText('Вкажіть причину — її побачать члени правління')).toBeInTheDocument();
    expect(within(rowOf(/Липки/)).getByText(/підключене/i)).toBeInTheDocument();
    expect(rows[0]!.state).toBe('connected');
  });

  it('enables a disabled association and the badge flips back to підключене (AC-23)', async () => {
    await renderPage();
    const u = user();
    const row = rowOf(/Сонячний/);
    expect(within(row).getByText(/вимкнене/i)).toBeInTheDocument();

    await u.click(within(row).getByRole('button', { name: /Увімкнути/ }));

    await waitFor(() => expect(within(rowOf(/Сонячний/)).getByText(/підключене/i)).toBeInTheDocument());
    expect(calls).toEqual([{ url: expect.stringContaining(`/${sonyachnyiId}/enabling`), body: undefined }]);
  });

  it('shows «ОСББ уже в цьому стані» when the state already changed meanwhile (AC-24)', async () => {
    await renderPage();
    const u = user();
    // Another operator enabled it after this list was loaded.
    rows[1]!.state = 'connected';
    await u.click(within(rowOf(/Сонячний/)).getByRole('button', { name: /Увімкнути/ }));

    expect(await screen.findByText('ОСББ уже в цьому стані')).toBeInTheDocument();
  });
});
