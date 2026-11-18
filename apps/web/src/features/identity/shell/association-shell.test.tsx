import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createMemoryHistory, createRootRoute, createRouter, RouterProvider } from '@tanstack/react-router';
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import type { ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { getGetAssociationCardQueryKey } from '@/api/gen/endpoints/identity/identity';
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
let client: QueryClient;
// The next_cursor of the first page of the справи list: a full page of 20 has one, the last page has none.
let casesNextCursor: string | null;

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
  // The справи list and the roster are read from the API; a вимкнене ОСББ refuses them like every scoped read (ADR-0007).
  const governance = path.match(/\/governance\/associations\/([^/]+)\/(cases|board-members)$/);
  if (governance) {
    const id = governance[1] ?? '';
    if (disabled.has(id)) {
      return Promise.resolve(json(403, { type: 'about:blank', title: 'Forbidden', status: 403, detail: reason, code: 'identity.association_disabled' }));
    }
    if (governance[2] === 'board-members') {
      return Promise.resolve(
        json(200, {
          items: [
            {
              id: '00000000-0000-7000-8000-00000000b001',
              full_name: 'Тестовий Голова Правління',
              sign_in_email: 'head@example.test',
              term_start: '2026-11-11',
              term_end: null,
              revised_by_member_id: null,
              revised_at: '2026-11-11T08:05:00Z',
              position: 'head',
              decision_id: null,
            },
          ],
          next_cursor: null,
        }),
      );
    }
    return Promise.resolve(
      json(200, { items: [{ id: caseId, title: caseTitle, created_by_member_id: principal.account_id, created_at: '2026-10-20T09:00:00Z' }], next_cursor: casesNextCursor }),
    );
  }
  return Promise.resolve(json(404, {}));
}

function open(initial: string) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  client = queryClient;
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

// The layout owns the notice (AC-21, AC-25): exactly one alert «Доступ призупинено», and the page under it
// adds none of its own. The e2e specs select it with getByText in strict mode, so one element only. The
// short wait lets a second notice from the page (its own read of the same refusal) arrive before counting.
const suspended = async () => {
  await screen.findAllByRole('alert');
  await waitFor(() => expect(screen.queryAllByRole('alert').some((a) => a.textContent?.includes('Доступ призупинено'))).toBe(true));
  await new Promise((resolve) => setTimeout(resolve, 100));
  const notices = screen.queryAllByRole('alert').filter((a) => a.textContent?.includes('Доступ призупинено'));
  expect(notices, 'exactly one alert «Доступ призупинено»').toHaveLength(1);
  expect(screen.queryAllByText(/Доступ призупинено/), 'exactly one element with «Доступ призупинено»').toHaveLength(1);
  expect(notices[0]).toHaveTextContent(reason);
  return notices[0];
};

