import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createMemoryHistory, createRouter, RouterProvider } from '@tanstack/react-router';
import { cleanup, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { routeTree } from '@/routeTree.gen';

// The roster page is reachable from the association sidebar and shows the roster read through the
// real route tree (AC-09: the approved applicant opens the roster of the new association).

const lypky = '00000000-0000-7000-8000-000000000001';
const json = (status: number, body: unknown) => new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });

const membership = {
  association_id: lypky,
  association_name: 'ОСББ «Липки»',
  association_state: 'connected',
  board_member_id: '00000000-0000-7000-8000-00000000b001',
  position: 'head',
};
const head = {
  id: '00000000-0000-7000-8000-00000000b001',
  full_name: 'Тестовий Голова Правління',
  sign_in_email: 'head@example.test',
  term_start: '2026-11-11',
  term_end: null,
  revised_by_member_id: null,
  revised_at: '2026-11-11T08:05:00Z',
  position: 'head',
  decision_id: null,
};

function fakeApi(input: RequestInfo | URL): Response {
  const path = new URL(String(input), 'http://localhost').pathname.replace(/^\/api/, '');
  if (path === '/identity/session') return json(200, { account_id: '00000000-0000-7000-8000-00000000a001', email: 'head@example.test', is_operator: false });
  if (path === '/identity/memberships') return json(200, { items: [membership], next_cursor: null });
  if (path === `/governance/associations/${lypky}/board-members`) return json(200, { items: [head], next_cursor: null });
  if (path === `/identity/associations/${lypky}`) return json(200, { id: lypky, name: 'ОСББ «Липки»', edrpou: '12345678', address: 'м. Київ', state: 'connected', head_full_name: head.full_name, connected_by_hand: false, history: [] });
  if (path === '/governance/cases') return json(200, { items: [], next_cursor: null });
  return json(404, {});
}

beforeEach(() => vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL) => fakeApi(input))));
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe('the roster route', () => {
  it('opens at /o/<id>/roster with the roster of the association and a sidebar link to it', async () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
    const router = createRouter({ routeTree, context: { queryClient }, history: createMemoryHistory({ initialEntries: [`/o/${lypky}/roster`] }) });
    render(
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>,
    );
    expect(await screen.findByRole('heading', { level: 1, name: 'Правління' })).toBeInTheDocument();
    expect(await screen.findByText(head.full_name)).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Правління' }).getAttribute('href')).toBe(`/o/${lypky}/roster`);
  });
});
