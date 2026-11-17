import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createMemoryHistory, createRootRoute, createRouter, RouterProvider } from '@tanstack/react-router';
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { getDownloadFileUrl } from '@/api/gen/endpoints/governance/governance';
import { Sidebar } from '@/components/app-shell';
import { formatUah, uk } from '@/i18n/uk';
import { routeTree } from '@/routeTree.gen';

// The реєстр протоколів through the real route tree (AC-26, AC-09; board-protocol T43 scope): search with
// cursor pages, an empty registry explained, the signed protocol with its download links and no way to
// change it, all of it under the suspended notice too (ADR-0007: the registry reads and the signed-copy
// download stay open in a вимкнене ОСББ). The fake answers like the server.
//
// Contract for the implementer (labels are matched by regex, so the exact copy is free):
// - filter fields labelled «Контрагент», «Предмет», «Сума від», «Сума до» (hryvnias, sent as kopiyky),
//   «Дата від», «Дата до»; a submit button named «Знайти» or «Шукати»;
// - a results list whose rows are li / tr / article / role=row|listitem, each holding a link to the signed view;
// - a «показати ще / далі / наступна» button while next_cursor is set;
// - the signed view links every file through the generated getDownloadFileUrl.

const lypky = '00000000-0000-7000-8000-000000000001';
const protocolId = '00000000-0000-7000-8000-000000000e01';
const principal = { account_id: '00000000-0000-7000-8000-00000000a001', email: 'user-1@example.test', is_operator: false };
const reason = 'спір щодо складу правління';
const memberId = '00000000-0000-7000-8000-00000000b001';

const decision = (n: number, subject: string, counterparty: string | null, amount: number | null, id = protocolId) => ({
  protocol_id: id,
  agenda_item_id: `00000000-0000-7000-8000-0000000010${n}`,
  number: n,
  meeting_date: '2026-11-05',
  subject,
  counterparty,
  amount_kopiyky: amount,
  currency: amount === null ? null : 'UAH',
});

const pageOne = {
  items: [
    decision(14, 'Про затвердження договору на гідроізоляцію покрівлі', 'ТОВ Тест Покрівля', 18450000),
    decision(12, 'Про заміну лічильників води', 'ФОП Тест Лічильник', 4200000, '00000000-0000-7000-8000-000000000e02'),
  ],
  withdrawn_numbers: [13],
  next_cursor: 'cursor-2',
};
const pageTwo = {
  items: [decision(11, 'Про прибирання прибудинкової території', 'ТОВ Тест Чисто', 1500000, '00000000-0000-7000-8000-000000000e03')],
  withdrawn_numbers: [],
  next_cursor: null,
};

const fileRef = (id: string, name: string) => ({
  id,
  original_name: name,
  media_type: 'application/pdf',
  size_bytes: 91234,
  sha256: '2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae',
  uploaded_by_member_id: memberId,
  created_at: '2026-11-01T12:00:04Z',
});
const scanId = '00000000-0000-7000-8000-000000000704';
const printId = '00000000-0000-7000-8000-000000000703';
const invoiceId = '00000000-0000-7000-8000-000000000702';
const agendaItemId = '00000000-0000-7000-8000-000000001001';
const informalLabel = 'Попереднє внутрішнє опитування правління — не є голосуванням засідання і не замінює його';

