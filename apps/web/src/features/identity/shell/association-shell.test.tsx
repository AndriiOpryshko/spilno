import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createMemoryHistory, createRootRoute, createRouter, RouterProvider } from '@tanstack/react-router';
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import type { ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { Sidebar } from '@/components/app-shell';
import { uk } from '@/i18n/uk';
import { routeTree } from '@/routeTree.gen';

import { AssociationPicker } from './association-picker';

// The association shell through the real route tree: the card in the navigation, the suspended notice
// in the layout, the association's own name on every page, the mobile case card (AC-04d, AC-21, AC-25,
// AC-27, AC-29). The fake answers like the server: a вимкнене ОСББ refuses every association-scoped
// read with 403 `identity.association_disabled` and the reason in `detail` (ADR-0007), and its
// membership carries `association_state: disabled`.

const lypky = '00000000-0000-7000-8000-000000000001';
const sonyachnyi = '00000000-0000-7000-8000-000000000002';
const names: Record<string, string> = { [lypky]: 'ОСББ «Липки»', [sonyachnyi]: 'ОСББ «Сонячний»' };
const reason = 'спір щодо складу правління';
const principal = { account_id: '00000000-0000-7000-8000-00000000a001', email: 'user-1@example.test', is_operator: false };

const caseId = '00000000-0000-7000-8000-000000000c01';
const caseTitle = 'Гідроізоляція покрівлі, будинок 1';

let rosters: string[];
let disabled: Set<string>;

const json = (status: number, body: unknown) =>
  new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': status >= 400 ? 'application/problem+json' : 'application/json' },
  });

const membershipOf = (id: string) => ({
  association_id: id,
  association_name: names[id] ?? '',
  association_state: disabled.has(id) ? 'disabled' : 'connected',
  board_member_id: '00000000-0000-7000-8000-00000000b001',
  position: 'member',
});

