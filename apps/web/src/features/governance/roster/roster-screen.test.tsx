import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createMemoryHistory, createRootRoute, createRouter, RouterProvider } from '@tanstack/react-router';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { getDownloadConfirmationExtractUrl } from '@/api/gen/endpoints/identity/identity';
import { getCreateBoardMemberUrl, getListBoardMembersUrl, getUpdateBoardMemberUrl } from '@/api/gen/endpoints/governance/governance';
import type { AssociationCard, BoardMember } from '@/api/gen/model';

import { RosterScreen } from './roster-screen';

// The core API may still answer 501 for these operations: the screen runs against a stubbed fetch
// that behaves like the contract (roster of one association, the card with its decision history).

const associationId = '00000000-0000-7000-8000-000000000002';
const decisionId = '00000000-0000-7000-8000-00000000f002';
const extractId = '00000000-0000-7000-8000-00000000e001';
const previousId = '00000000-0000-7000-8000-00000000b001';
const currentId = '00000000-0000-7000-8000-00000000b002';
const memberId = '00000000-0000-7000-8000-00000000b003';

const previousHead: BoardMember = {
  id: previousId,
  full_name: 'Колишній Голова Тестовий',
  sign_in_email: 'previous@example.test',
  term_start: '2024-01-15',
  term_end: '2026-10-24',
  revised_by_member_id: null,
  revised_at: '2026-10-21T08:05:00Z',
  position: 'head',
  decision_id: decisionId,
};
const newHead: BoardMember = {
  id: currentId,
  full_name: 'Новий Голова Тестовий',
  sign_in_email: 'new-head@example.test',
  term_start: '2026-10-25',
  term_end: null,
  revised_by_member_id: null,
  revised_at: '2026-10-21T08:05:00Z',
  position: 'head',
  decision_id: decisionId,
};
const plainMember: BoardMember = {
  id: memberId,
  full_name: 'Звичайний Член Тестовий',
  sign_in_email: 'member@example.test',
  term_start: '2025-03-01',
  term_end: null,
  revised_by_member_id: null,
  revised_at: '2025-03-01T08:00:00Z',
  position: 'member',
  decision_id: null,
};

const card: AssociationCard = {
  id: associationId,
  name: 'ОСББ «Тестовий Будинок»',
  edrpou: '12345678',
  address: 'м. Київ, вул. Тестова, 1',
  state: 'connected',
  head_full_name: 'Новий Голова Тестовий',
  connected_by_hand: false,
  history: [
    {
      kind: 'transfer_approved',
      occurred_at: '2026-10-21T08:05:00Z',
      decision_id: decisionId,
      filed_at: '2026-10-20T09:00:00Z',
      reason: null,
      head_full_name: 'Новий Голова Тестовий',
      ended_head_full_name: 'Колишній Голова Тестовий',
      signer_full_name: 'Новий Голова Тестовий',
      extract_id: extractId,
      extract_dated_on: '2026-10-18',
    },
  ],
};
const filedOn = /20\.10\.2026|20 жовтня 2026/;
const approvedOn = /21\.10\.2026|21 жовтня 2026/;

const json = (status: number, body: unknown) =>
  new Response(JSON.stringify(body), { status, headers: { 'Content-Type': status >= 400 ? 'application/problem+json' : 'application/json' } });
const problem = (status: number, code: string | undefined, detail: string, errors?: { pointer: string; detail: string }[]) =>
  json(status, { type: 'about:blank', title: 'Problem', status, detail, ...(code ? { code } : {}), ...(errors ? { errors } : {}) });

let members: BoardMember[];
let refuseWrite: (() => Response) | null;
let refuseRead: (() => Response) | null;
let calls: { path: string; method: string; body?: unknown }[];

const path = (url: string) => new URL(url, 'http://localhost').pathname;

function fakeApi(url: string, init?: RequestInit): Response {
  const p = path(url);
  const method = init?.method ?? 'GET';
  const body = typeof init?.body === 'string' ? JSON.parse(init.body) : undefined;
  calls.push({ path: p, method, body });
  if (refuseRead && method === 'GET') return refuseRead();
  if (p === getListBoardMembersUrl(associationId) && method === 'GET') return json(200, { items: members, next_cursor: null });
  if (p === getCreateBoardMemberUrl(associationId) && method === 'POST') {
    if (refuseWrite) return refuseWrite();
    const created: BoardMember = { ...plainMember, id: '00000000-0000-7000-8000-00000000b099', sign_in_email: null, term_end: null, decision_id: null, ...body };
    members = [created, ...members];
    return json(201, created);
  }
  if (p.startsWith(`${getCreateBoardMemberUrl(associationId)}/`) && method === 'PUT') {
    if (refuseWrite) return refuseWrite();
    const id = p.split('/').pop();
    members = members.map((m) => (m.id === id ? { ...m, ...body } : m));
    return json(200, members.find((m) => m.id === id));
  }
  if (p === `/api/identity/associations/${associationId}` && method === 'GET') return json(200, card);
  return json(404, {});
}