const signed = {
  protocol_id: protocolId,
  number: 14,
  meeting_date: '2026-11-05',
  meeting: { id: '00000000-0000-7000-8000-000000000f01', held_at: '2026-11-05T16:00:00Z', meeting_date: '2026-11-05', chair_board_member_id: memberId },
  agenda_items: [
    {
      id: agendaItemId,
      position: 1,
      case_id: null,
      subject: 'Про затвердження договору на гідроізоляцію покрівлі',
      fragments: [{ text: 'Затвердити договір на гідроізоляцію покрівлі з ТОВ Тест Покрівля.' }],
      wording_author_member_id: memberId,
      wording_revised_at: '2026-11-01T12:00:00Z',
      slots: [],
    },
  ],
  documents: [
    {
      id: '00000000-0000-7000-8000-000000000d02',
      case_id: '00000000-0000-7000-8000-000000000c01',
      file: fileRef(invoiceId, 'rakhunok-123.pdf'),
      previous_version_id: null,
      document_date: '2026-10-18',
      created_by_member_id: memberId,
      created_at: '2026-10-20T09:10:00Z',
    },
  ],
  print_copy: {
    id: '00000000-0000-7000-8000-000000001301',
    protocol_revision: 9,
    number: 14,
    file: fileRef(printId, 'protocol-14.pdf'),
    requested_by_member_id: memberId,
    created_at: '2026-11-01T12:00:04Z',
    voided_at: null,
  },
  meeting_facts: {
    signed_scan: fileRef(scanId, 'protocol-14-signed.pdf'),
    present_member_ids: [memberId],
    item_votes: [{ agenda_item_id: agendaItemId, votes_for: 3, votes_against: 0, votes_abstained: 0 }],
    has_handwritten_corrections: true,
    corrections_note: 'У пункті 1 виправлено номер рахунку від руки, підписано всіма присутніми',
    entered_by_member_id: memberId,
    entered_at: '2026-11-06T08:00:00Z',
  },
  poll_history: [
    {
      id: '00000000-0000-7000-8000-000000001401',
      round_number: 1,
      protocol_revision: 8,
      label: informalLabel,
      opened_by_member_id: memberId,
      created_at: '2026-10-27T11:00:00Z',
      closed_at: '2026-10-27T12:00:00Z',
      closed_by_member_id: memberId,
      frozen_agenda_items: [],
      items: [{ agenda_item_id: agendaItemId, tally: { for: 3, against: 1, abstain: 1 }, votes: [] }],
    },
  ],
  locked_by_member_id: memberId,
  locked_at: '2026-11-06T08:00:00Z',
};

let disabled: boolean;
let registry: (params: URLSearchParams) => Response;
let calls: string[];

const json = (status: number, body: unknown) =>
  new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': status >= 400 ? 'application/problem+json' : 'application/json' },
  });

const filterKeys = ['counterparty', 'subject', 'amount_kopiyky', 'amount_from_kopiyky', 'amount_to_kopiyky', 'date_from', 'date_to'];
const hasFilter = (params: URLSearchParams) => filterKeys.some((k) => params.has(k));

