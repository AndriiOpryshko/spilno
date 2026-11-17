import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createMemoryHistory, createRootRoute, createRouter, RouterProvider } from '@tanstack/react-router';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { getDownloadConfirmationExtractUrl, getGetAssociationCardUrl } from '@/api/gen/endpoints/identity/identity';
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
  decision_effect: 'ended',
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
  decision_effect: 'seated',
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
  decision_effect: null,
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
  if (p === getGetAssociationCardUrl(associationId) && method === 'GET') return json(200, card);
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

  it('keeps the new голова «призначено» when that голова later gets a term end: the kind comes from the decision, not from the term', async () => {
    members = [{ ...newHead, term_end: '2027-12-31' }, previousHead, plainMember];
    await renderScreen();
    const appointed = await entryOf(newHead.full_name);
    expect(appointed.getByText(/призначено за запитом на передачу від/)).toBeInTheDocument();
    expect(appointed.queryByText(/повноваження завершено/)).not.toBeInTheDocument();
    expect(appointed.getByText(/31\.12\.2027|31 грудня 2027/)).toBeInTheDocument();
    const ended = await entryOf(previousHead.full_name);
    expect(ended.getByText(/повноваження завершено за запитом на передачу від/)).toBeInTheDocument();
    expect(ended.queryByText(/призначено за запитом/)).not.toBeInTheDocument();
  });

  it('marks a голова with a term end «призначено» when the decision ended no one by that name', async () => {
    const connectedOnly = { ...card.history[0]!, ended_head_full_name: null };
    const entries = [{ ...newHead, term_end: '2027-12-31' }];
    members = entries;
    const original = fakeApi;
    vi.stubGlobal('fetch', vi.fn(async (url: string, init?: RequestInit) => (path(String(url)) === getGetAssociationCardUrl(associationId) ? json(200, { ...card, history: [connectedOnly] }) : original(String(url), init))));
    await renderScreen();
    const entry = await entryOf(newHead.full_name);
    expect(await entry.findByText(/призначено за запитом на передачу від/)).toBeInTheDocument();
    expect(entry.queryByText(/повноваження завершено/)).not.toBeInTheDocument();
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

describe('RosterScreen: the mark follows the decision effect and kind (AC-20, AC-29)', () => {
  it('marks the entry with decision_effect ended as ended when both голови share one full name', async () => {
    const same = 'Однаковий Голова Тестовий';
    members = [{ ...newHead, full_name: same }, { ...previousHead, full_name: same }, plainMember];
    await renderScreen();
    await screen.findAllByText(same);
    const entries = screen.getAllByText(same).map((el) => el.closest('li') as HTMLElement);
    const texts = entries.map((li) => li.textContent ?? '');
    expect(texts.filter((x) => x.includes('призначено за запитом на передачу від'))).toHaveLength(1);
    expect(texts.filter((x) => x.includes('повноваження завершено за запитом на передачу від'))).toHaveLength(1);
    const seated = entries.find((li) => li.textContent?.includes('з 25.10.2026') || /25\.10\.2026|25 жовтня 2026/.test(li.textContent ?? ''));
    expect(seated?.textContent).toContain('призначено за запитом на передачу від');
  });

  it('words the mark of the голова of a connection by the заявка на підключення', async () => {
    const connected = { ...card.history[0]!, kind: 'connected' as const, ended_head_full_name: null };
    members = [newHead, plainMember];
    const original = fakeApi;
    vi.stubGlobal('fetch', vi.fn(async (url: string, init?: RequestInit) => (path(String(url)) === getGetAssociationCardUrl(associationId) ? json(200, { ...card, history: [connected] }) : original(String(url), init))));
    await renderScreen();
    const entry = await entryOf(newHead.full_name);
    expect(await entry.findByText(/призначено за заявкою на підключення від/)).toBeInTheDocument();
    expect(entry.queryByText(/запитом на передачу/)).not.toBeInTheDocument();
  });

  it('replaces the term fields of an ended entry by a hint to add the person again', async () => {
    await renderScreen();
    fireEvent.click((await entryOf(previousHead.full_name)).getByRole('button', { name: /Змінити/ }));
    expect(screen.queryByLabelText(/Початок строку/)).not.toBeInTheDocument();
    expect(screen.queryByLabelText(/Кінець строку/)).not.toBeInTheDocument();
    expect(screen.getByText('Щоб повернути людину до правління, додайте запис знову')).toBeInTheDocument();
  });

  it('keeps the term fields on a seated entry', async () => {
    await renderScreen();
    fireEvent.click((await entryOf(newHead.full_name)).getByRole('button', { name: /Змінити/ }));
    expect(screen.getByLabelText(/Кінець строку/)).toBeInTheDocument();
    expect(screen.queryByText(/Щоб повернути людину/)).not.toBeInTheDocument();
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

  // The real server answers in English (governance/domain/board_member.go): the screen names each field in
  // Ukrainian, from its own strings, and never shows the server's raw detail.
  const serverDetails = {
    '/full_name': 'must be 1 to 200 characters',
    '/sign_in_email': 'must be at most 320 characters',
    '/term_start': 'is required',
    '/term_end': 'must not be before term_start',
  };
  const refuseWithRealDetails = (pointers: (keyof typeof serverDetails | '/position')[]) => () =>
    problem(
      422,
      'governance.roster_entry_invalid',
      'The request is invalid',
      pointers.map((pointer) => ({ pointer, detail: pointer === '/position' ? 'must be head or member' : serverDetails[pointer] })),
    );
  const isUkrainian = (text: string) => /[А-Яа-яІіЇїЄєҐґ]/.test(text) && !/[A-Za-z]{4,}/.test(text);
  const descriptionOf = (label: RegExp) => screen.getByLabelText(label).getAttribute('aria-describedby') ?? '';
  const describedText = (label: RegExp) =>
    descriptionOf(label)
      .split(/\s+/)
      .map((id) => document.getElementById(id)?.textContent ?? '')
      .join(' ')
      .trim();

  it('names every field the server refused, next to that field, in Ukrainian, all at once', async () => {
    refuseWrite = refuseWithRealDetails(['/full_name', '/sign_in_email', '/term_start', '/term_end']);
    await renderScreen();
    fireEvent.click(await screen.findByRole('button', { name: 'Додати члена правління' }));
    fillAdd();
    fireEvent.click(screen.getByRole('button', { name: 'Зберегти' }));

    const labels = [/ПІБ/, /Електронна пошта/, /Початок строку/, /Кінець строку/];
    await waitFor(() => expect(describedText(/ПІБ/)).not.toBe(''));
    const texts = labels.map(describedText);
    for (const [i, text] of texts.entries()) {
      expect(isUkrainian(text), `a Ukrainian message for ${String(labels[i])}, got «${text}»`).toBe(true);
      expect(screen.getByLabelText(labels[i]!)).toBeInvalid();
    }
    // Each field says its own thing, not one shared line.
    expect(new Set(texts).size).toBe(4);
    for (const raw of Object.values(serverDetails)) expect(document.body.textContent).not.toContain(raw);
    // The form stays open with what was typed, and the roster did not change.
    expect(screen.getByLabelText(/ПІБ/)).toHaveValue('Додана Особа Тестова');
    expect(screen.queryByText('Додана Особа Тестова', { selector: 'li *, tr *, article *' })).not.toBeInTheDocument();
  });

  it('shows a refused pointer that matches no visible field in the refusal alert, not silently', async () => {
    refuseWrite = refuseWithRealDetails(['/full_name', '/position']);
    await renderScreen();
    fireEvent.click(await screen.findByRole('button', { name: 'Додати члена правління' }));
    fillAdd();
    fireEvent.click(screen.getByRole('button', { name: 'Зберегти' }));

    const alert = await screen.findByRole('alert');
    expect(isUkrainian(alert.textContent ?? ''), `a Ukrainian refusal, got «${alert.textContent}»`).toBe(true);
    expect(alert).not.toHaveTextContent('must be head or member');
    expect(alert).not.toHaveTextContent('The request is invalid');
    await waitFor(() => expect(isUkrainian(describedText(/ПІБ/))).toBe(true));
  });

  it('shows a refusal that names only an unknown pointer in the alert', async () => {
    refuseWrite = refuseWithRealDetails(['/position']);
    await renderScreen();
    fireEvent.click(await screen.findByRole('button', { name: 'Додати члена правління' }));
    fillAdd();
    fireEvent.click(screen.getByRole('button', { name: 'Зберегти' }));

    const alert = await screen.findByRole('alert');
    expect(isUkrainian(alert.textContent ?? ''), `a Ukrainian refusal, got «${alert.textContent}»`).toBe(true);
    expect(alert).not.toHaveTextContent('must be head or member');
    for (const label of [/ПІБ/, /Електронна пошта/, /Початок строку/, /Кінець строку/]) expect(screen.getByLabelText(label)).not.toBeInvalid();
  });

  it('shows «Доступ призупинено: <reason>» when the ОСББ was disabled while the form was open', async () => {
    refuseWrite = () => problem(403, 'identity.association_disabled', 'спір щодо складу правління');
    await renderScreen();
    fireEvent.click(await screen.findByRole('button', { name: 'Додати члена правління' }));
    fillAdd();
    fireEvent.click(screen.getByRole('button', { name: 'Зберегти' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('Доступ призупинено: спір щодо складу правління');
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

  it('names every field the server refused on a change, in Ukrainian', async () => {
    refuseWrite = () =>
      problem(422, 'governance.roster_entry_invalid', 'The request is invalid', [
        { pointer: '/full_name', detail: 'must be 1 to 200 characters' },
        { pointer: '/term_end', detail: 'must not be before term_start' },
      ]);
    await renderScreen();
    fireEvent.click((await entryOf(plainMember.full_name)).getByRole('button', { name: /Змінити/ }));
    fireEvent.change(screen.getByLabelText(/Кінець строку/), { target: { value: '2020-01-01' } });
    fireEvent.click(screen.getByRole('button', { name: 'Зберегти' }));

    const text = (label: RegExp) =>
      (screen.getByLabelText(label).getAttribute('aria-describedby') ?? '')
        .split(/\s+/)
        .map((id) => document.getElementById(id)?.textContent ?? '')
        .join(' ')
        .trim();
    await waitFor(() => expect(text(/ПІБ/)).not.toBe(''));
    for (const label of [/ПІБ/, /Кінець строку/]) {
      expect(text(label)).toMatch(/[А-Яа-яІіЇїЄє]/);
      expect(text(label)).not.toMatch(/must|characters|term_start/);
    }
    expect(text(/ПІБ/)).not.toBe(text(/Кінець строку/));
  });
});

describe('RosterScreen: a вимкнене ОСББ (AC-25)', () => {
  it('renders nothing of its own: the layout owns the suspended notice, and there is no way to add or change', async () => {
    refuseRead = () => problem(403, 'identity.association_disabled', 'спір щодо складу правління');
    await renderScreen();
    await waitFor(() => expect(calls.some((c) => c.path === getListBoardMembersUrl(associationId))).toBe(true));
    await new Promise((resolve) => setTimeout(resolve, 100));
    expect(screen.queryByText(/Доступ призупинено/)).not.toBeInTheDocument();
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
    expect(screen.queryByLabelText(/ПІБ/)).not.toBeInTheDocument();
  });
});
