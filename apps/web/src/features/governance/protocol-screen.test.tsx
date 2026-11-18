import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createMemoryHistory, createRouter, RouterProvider } from '@tanstack/react-router';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { getGetProtocolQueryKey } from '@/api/gen/endpoints/governance/governance';
import { uk } from '@/i18n/uk';
import { routeTree } from '@/routeTree.gen';

import { protocols } from './mock-data';

// T120 (review S08, AC-25): the protocol screen reads the draft and both writes (a slot save, a lifecycle
// transition) put the new draft under that same cache entry, so the screen shows it without a reload.
// The screen still runs on the in-memory governance mock; only the memberships come over the wire.

const lypky = '00000000-0000-7000-8000-000000000001';
const principal = { account_id: '00000000-0000-7000-8000-00000000a001', email: 'user-1@example.test', is_operator: false };
const membership = {
  association_id: lypky,
  association_name: 'ОСББ «Липки»',
  association_state: 'connected',
  board_member_id: '00000000-0000-7000-8000-00000000b001',
  position: 'member',
};

const json = (body: unknown) => new Response(JSON.stringify(body), { status: 200, headers: { 'Content-Type': 'application/json' } });

const pristine = structuredClone(protocols);

beforeEach(() => {
  for (const key of Object.keys(protocols)) delete protocols[key];
  Object.assign(protocols, structuredClone(pristine));
  vi.stubGlobal(
    'fetch',
    vi.fn((input: RequestInfo | URL) => {
      const path = new URL(String(input), 'http://localhost').pathname;
      if (path.endsWith('/identity/session')) return Promise.resolve(json(principal));
      if (path.endsWith('/identity/memberships')) return Promise.resolve(json({ items: [membership], next_cursor: null }));
      return Promise.resolve(json({ items: [], next_cursor: null }));
    }),
  );
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

function openProtocol(caseId: string) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  const router = createRouter({
    routeTree,
    context: { queryClient: client },
    history: createMemoryHistory({ initialEntries: [`/o/${lypky}/cases/${caseId}/protocol`] }),
  });
  render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
  return client;
}

describe('the protocol screen shows the new draft after a write, without a reload', () => {
  it('a saved slot is no longer dirty', async () => {
    const client = openProtocol('c-101');
    const field = await screen.findByRole('textbox', { name: 'Питання порядку денного' }, { timeout: 3000 });
    fireEvent.change(field, { target: { value: 'Нова редакція питання порядку денного' } });
    fireEvent.click(await screen.findByRole('button', { name: 'Зберегти' }));
    // While saving the button reads «Збереження…», so wait on the cancel button: it is there only while the slot is dirty.
    await waitFor(() => expect(screen.queryByRole('button', { name: uk.protocol.cancel })).not.toBeInTheDocument(), { timeout: 3000 });
    expect(screen.queryByRole('button', { name: 'Зберегти' })).not.toBeInTheDocument();
    expect(client.getQueryData<{ slots: { id: string; value: string }[] }>(getGetProtocolQueryKey(lypky, 'c-101'))?.slots.find((x) => x.id === 's-agenda')?.value).toBe(
      'Нова редакція питання порядку денного',
    );
    expect(screen.getByRole('textbox', { name: 'Питання порядку денного' })).toHaveValue('Нова редакція питання порядку денного');
  });

  it('a transition changes the state and the buttons', async () => {
    const client = openProtocol('c-101');
    fireEvent.click(await screen.findByRole('button', { name: 'Подати на попереднє голосування' }, { timeout: 3000 }));
    expect(await screen.findByRole('button', { name: 'Повернути в чернетку' }, { timeout: 3000 })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Подати на попереднє голосування' })).not.toBeInTheDocument();
    expect(client.getQueryData<{ state: string }>(getGetProtocolQueryKey(lypky, 'c-101'))?.state).toBe('ready_for_vote');
  });
});