function fakeApi(input: RequestInfo | URL) {
  const url = new URL(String(input), 'http://localhost');
  const path = url.pathname;
  calls.push(`${path}${url.search}`);
  if (path.endsWith('/identity/session')) return Promise.resolve(json(200, principal));
  if (path.endsWith('/identity/memberships')) {
    return Promise.resolve(
      json(200, {
        items: [{ association_id: lypky, association_name: 'ОСББ «Липки»', association_state: disabled ? 'disabled' : 'connected', board_member_id: memberId, position: 'member' }],
        next_cursor: null,
      }),
    );
  }
  if (path.includes('/governance/associations/')) {
    if (path.endsWith('/registry/decisions')) return Promise.resolve(registry(url.searchParams));
    if (path.endsWith(`/registry/protocols/${protocolId}`)) return Promise.resolve(json(200, signed));
    return Promise.resolve(json(404, { type: 'about:blank', title: 'Not found', status: 404, code: 'governance.not_found' }));
  }
  if (path.includes('/identity/associations/')) {
    if (disabled) {
      return Promise.resolve(json(403, { type: 'about:blank', title: 'Forbidden', status: 403, detail: reason, code: 'identity.association_disabled' }));
    }
    return Promise.resolve(json(404, {}));
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

const listPath = `/o/${lypky}/registry`;
const signedPath = `/o/${lypky}/registry/${protocolId}`;
const registryCalls = () => calls.filter((c) => c.includes('/registry/decisions'));
const lastRegistryParams = () => new URL(registryCalls().at(-1) ?? '', 'http://localhost').searchParams;
const norm = (s: string) => s.replace(/\s/g, ' ');
const rowOf = (el: HTMLElement) => el.closest<HTMLElement>('li, tr, article, [role="row"], [role="listitem"]') ?? el;
const searchButton = () => screen.findByRole('button', { name: /Знайти|Шукати/ });
const fill = (label: RegExp, value: string) => fireEvent.change(screen.getByLabelText(label), { target: { value } });

beforeEach(() => {
  disabled = false;
  calls = [];
  registry = (params) => json(200, params.get('cursor') === 'cursor-2' ? pageTwo : pageOne);
  vi.stubGlobal('fetch', vi.fn(fakeApi));
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe('searching the реєстр протоколів (AC-18, AC-26)', () => {
  it('lists the signed decisions with number, date, subject, counterparty and amount, each opening its protocol', async () => {
    open(listPath);
    expect(await screen.findByRole('heading', { level: 1, name: /Реєстр протоколів/ })).toBeInTheDocument();

    const first = await screen.findByRole('link', { name: /гідроізоляці/ });
    expect(first.getAttribute('href')).toBe(signedPath);
    const row = rowOf(first);
    expect(row).toHaveTextContent('14');
    expect(row).toHaveTextContent('ТОВ Тест Покрівля');
    expect(norm(row.textContent ?? '')).toContain(norm(formatUah(184500)));

    const second = screen.getByRole('link', { name: /лічильників води/ });
    expect(second.getAttribute('href')).toBe(`/o/${lypky}/registry/00000000-0000-7000-8000-000000000e02`);
  });

  it('reads the first page without any filter', async () => {
    open(listPath);
    await screen.findByRole('link', { name: /гідроізоляці/ });
    expect(registryCalls().length).toBeGreaterThan(0);
    const params = new URL(registryCalls()[0] ?? '', 'http://localhost').searchParams;
    expect(filterKeys.filter((k) => params.has(k))).toEqual([]);
    expect(params.has('cursor')).toBe(false);
  });

  it('shows a withdrawn number as a gap marked «знято з розгляду»', async () => {
    open(listPath);
    await screen.findByRole('link', { name: /гідроізоляці/ });
    const gap = await screen.findByText(/знято з розгляду/i);
    expect(rowOf(gap)).toHaveTextContent('13');
    // The gap is a marker, not a decision: it opens nothing.
    expect(within(rowOf(gap)).queryByRole('link')).toBeNull();
  });

  it('pages by cursor: the next page is read with the cursor of the previous one, the last page has no more', async () => {
    open(listPath);
    await screen.findByRole('link', { name: /гідроізоляці/ });
    fireEvent.click(await screen.findByRole('button', { name: /ще|далі|наступн/i }));

    expect(await screen.findByRole('link', { name: /прибирання/ })).toBeInTheDocument();
    expect(lastRegistryParams().get('cursor')).toBe('cursor-2');
    await waitFor(() => expect(screen.queryByRole('button', { name: /ще|далі|наступн/i })).toBeNull());
  });

  it('sends the filters of the form as the contract names them, amounts in kopiyky', async () => {
    open(listPath);
    await screen.findByRole('link', { name: /гідроізоляці/ });

    fill(/Контрагент/, 'Покрівля');
    fill(/Предмет/, 'гідроізоляція');
    fill(/Сума від/, '100000');
    fill(/Сума до/, '200000');
    fill(/Дата від/, '2026-01-01');
    fill(/Дата до/, '2026-12-31');
    fireEvent.click(await searchButton());

    await waitFor(() => expect(lastRegistryParams().get('counterparty')).toBe('Покрівля'));
    const params = lastRegistryParams();
    expect(params.get('subject')).toBe('гідроізоляція');
    expect(params.get('amount_from_kopiyky')).toBe('10000000');
    expect(params.get('amount_to_kopiyky')).toBe('20000000');
    expect(params.get('date_from')).toBe('2026-01-01');
    expect(params.get('date_to')).toBe('2026-12-31');
    expect(params.has('cursor')).toBe(false);
  });

  it('marks the broken filters next to them when the search is refused with governance.search_invalid', async () => {
    registry = (params) =>
      hasFilter(params)
        ? json(422, {
            type: 'about:blank',
            title: 'Validation failed',
            status: 422,
            code: 'governance.search_invalid',
            errors: [
              { pointer: '/query/amount_from_kopiyky', detail: 'above amount_to_kopiyky' },
              { pointer: '/query/date_from', detail: 'above date_to' },
            ],
          })
        : json(200, pageOne);
    open(listPath);
    await screen.findByRole('link', { name: /гідроізоляці/ });

    fill(/Сума від/, '300000');
    fill(/Сума до/, '1000');
    fill(/Дата від/, '2026-12-31');
    fill(/Дата до/, '2026-01-01');
    fireEvent.click(await searchButton());

    await waitFor(() => expect(screen.getByLabelText(/Сума від/)).toHaveAttribute('aria-invalid', 'true'));
    expect(screen.getByLabelText(/Сума від/)).toHaveAccessibleDescription(/\S/);
    expect(screen.getByLabelText(/Дата від/)).toHaveAttribute('aria-invalid', 'true');
    expect(screen.getByLabelText(/Дата від/)).toHaveAccessibleDescription(/\S/);
    expect(screen.getByLabelText(/Контрагент/)).not.toHaveAttribute('aria-invalid', 'true');
    expect(screen.getByLabelText(/Предмет/)).not.toHaveAttribute('aria-invalid', 'true');
    // The typed values stay, so the member fixes them instead of retyping.
    expect(screen.getByLabelText(/Сума від/)).toHaveDisplayValue('300000');
  });

  it('explains an empty registry instead of showing a bare list', async () => {
    registry = () => json(200, { items: [], withdrawn_numbers: [], next_cursor: null });
    open(listPath);
    expect(await screen.findByText(/немає підписаних протоколів/i)).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: /гідроізоляці/ })).toBeNull();
    expect(screen.queryByRole('button', { name: /ще|далі|наступн/i })).toBeNull();
  });

  it('says nothing was found and keeps the filters when a search matches nothing', async () => {
    registry = (params) => json(200, hasFilter(params) ? { items: [], withdrawn_numbers: [], next_cursor: null } : pageOne);
    open(listPath);
    await screen.findByRole('link', { name: /гідроізоляці/ });

    fill(/Контрагент/, 'Немає такого');
    fireEvent.click(await searchButton());

    expect(await screen.findByText(/нічого не знайдено/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Контрагент/)).toHaveValue('Немає такого');
    expect(screen.queryByText(/немає підписаних протоколів/i)).toBeNull();
  });
});

describe('the signed protocol (AC-18, AC-26, AC-15, AC-17, AC-41)', () => {
  it('shows the number, the state, the frozen text, the meeting facts and the correction note beside the text', async () => {
    open(signedPath);
    expect(await screen.findByRole('heading', { level: 1, name: /14/ })).toBeInTheDocument();
    expect(screen.getByText(uk.protocol.state.signed)).toBeInTheDocument();
    expect(screen.getByText(/Затвердити договір на гідроізоляцію покрівлі з ТОВ Тест Покрівля/)).toBeInTheDocument();
    expect(screen.getAllByText(/05\.11\.2026/).length).toBeGreaterThan(0);
    expect(screen.getByText(/У пункті 1 виправлено номер рахунку від руки/)).toBeInTheDocument();
  });

  it('links the signed scan, the print copy and a cited document through the generated download URL', async () => {
    open(signedPath);
    await screen.findByRole('heading', { level: 1, name: /14/ });

    const links = screen.getAllByRole('link');
    for (const [id, name] of [
      [scanId, 'protocol-14-signed.pdf'],
      [printId, 'protocol-14.pdf'],
      [invoiceId, 'rakhunok-123.pdf'],
    ] as const) {
      const link = links.find((a) => a.getAttribute('href') === getDownloadFileUrl(lypky, id));
      expect(link, `a link to ${name}`).toBeDefined();
      expect(link).toHaveAccessibleName(new RegExp(`Завантажити|${name.replace('.', '\\.')}`));
    }
  });

  it('labels the poll history informal, not the meeting vote (AC-41)', async () => {
    open(signedPath);
    await screen.findByRole('heading', { level: 1, name: /14/ });
    expect((await screen.findAllByText(/не (є )?голосуванн(я|ям) засідання/i)).length).toBeGreaterThan(0);
  });

  it('explains that a locked protocol changes only through a new decision (AC-17)', async () => {
    open(signedPath);
    await screen.findByRole('heading', { level: 1, name: /14/ });
    expect(screen.getByText(/лише новим/i)).toBeInTheDocument();
  });

  it('links «нове рішення, що посилається на цей протокол» to the справи of the association, where a draft starts (AC-17, T43, F4.1)', async () => {
    open(signedPath);
    await screen.findByRole('heading', { level: 1, name: /14/ });
    const link = within(screen.getByRole('main')).getByRole('link', { name: /нове рішення/i });
    expect(link.getAttribute('href')).toBe(`/o/${lypky}/cases`);
  });

  it('names the association in the top bar of the signed protocol (AC-04d, F4.2)', async () => {
    open(signedPath);
    await screen.findByRole('heading', { level: 1, name: /14/ });
    expect(await within(screen.getByRole('main')).findByText('ОСББ «Липки»')).toBeInTheDocument();
  });

  it('has no edit, upload or withdraw control', async () => {
    open(signedPath);
    await screen.findByRole('heading', { level: 1, name: /14/ });
    const main = screen.getByRole('main');

    expect(within(main).queryAllByRole('textbox')).toEqual([]);
    expect(main.querySelector('input[type="file"], textarea, [contenteditable="true"]')).toBeNull();
    expect(within(main).queryByRole('button', { name: /редагувати|зберегти|завантажити скан|завантажити підписан|зняти з розгляду|повернути в чернетку|надрукувати/i })).toBeNull();
    expect(within(main).queryByRole('link', { name: /редагувати|зняти з розгляду/i })).toBeNull();
  });

  it('links back to the registry', async () => {
    open(signedPath);
    await screen.findByRole('heading', { level: 1, name: /14/ });
    const back = within(screen.getByRole('main')).getAllByRole('link').find((a) => a.getAttribute('href') === listPath);
    expect(back, 'a link back to the registry list').toBeDefined();
  });
});

describe('the registry of a вимкнене ОСББ stays readable (AC-26)', () => {
  beforeEach(() => {
    disabled = true;
  });

  const notice = async () => {
    const alerts = await screen.findAllByRole('alert');
    const found = alerts.find((a) => a.textContent?.includes('Доступ призупинено'));
    expect(found, 'an alert «Доступ призупинено»').toBeDefined();
    expect(found).toHaveTextContent(reason);
  };

  it('lists the decisions under the suspended notice', async () => {
    open(listPath);
    await notice();
    expect(await screen.findByRole('link', { name: /гідроізоляці/ })).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 1, name: /Реєстр протоколів/ })).toBeInTheDocument();
  });

  it('opens the signed protocol with its download links under the suspended notice', async () => {
    open(signedPath);
    await notice();
    await screen.findByRole('heading', { level: 1, name: /14/ });
    const hrefs = screen.getAllByRole('link').map((a) => a.getAttribute('href'));
    expect(hrefs).toContain(getDownloadFileUrl(lypky, scanId));
  });
});