async function renderScreen() {
  const router = createRouter({
    routeTree: createRootRoute({
      component: () => (
        <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } })}>
          <RosterScreen associationId={associationId} />
        </QueryClientProvider>
      ),
    }),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  });
  render(<RouterProvider router={router} />);
}

// The block of one roster entry: the nearest list item, row or article around the person's name.
const entryOf = async (name: string) => {
  const entry = (await screen.findByText(name)).closest('li, tr, article, [role="row"]');
  expect(entry, `an entry element around «${name}»`).not.toBeNull();
  return within(entry as HTMLElement);
};

const writes = () => calls.filter((c) => c.method !== 'GET');

beforeEach(() => {
  members = [newHead, previousHead, plainMember];
  refuseWrite = null;
  refuseRead = null;
  calls = [];
  vi.stubGlobal('fetch', vi.fn(async (url: string, init?: RequestInit) => fakeApi(String(url), init)));
});
afterEach(() => vi.unstubAllGlobals());

describe('RosterScreen: the roster listed (AC-09)', () => {
  it('lists every член правління read from the roster with the position and the term', async () => {
    await renderScreen();
    expect(await screen.findByRole('heading', { level: 1, name: 'Правління' })).toBeInTheDocument();
    for (const m of [newHead, previousHead, plainMember]) expect(await screen.findByText(m.full_name)).toBeInTheDocument();
    expect(calls.some((c) => c.path === getListBoardMembersUrl(associationId) && c.method === 'GET')).toBe(true);
    expect((await entryOf(newHead.full_name)).getByText(/голова правління/i)).toBeInTheDocument();
    const member = await entryOf(plainMember.full_name);
    expect(member.getByText(/член правління/i)).toBeInTheDocument();
    expect(member.getByText(/01\.03\.2025|1 березня 2025/)).toBeInTheDocument();
  });

  it('says the roster is empty when there are no entries', async () => {
    members = [];
    await renderScreen();
    expect(await screen.findByText('Правління ще не заповнене')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Додати члена правління' })).toBeInTheDocument();
  });

  it('shows one alert with a retry when the roster cannot be read', async () => {
    refuseRead = () => new Response('boom', { status: 500 });
    await renderScreen();
    expect(await screen.findByRole('alert')).toHaveTextContent('Не вдалося виконати запит');
    expect(screen.getByRole('button', { name: 'Спробувати ще раз' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Додати члена правління' })).not.toBeInTheDocument();
  });
});

describe('RosterScreen: the decision mark (AC-20)', () => {
  it('marks the new голова as appointed and the previous голова as ended, with the dates and the extract link', async () => {
    await renderScreen();
    const extractHref = getDownloadConfirmationExtractUrl(associationId, extractId);

    const appointed = await entryOf(newHead.full_name);
    const appointedText = (await screen.findByText(/призначено за запитом на передачу від/)).closest('li, tr, article, [role="row"]')?.textContent ?? '';
    expect(appointedText).toContain(newHead.full_name);
    expect(appointedText).toMatch(filedOn);
    expect(appointedText).toContain('схвалено оператором');
    expect(appointedText).toMatch(approvedOn);
    expect(appointed.getByRole('link', { name: 'Витяг' }).getAttribute('href')).toBe(extractHref);

    const ended = await entryOf(previousHead.full_name);
    const endedText = (await screen.findByText(/повноваження завершено за запитом на передачу від/)).closest('li, tr, article, [role="row"]')?.textContent ?? '';
    expect(endedText).toContain(previousHead.full_name);
    expect(endedText).toMatch(filedOn);
    expect(endedText).toMatch(approvedOn);
    expect(ended.getByRole('link', { name: 'Витяг' }).getAttribute('href')).toBe(extractHref);
  });

  it('shows no mark and no extract link on an entry without a decision', async () => {
    await renderScreen();
    const member = await entryOf(plainMember.full_name);
    expect(member.queryByText(/за запитом на передачу/)).not.toBeInTheDocument();
    expect(member.queryByRole('link', { name: 'Витяг' })).not.toBeInTheDocument();
    // Exactly the two голова entries carry a mark.
    expect(screen.getAllByText(/за запитом на передачу від/)).toHaveLength(2);
  });
});

describe('RosterScreen: add a член правління (AC-09)', () => {
  const fillAdd = () => {
    fireEvent.change(screen.getByLabelText(/ПІБ/), { target: { value: 'Додана Особа Тестова' } });
    fireEvent.change(screen.getByLabelText(/Електронна пошта/), { target: { value: 'added@example.test' } });
    fireEvent.change(screen.getByLabelText(/Початок строку/), { target: { value: '2026-11-01' } });
  };

  it('posts the person with the term through the generated operation and lists them afterwards', async () => {
    await renderScreen();
    fireEvent.click(await screen.findByRole('button', { name: 'Додати члена правління' }));
    fillAdd();
    fireEvent.click(screen.getByRole('button', { name: 'Зберегти' }));

    expect(await screen.findByText('Додана Особа Тестова')).toBeInTheDocument();
    const post = writes().find((c) => c.method === 'POST');
    expect(post?.path).toBe(getCreateBoardMemberUrl(associationId));
    expect(post?.body).toMatchObject({ full_name: 'Додана Особа Тестова', sign_in_email: 'added@example.test', term_start: '2026-11-01' });
    expect((post?.body as { term_end?: string | null }).term_end ?? null).toBeNull();
  });

  it('names every field the server refused, next to that field, all at once', async () => {
    refuseWrite = () =>
      problem(422, undefined, 'Перевірте поля', [
        { pointer: '/full_name', detail: 'Вкажіть ПІБ' },
        { pointer: '/sign_in_email', detail: 'Некоректна адреса пошти' },
        { pointer: '/term_start', detail: 'Вкажіть початок строку' },
        { pointer: '/term_end', detail: 'Кінець строку не раніше початку' },
      ]);
    await renderScreen();
    fireEvent.click(await screen.findByRole('button', { name: 'Додати члена правління' }));
    fillAdd();
    fireEvent.click(screen.getByRole('button', { name: 'Зберегти' }));

    await waitFor(() => expect(screen.getByLabelText(/ПІБ/)).toHaveAccessibleDescription(expect.stringContaining('Вкажіть ПІБ')));
    expect(screen.getByLabelText(/Електронна пошта/)).toHaveAccessibleDescription(expect.stringContaining('Некоректна адреса пошти'));
    expect(screen.getByLabelText(/Початок строку/)).toHaveAccessibleDescription(expect.stringContaining('Вкажіть початок строку'));
    expect(screen.getByLabelText(/Кінець строку/)).toHaveAccessibleDescription(expect.stringContaining('Кінець строку не раніше початку'));
    for (const label of [/ПІБ/, /Електронна пошта/, /Початок строку/, /Кінець строку/]) expect(screen.getByLabelText(label)).toBeInvalid();
    // The form stays open with what was typed, and the roster did not change.
    expect(screen.getByLabelText(/ПІБ/)).toHaveValue('Додана Особа Тестова');
    expect(screen.queryByText('Додана Особа Тестова', { selector: 'li *, tr *, article *' })).not.toBeInTheDocument();
  });
});

describe('RosterScreen: change a член правління (AC-09)', () => {
  it('opens the entry in the form, saves a new revision with the term end through the generated operation', async () => {
    await renderScreen();
    const member = await entryOf(plainMember.full_name);
    fireEvent.click(member.getByRole('button', { name: /Змінити/ }));

    expect(screen.getByLabelText(/ПІБ/)).toHaveValue(plainMember.full_name);
    expect(screen.getByLabelText(/Електронна пошта/)).toHaveValue('member@example.test');
    expect(screen.getByLabelText(/Початок строку/)).toHaveValue('2025-03-01');
    fireEvent.change(screen.getByLabelText(/Кінець строку/), { target: { value: '2026-12-31' } });
    fireEvent.click(screen.getByRole('button', { name: 'Зберегти' }));

    await waitFor(() => expect(writes()).toHaveLength(1));
    const put = writes()[0];
    expect(put?.method).toBe('PUT');
    expect(put?.path).toBe(getUpdateBoardMemberUrl(associationId, memberId));
    expect(put?.body).toMatchObject({ full_name: plainMember.full_name, sign_in_email: 'member@example.test', term_start: '2025-03-01', term_end: '2026-12-31' });
    expect(await screen.findByText(/31\.12\.2026|31 грудня 2026/)).toBeInTheDocument();
  });

  it('names every field the server refused on a change', async () => {
    refuseWrite = () =>
      problem(422, undefined, 'Перевірте поля', [
        { pointer: '/full_name', detail: 'Вкажіть ПІБ' },
        { pointer: '/term_end', detail: 'Кінець строку не раніше початку' },
      ]);
    await renderScreen();
    fireEvent.click((await entryOf(plainMember.full_name)).getByRole('button', { name: /Змінити/ }));
    fireEvent.change(screen.getByLabelText(/Кінець строку/), { target: { value: '2020-01-01' } });
    fireEvent.click(screen.getByRole('button', { name: 'Зберегти' }));

    await waitFor(() => expect(screen.getByLabelText(/ПІБ/)).toHaveAccessibleDescription(expect.stringContaining('Вкажіть ПІБ')));
    expect(screen.getByLabelText(/Кінець строку/)).toHaveAccessibleDescription(expect.stringContaining('Кінець строку не раніше початку'));
  });
});

describe('RosterScreen: a вимкнене ОСББ (AC-25)', () => {
  it('shows the suspended notice with the reason and no way to add or change', async () => {
    refuseRead = () => problem(403, 'identity.association_disabled', 'спір щодо складу правління');
    await renderScreen();
    expect(await screen.findByText('Доступ призупинено: спір щодо складу правління')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Додати члена правління' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /Змінити/ })).not.toBeInTheDocument();
    expect(screen.queryByLabelText(/ПІБ/)).not.toBeInTheDocument();
  });
});