beforeEach(() => {
  rosters = [lypky, sonyachnyi];
  disabled = new Set();
  casesNextCursor = null;
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

  it.each([
    ['one roster', [lypky]],
    ['several rosters', [lypky, sonyachnyi]],
  ])('links «Мої заявки» and «Подати заявку на підключення» in the sidebar for a person on %s (AC-01, AC-05, AC-06, AC-14)', async (_label, roster) => {
    rosters = roster;
    open(`/o/${lypky}/cases`);
    const nav = await screen.findByRole('complementary');
    expect((await within(nav).findByRole('link', { name: 'Мої заявки' })).getAttribute('href')).toBe('/applications');
    expect(within(nav).getByRole('link', { name: 'Подати заявку на підключення' }).getAttribute('href')).toBe('/applications/new?kind=connection');
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
    ['the roster', `/o/${lypky}/roster`],
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
  const pages = [
    ['the cases page', (id: string) => `/o/${id}/cases`],
    ['the protocol page', (id: string) => `/o/${id}/cases/c-101/protocol`],
    ['the roster page', (id: string) => `/o/${id}/roster`],
  ] as const;

  describe.each(pages)('%s', (_page, pathOf) => {
    it.each([
      ['Липки', lypky],
      ['Сонячний', sonyachnyi],
    ])('names «%s» in the top bar and no fixed association', async (_name, id) => {
      open(pathOf(id));
      const main = await screen.findByRole('main');
      expect(await within(main).findByText(names[id] ?? '')).toBeInTheDocument();
      expect(main).not.toHaveTextContent('ОСББ 12');
    });
  });

  it('has dropped the fixed short name from the strings', () => {
    expect('osbbShort' in uk.app).toBe(false);
  });

  it('has dropped the strings nothing reads', () => {
    expect('osbb' in uk.app).toBe(false);
    expect('notOperator' in uk.operator).toBe(false);
  });
});

describe('the справи count in the sidebar (AC-04d)', () => {
  const casesLink = async () => within(await screen.findByRole('complementary')).findByRole('link', { name: /Справи/ });

  it('shows the number of справи when the list is one page', async () => {
    casesNextCursor = null;
    open(`/o/${lypky}/cases`);
    await waitFor(async () => expect((await casesLink()).textContent?.trim()).toMatch(/^Справи\s*1$/));
  });

  it('shows no number when the list has a next page: the first page would be a wrong count', async () => {
    casesNextCursor = 'opaque-cursor';
    open(`/o/${lypky}/cases`);
    const link = await casesLink();
    // Wait until the list itself arrived, so the absence of a number is not just a slow read.
    expect(await screen.findByRole('link', { name: /Гідроізоляція покрівлі/ })).toBeInTheDocument();
    expect(link.textContent?.trim()).toBe('Справи');
  });
});

describe('the card is read through the generated hook and key (T78, T68)', () => {
  const keysInCache = () => client.getQueryCache().findAll().map((q) => q.queryKey);
  const handMade = () => keysInCache().filter((k) => k[0] === 'association-card');

  it('shares one generated cache entry between the card page and its layout', async () => {
    open(`/o/${lypky}/association`);
    expect(await screen.findByRole('heading', { level: 1, name: 'ОСББ «Липки»' })).toBeInTheDocument();
    expect(client.getQueryData(getGetAssociationCardQueryKey(lypky))).toBeDefined();
    expect(handMade()).toEqual([]);
  });

  it('reads the card of the roster page under the generated key', async () => {
    open(`/o/${lypky}/roster`);
    expect(await screen.findByRole('heading', { level: 1, name: 'Правління' })).toBeInTheDocument();
    await waitFor(() => expect(client.getQueryData(getGetAssociationCardQueryKey(lypky))).toBeDefined());
    expect(handMade()).toEqual([]);
  });

  it('reads the reason for the suspended notice under the generated key', async () => {
    disabled.add(lypky);
    open(`/o/${lypky}/association`);
    await screen.findAllByRole('alert');
    await waitFor(() => expect(client.getQueryState(getGetAssociationCardQueryKey(lypky))?.status).toBe('error'));
    expect(client.getQueryState(getGetAssociationCardQueryKey(lypky))?.status).toBe('error');
    expect(handMade()).toEqual([]);
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

describe('a person on more than one page of memberships (AC-04d)', () => {
  it('opens the twenty-first association: the layout reads every page of memberships', async () => {
    const ids = Array.from({ length: 21 }, (_, i) => `00000000-0000-7000-8000-0000000d${String(i).padStart(4, '0')}`);
    for (const id of ids) names[id] = `ОСББ ${id.slice(-4)}`;
    const last = ids[20] ?? '';
    const fake = vi.fn((input: RequestInfo | URL) => {
      const u = new URL(String(input), 'http://localhost');
      if (u.pathname.endsWith('/identity/memberships')) {
        const second = u.searchParams.get('cursor') === 'm-2';
        return Promise.resolve(json(200, { items: (second ? ids.slice(20) : ids.slice(0, 20)).map(membershipOf), next_cursor: second ? null : 'm-2' }));
      }
      return fakeApi(input);
    });
    vi.stubGlobal('fetch', fake);
    const router = open(`/o/${last}/cases`);
    expect((await screen.findAllByText(names[last] ?? '')).length).toBeGreaterThan(0);
    expect(router.state.location.pathname).toBe(`/o/${last}/cases`);
  });
});