describe('the registry in the navigation (AC-09)', () => {
  it('links «Реєстр протоколів» in the sidebar to the registry of its own association', async () => {
    const router = createRouter({
      routeTree: createRootRoute({ component: () => <Sidebar associationId={lypky} associationName="ОСББ «Липки»" isOperator={false} /> }),
      history: createMemoryHistory({ initialEntries: ['/'] }),
    });
    render(<RouterProvider router={router} />);
    const nav = await screen.findByRole('complementary');
    expect(within(nav).getByRole('link', { name: /Реєстр протоколів/ }).getAttribute('href')).toBe(listPath);
  });

  it('opens the registry from the sidebar of the association layout', async () => {
    const router = open(`/o/${lypky}/association`);
    const nav = await screen.findByRole('complementary');
    fireEvent.click(await within(nav).findByRole('link', { name: /Реєстр протоколів/ }));
    await waitFor(() => expect(router.state.location.pathname).toBe(listPath));
    expect(await screen.findByRole('link', { name: /гідроізоляці/ })).toBeInTheDocument();
  });
});

describe('the registry screens read through the layout cache and the shared fallback (AC-19, AC-26, T103)', () => {
  const expired = (match: (path: string) => boolean) =>
    vi.stubGlobal('fetch', vi.fn((input: RequestInfo | URL) => {
      const path = new URL(String(input), 'http://localhost').pathname;
      return match(path)
        ? Promise.resolve(json(401, { type: 'about:blank', title: 'Unauthorized', status: 401, code: 'identity.session_required' }))
        : fakeApi(input);
    }));

  it('counts one memberships request for a visit to the signed protocol', async () => {
    open(signedPath);
    await screen.findByRole('heading', { level: 1 });
    await new Promise((resolve) => setTimeout(resolve, 200));
    expect(calls.filter((c) => c.endsWith('/identity/memberships'))).toHaveLength(1);
  });

  it('counts one memberships request for a visit to the registry', async () => {
    open(listPath);
    await screen.findByRole('link', { name: /гідроізоляці/ });
    await new Promise((resolve) => setTimeout(resolve, 200));
    expect(calls.filter((c) => c.endsWith('/identity/memberships'))).toHaveLength(1);
  });

  it('leads an expired session to sign-in from the registry', async () => {
    expired((path) => path.endsWith('/registry/decisions'));
    const router = open(listPath);
    await waitFor(() => expect(router.state.location.pathname).toBe('/sign-in'));
  });

  it('leads an expired session to sign-in from the signed protocol', async () => {
    expired((path) => path.endsWith(`/registry/protocols/${protocolId}`));
    const router = open(signedPath);
    await waitFor(() => expect(router.state.location.pathname).toBe('/sign-in'));
  });
});