function fakeApi(input: RequestInfo | URL) {
  const path = new URL(String(input), 'http://localhost').pathname;
  if (path.endsWith('/identity/session')) return Promise.resolve(json(200, principal));
  if (path.endsWith('/identity/memberships')) return Promise.resolve(json(200, { items: rosters.map(membershipOf), next_cursor: null }));
  const scoped = path.match(/\/identity\/associations\/([^/]+)(\/.*)?$/);
  if (scoped) {
    const id = scoped[1] ?? '';
    if (disabled.has(id)) {
      return Promise.resolve(json(403, { type: 'about:blank', title: 'Forbidden', status: 403, detail: reason, code: 'identity.association_disabled' }));
    }
    if (scoped[2]?.startsWith('/storage-connection')) {
      return Promise.resolve(json(200, { connected: false, google_account_email: null, connected_by_member_id: null, connected_at: null }));
    }
    if (!scoped[2]) {
      return Promise.resolve(
        json(200, { id, name: names[id], edrpou: '12345678', address: 'м. Київ, вул. Тестова, 1', state: 'connected', head_full_name: 'Test User', connected_by_hand: false, history: [] }),
      );
    }
  }
  // The справи list is read from the API; a вимкнене ОСББ refuses it like every scoped read (ADR-0007).
  const governance = path.match(/\/governance\/associations\/([^/]+)\/cases$/);
  if (governance) {
    const id = governance[1] ?? '';
    if (disabled.has(id)) {
      return Promise.resolve(json(403, { type: 'about:blank', title: 'Forbidden', status: 403, detail: reason, code: 'identity.association_disabled' }));
    }
    return Promise.resolve(json(200, { items: [{ id: caseId, title: caseTitle, created_by_member_id: principal.account_id, created_at: '2026-10-20T09:00:00Z' }], next_cursor: null }));
  }
  return Promise.resolve(json(404, {}));
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

// Links need a router; the component under test is the root route's component.
async function renderInRouter(ui: ReactNode) {
  const router = createRouter({
    routeTree: createRootRoute({ component: () => <>{ui}</> }),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  });
  render(<RouterProvider router={router} />);
  await waitFor(() => expect(document.body.textContent).not.toBe(''));
}

const suspended = async () => {
  const alerts = await screen.findAllByRole('alert');
  const notice = alerts.find((a) => a.textContent?.includes('Доступ призупинено'));
  expect(notice, 'an alert «Доступ призупинено»').toBeDefined();
  expect(notice).toHaveTextContent(reason);
  return notice;
};

beforeEach(() => {
  rosters = [lypky, sonyachnyi];
  disabled = new Set();
  vi.stubGlobal('fetch', vi.fn(fakeApi));
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe('the association card in the navigation (AC-04d, AC-27, AC-29)', () => {
  it('links «ОСББ» in the sidebar to the card of its own association', async () => {
    await renderInRouter(<Sidebar associationId={sonyachnyi} associationName="ОСББ «Сонячний»" isOperator={false} />);
    const nav = screen.getByRole('complementary');
    expect(within(nav).getByRole('link', { name: 'ОСББ' }).getAttribute('href')).toBe(`/o/${sonyachnyi}/association`);
  });

  it('offers the card of each association next to its cases in the picker', async () => {
    const memberships = [membershipOf(lypky), membershipOf(sonyachnyi)] as never;
    await renderInRouter(<AssociationPicker memberships={memberships} isOperator={false} />);
    const hrefs = screen.getAllByRole('link').map((l) => l.getAttribute('href'));
    expect(hrefs).toContain(`/o/${lypky}/association`);
    expect(hrefs).toContain(`/o/${sonyachnyi}/association`);
    expect(hrefs).toContain(`/o/${lypky}/cases`);
    expect(hrefs).toContain(`/o/${sonyachnyi}/cases`);
  });

  it('opens the card from the sidebar of the association layout', async () => {
    const router = open(`/o/${lypky}/cases`);
    const nav = await screen.findByRole('complementary');
    fireEvent.click(await within(nav).findByRole('link', { name: 'ОСББ' }));
    expect(await screen.findByRole('heading', { level: 1, name: 'ОСББ «Липки»' })).toBeInTheDocument();
    expect(router.state.location.pathname).toBe(`/o/${lypky}/association`);
    expect(await screen.findByText(`Test User`)).toBeInTheDocument();
  });
});

describe('the suspended notice in the association layout (AC-21, AC-25)', () => {
  it('shows no notice on any page of an enabled association', async () => {
    open(`/o/${lypky}/cases`);
    expect(await screen.findByRole('heading', { name: 'Справи' })).toBeInTheDocument();
    expect(await within(await screen.findByRole('complementary')).findByRole('link', { name: 'ОСББ' })).toBeInTheDocument();
    expect(screen.queryByText(/Доступ призупинено/)).toBeNull();
  });

  it.each([
    ['the cases list', `/o/${lypky}/cases`],
    ['a protocol', `/o/${lypky}/cases/c-101/protocol`],
    ['the card', `/o/${lypky}/association`],
  ])('shows «Доступ призупинено» with the reason on %s of a вимкнене ОСББ', async (_page, path) => {
    disabled.add(lypky);
    open(path);
    await suspended();
  });

  it('keeps the notice to the вимкнене ОСББ: the other association of the same person has none (AC-04d)', async () => {
    disabled.add(lypky);
    open(`/o/${sonyachnyi}/cases`);
    expect(await screen.findByRole('heading', { name: 'Справи' })).toBeInTheDocument();
    expect(screen.queryByText(/Доступ призупинено/)).toBeNull();
  });

  it('shows the notice on the next action after the оператор disabled the association (AC-21)', async () => {
    const router = open(`/o/${lypky}/association`);
    expect(await screen.findByRole('heading', { level: 1, name: 'ОСББ «Липки»' })).toBeInTheDocument();
    expect(screen.queryByText(/Доступ призупинено/)).toBeNull();

    disabled.add(lypky);
    fireEvent.click(within(screen.getByRole('complementary')).getByRole('link', { name: /Справи/ }));
    await waitFor(() => expect(router.state.location.pathname).toBe(`/o/${lypky}/cases`));
    await suspended();
  });
});

describe('every page names its association (AC-04d)', () => {
  it.each([lypky, sonyachnyi])('names %s in the top bar of the cases page and no fixed association', async (id) => {
    open(`/o/${id}/cases`);
    const main = await screen.findByRole('main');
    expect(await within(main).findByText(names[id] ?? '')).toBeInTheDocument();
    expect(main).not.toHaveTextContent('ОСББ 12');
  });

  it('has dropped the fixed short name from the strings', () => {
    expect('osbbShort' in uk.app).toBe(false);
  });
});

describe('the справа link in the list (AC-04d)', () => {
  it('opens the case under the association, not the retired address', async () => {
    const router = open(`/o/${lypky}/cases`);
    const card = await screen.findByRole('link', { name: /Гідроізоляція покрівлі/ });
    expect(card.getAttribute('href')).toBe(`/o/${lypky}/cases/${caseId}`);
    fireEvent.click(card);
    await waitFor(() => expect(router.state.location.pathname).toBe(`/o/${lypky}/cases/${caseId}`));
  });

  it('carries the id of the association it is opened under', async () => {
    open(`/o/${sonyachnyi}/cases`);
    const card = await screen.findByRole('link', { name: /Гідроізоляція покрівлі/ });
    expect(card.getAttribute('href')).toBe(`/o/${sonyachnyi}/cases/${caseId}`);
  });
});

describe('the retired /cases routes keep redirecting', () => {
  it.each(['/cases', '/cases/c-101/protocol'])('sends %s through the entry into the only association', async (path) => {
    rosters = [lypky];
    const router = open(path);
    await waitFor(() => expect(router.state.location.pathname).toBe(`/o/${lypky}/cases`));
  });
});
