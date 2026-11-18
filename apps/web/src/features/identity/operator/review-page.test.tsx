import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createMemoryHistory, createRootRoute, createRouter, RouterProvider } from '@tanstack/react-router';
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import type { ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { getDownloadExtractForReviewUrl } from '@/api/gen/endpoints/identity/identity';
import type { ApplicationReview } from '@/api/gen/model';

import { formatDate, uk } from '@/i18n/uk';

import { OperatorQueue } from './queue';
import { ReviewPage } from './review-page';

const appId = '00000000-0000-7000-8000-0000000000a1';
const filingId = '00000000-0000-7000-8000-0000000000f1';
const otherId = '00000000-0000-7000-8000-0000000000a2';
const refiledId = '00000000-0000-7000-8000-0000000000f2';

const application = {
  id: appId,
  kind: 'connection',
  edrpou: '12345678',
  association_id: null,
  association_name: 'ОСББ «Тестовий Будинок»',
  state: 'pending',
  filed_at: '2026-11-10T09:00:00Z',
  rejection_reason: null,
  created_at: '2026-11-10T09:00:00Z',
} as ApplicationReview['application'];

const base: ApplicationReview = {
  application,
  applicant_email: 'head@example.test',
  filings: [
    {
      id: filingId,
      association_name: 'ОСББ «Тестовий Будинок»',
      association_address: 'м. Київ, вул. Тестова, 1',
      applicant_full_name: 'Тестовий Голова Правління',
      position: 'head',
      term_start: '2026-01-01',
      term_end: null,
      extract: { id: 'e1', original_name: 'vytyah.pdf', media_type: 'application/pdf', size_bytes: 10, sha256: 'a'.repeat(64) },
      created_at: '2026-11-10T09:00:00Z',
    },
  ],
  checks: [],
  decisions: [],
  warnings: [],
  competing_application_ids: [],
  storage_account_email: null,
};

const checked: ApplicationReview = {
  ...base,
  checks: [
    {
      id: 'c1',
      filing_id: filingId,
      signature_valid: true,
      signer_full_name: 'Тестовий Голова Правління',
      signed_on: '2026-11-08',
      head_full_name: 'Тестовий Голова Правління',
      extract_edrpou: '12345678',
      extract_dated_on: '2026-11-05',
      checked_at: '2026-11-11T08:05:00Z',
    },
  ],
};

let review: ApplicationReview;
type QueueItem = ApplicationReview['application'] & { competing: boolean };
let queueItems: QueueItem[];
let otherReviews: Record<string, ApplicationReview>;
let queueStatus: number;
let urls: string[];
let sent: { path: string; body: unknown }[];

const json = (status: number, body: unknown) =>
  new Response(JSON.stringify(body), { status, headers: { 'Content-Type': status >= 400 ? 'application/problem+json' : 'application/json' } });

function fakeApi(input: RequestInfo | URL, init?: RequestInit) {
  const path = new URL(String(input), 'http://localhost').pathname;
  urls.push(path);
  if (init?.method === 'POST') {
    sent.push({ path, body: JSON.parse(String(init.body ?? '{}')) });
    return Promise.resolve(json(200, {}));
  }
  if (path.endsWith('/identity/operator/applications')) {
    if (queueStatus !== 200) {
      const code = queueStatus === 403 ? { code: 'identity.not_operator' } : {};
      return Promise.resolve(json(queueStatus, { type: 'about:blank', title: 'x', status: queueStatus, ...code }));
    }
    return Promise.resolve(json(200, { items: queueItems, next_cursor: null }));
  }
  if (path.endsWith(`/identity/operator/applications/${appId}`)) return Promise.resolve(json(200, review));
  const other = Object.entries(otherReviews).find(([id]) => path.endsWith(`/identity/operator/applications/${id}`));
  if (other) return Promise.resolve(json(200, other[1]));
  return Promise.resolve(json(404, {}));
}

async function renderInRouter(ui: ReactNode) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  const router = createRouter({
    routeTree: createRootRoute({ component: () => <QueryClientProvider client={client}>{ui}</QueryClientProvider> }),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  });
  render(<RouterProvider router={router} />);
}

beforeEach(() => {
  review = base;
  queueItems = [{ ...application, competing: false }];
  otherReviews = {};
  queueStatus = 200;
  urls = [];
  sent = [];
  vi.stubGlobal('fetch', vi.fn(fakeApi));
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe('ReviewPage', () => {
  // The payloads the backend sends (review_warning.go, TestReview_NameWarningSaysWhichValuesDiffer): `detail` is
  // English and `differing` says which of the three values differ, in the order typed, signer, head.
  const typed = 'Іваненко Іван Іванович';
  const signer = 'Петренко Петро Петрович';
  const head = 'Сидоренко Сидір Сидорович';
  const mismatch = (signerName: string, headName: string, warning: { detail: string; differing: ('typed' | 'signer' | 'head')[] }): ApplicationReview => ({
    ...checked,
    filings: [{ ...checked.filings[0]!, applicant_full_name: typed }],
    checks: [{ ...checked.checks[0]!, signer_full_name: signerName, head_full_name: headName }],
    warnings: [{ kind: 'name_mismatch', ...warning }],
  });
  const nameCases: [string, ApplicationReview, string[], string[]][] = [
    [
      'all three differ',
      mismatch(signer, head, {
        detail: `names differ: typed name "${typed}", extract signer "${signer}", head in extract "${head}"`,
        differing: ['typed', 'signer', 'head'],
      }),
      [typed, signer, head],
      [],
    ],
    ['only the head differs', mismatch(typed, head, { detail: `names differ: head in extract "${head}"`, differing: ['head'] }), [head], [typed, signer]],
    [
      'only the signer differs',
      mismatch(signer, typed, { detail: `names differ: extract signer "${signer}"`, differing: ['signer'] }),
      [signer],
      [typed, head],
    ],
    [
      'only the typed name differs',
      mismatch(signer, signer, { detail: `names differ: typed name "${typed}"`, differing: ['typed'] }),
      [typed],
      [signer, head],
    ],
  ];

  it.each(nameCases)('writes the name warning in Ukrainian naming only the differing values: %s (AC-11)', async (_name, fixture, named, notNamed) => {
    review = fixture;
    await renderInRouter(<ReviewPage applicationId={appId} />);
    const warning = await screen.findByText(/розбіжність імен/i);
    const box = warning.closest('[role="status"]') as HTMLElement;
    for (const n of named) expect(box).toHaveTextContent(n);
    for (const n of notNamed) expect(box).not.toHaveTextContent(n);
    expect(screen.queryByText(/names differ|typed name|extract signer|head in extract/)).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Схвалити/ })).toBeEnabled();
  });

  it('writes the extract-age warning with a дд.мм.рррр date, never the server text, and keeps «Схвалити» available (AC-11b)', async () => {
    review = {
      ...checked,
      checks: [{ ...checked.checks[0]!, extract_dated_on: '2026-09-01' }],
      warnings: [{ kind: 'extract_age', detail: 'extract dated 2026-09-01 is more than 30 days old', differing: [] }],
    };
    await renderInRouter(<ReviewPage applicationId={appId} />);
    const warning = await screen.findByText(/старший за 30 днів/);
    expect(warning.closest('[role="status"]')).toHaveTextContent('Витяг сформовано 01.09.2026 — старший за 30 днів');
    expect(screen.queryByText(/extract dated|more than 30 days/)).not.toBeInTheDocument();
    expect(screen.queryByText(/2026-09-01/)).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Схвалити/ })).toBeEnabled();
  });

  it('marks a competing заявка (AC-04c)', async () => {
    review = { ...base, competing_application_ids: [otherId] };
    await renderInRouter(<ReviewPage applicationId={appId} />);
    expect(await screen.findByText(/конкуруюча заявка/i)).toBeInTheDocument();
  });

  it('marks a competing запит (AC-19b)', async () => {
    review = { ...base, application: { ...application, kind: 'transfer' }, competing_application_ids: [otherId] };
    await renderInRouter(<ReviewPage applicationId={appId} />);
    expect(await screen.findByText(/конкуруючий запит/i)).toBeInTheDocument();
  });

  it('shows the storage account that holds the files of a запит (AC-28b)', async () => {
    review = { ...base, application: { ...application, kind: 'transfer' }, storage_account_email: 'files@example.test' };
    await renderInRouter(<ReviewPage applicationId={appId} />);
    expect(await screen.findByText(/files@example\.test/)).toBeInTheDocument();
  });

  it('offers the check form with the fields of the check', async () => {
    await renderInRouter(<ReviewPage applicationId={appId} />);
    await screen.findAllByText(/введено заявником/);
    expect(screen.getByLabelText(/Підписант/)).toBeInTheDocument();
    expect(screen.getByLabelText(/Керівник/)).toBeInTheDocument();
    expect(screen.getByLabelText(/ЄДРПОУ/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Записати перевірку/ })).toBeInTheDocument();
  });
});

// AC-10: every value on the page names its source, one label per value.
const loaded = () => screen.findByText(uk.applications.filing.fullName, { selector: 'dt' });
const rowOf = (label: string) => screen.getByText(label, { selector: 'dt' }).parentElement as HTMLElement;
const termLabel = 'Строк повноважень';
const withTerm: ApplicationReview = {
  ...base,
  filings: [{ ...base.filings[0]!, term_start: '2026-01-01', term_end: '2027-12-31' }],
};
const typedRows: [string, string, string | RegExp][] = [
  ['the head name', uk.applications.filing.fullName, 'Тестовий Голова Правління'],
  ['the association name', uk.applications.filing.name, 'ОСББ «Тестовий Будинок»'],
  ['the ЄДРПОУ', uk.applications.filing.edrpou, '12345678'],
  ['the address', uk.applications.filing.address, 'м. Київ, вул. Тестова, 1'],
  ['the position', uk.applications.filing.positionCaption, uk.applications.filing.positionHead],
];
const sourceAt = (text: string, date: string) => `${text} ${formatDate(date)}`;
const checkedAt = '2026-11-11T08:05:00Z';

describe('ReviewPage source of every value (AC-10)', () => {
  it.each(typedRows)('labels %s as entered by the applicant', async (_name, label, value) => {
    review = withTerm;
    await renderInRouter(<ReviewPage applicationId={appId} />);
    await loaded();
    const row = rowOf(label);
    expect(within(row).getByText(value)).toBeInTheDocument();
    expect(within(row).getByText(uk.operator.enteredByApplicant)).toBeInTheDocument();
  });

  it('labels the term as entered by the applicant with both of its dates', async () => {
    review = withTerm;
    await renderInRouter(<ReviewPage applicationId={appId} />);
    await loaded();
    const row = rowOf(termLabel);
    expect(within(row).getByText(/01\.01\.2026|1 січня 2026/)).toBeInTheDocument();
    expect(within(row).getByText(/31\.12\.2027|31 грудня 2027/)).toBeInTheDocument();
    expect(within(row).getByText(uk.operator.enteredByApplicant)).toBeInTheDocument();
  });

  it('labels a term without an end as entered by the applicant', async () => {
    await renderInRouter(<ReviewPage applicationId={appId} />);
    await loaded();
    const row = rowOf(termLabel);
    expect(within(row).getByText(/01\.01\.2026|1 січня 2026/)).toBeInTheDocument();
    expect(within(row).getByText(uk.operator.enteredByApplicant)).toBeInTheDocument();
  });

  const checkedRows: [string, string, string | RegExp, 'signature' | 'extract'][] = [
    ['the signer', uk.operator.signer, 'Тестовий Голова Правління', 'signature'],
    ['the signing date', uk.operator.signedOn, /^(08\.11\.2026|8 листопада 2026)$/, 'signature'],
    ['the head named in the extract', uk.operator.head, 'Тестовий Голова Правління', 'extract'],
    ['the ЄДРПОУ of the extract', uk.operator.edrpou, '12345678', 'extract'],
    ['the extract date', uk.operator.extractDatedOn, /^(05\.11\.2026|5 листопада 2026)$/, 'extract'],
  ];

  it.each(checkedRows)('labels %s with the source and the check date', async (_name, label, value, source) => {
    review = checked;
    await renderInRouter(<ReviewPage applicationId={appId} />);
    await loaded();
    const row = rowOf(label);
    expect(within(row).getByText(value)).toBeInTheDocument();
    const text = source === 'signature' ? uk.operator.fromSignature : uk.operator.fromExtract;
    expect(within(row).getByText(sourceAt(text, checkedAt))).toBeInTheDocument();
    expect(within(row).queryByText(uk.operator.notChecked)).not.toBeInTheDocument();
  });

  it.each(checkedRows)('marks %s as «не перевірено» before the check', async (_name, label) => {
    await renderInRouter(<ReviewPage applicationId={appId} />);
    await loaded();
    expect(within(rowOf(label)).getByText(uk.operator.notChecked)).toBeInTheDocument();
  });
});

// AC-04c, AC-19b: competing заявки and запити side by side.
const otherIds = ['00000000-0000-7000-8000-0000000000a2', '00000000-0000-7000-8000-0000000000a3'];
const competitor = (
  id: string,
  kind: 'connection' | 'transfer',
  applicant: string,
  filedAt: string,
): ApplicationReview => ({
  ...base,
  application: { ...application, id, kind, filed_at: filedAt, created_at: filedAt },
  filings: [{ ...base.filings[0]!, id: `${id}-filing`, applicant_full_name: applicant, created_at: filedAt }],
});
const linkTo = (id: string) =>
  screen.getAllByRole('link').find((a) => a.getAttribute('href')?.endsWith(`/operator/applications/${id}`)) as HTMLElement;

describe('ReviewPage competing applications (AC-04c, AC-19b)', () => {
  it('keeps the list markup valid while a competing заявка loads: only <li> children, the skeleton inside one', async () => {
    review = { ...base, competing_application_ids: otherIds };
    const answer = fetch as unknown as typeof fetch;
    vi.stubGlobal('fetch', vi.fn((input: RequestInfo | URL, init?: RequestInit) =>
      String(input).endsWith(`/identity/operator/applications/${otherIds[0]}`) ? new Promise<Response>(() => {}) : answer(input, init),
    ));
    await renderInRouter(<ReviewPage applicationId={appId} />);
    const list = (await screen.findByText(uk.operator.competingConnection)).parentElement!.querySelector('ul') as HTMLElement;
    await waitFor(() => expect(list.querySelector('[data-slot="skeleton"]')).not.toBeNull());
    expect(Array.from(list.children).map((c) => c.tagName)).toEqual(['LI', 'LI']);
    expect(list.querySelector('[data-slot="skeleton"]')!.closest('li')).not.toBeNull();
  });

  it('lists two competing заявки with applicant, filing date and a link to each', async () => {
    review = { ...base, competing_application_ids: otherIds };
    otherReviews = {
      [otherIds[0]!]: competitor(otherIds[0]!, 'connection', 'Конкурент Перший', '2026-11-12T09:00:00Z'),
      [otherIds[1]!]: competitor(otherIds[1]!, 'connection', 'Конкурент Другий', '2026-11-13T09:00:00Z'),
    };
    await renderInRouter(<ReviewPage applicationId={appId} />);
    expect(await screen.findByText('Конкурент Перший')).toBeInTheDocument();
    expect(screen.getByText('Конкурент Другий')).toBeInTheDocument();

    const first = linkTo(otherIds[0]!).closest('li') as HTMLElement;
    expect(within(first).getByText('Конкурент Перший')).toBeInTheDocument();
    expect(within(first).getByText(/12\.11\.2026|12 листопада 2026/)).toBeInTheDocument();
    const second = linkTo(otherIds[1]!).closest('li') as HTMLElement;
    expect(within(second).getByText('Конкурент Другий')).toBeInTheDocument();
    expect(within(second).getByText(/13\.11\.2026|13 листопада 2026/)).toBeInTheDocument();
    expect(screen.getByText(uk.operator.competingConnection)).toBeInTheDocument();
  });

  it('lists two competing запити with applicant, filing date and a link to each', async () => {
    review = { ...base, application: { ...application, kind: 'transfer' }, competing_application_ids: otherIds };
    otherReviews = {
      [otherIds[0]!]: competitor(otherIds[0]!, 'transfer', 'Претендент Перший', '2026-11-14T09:00:00Z'),
      [otherIds[1]!]: competitor(otherIds[1]!, 'transfer', 'Претендент Другий', '2026-11-15T09:00:00Z'),
    };
    await renderInRouter(<ReviewPage applicationId={appId} />);
    expect(await screen.findByText('Претендент Перший')).toBeInTheDocument();

    const first = linkTo(otherIds[0]!).closest('li') as HTMLElement;
    expect(within(first).getByText('Претендент Перший')).toBeInTheDocument();
    expect(within(first).getByText(/14\.11\.2026|14 листопада 2026/)).toBeInTheDocument();
    const second = linkTo(otherIds[1]!).closest('li') as HTMLElement;
    expect(within(second).getByText('Претендент Другий')).toBeInTheDocument();
    expect(within(second).getByText(/15\.11\.2026|15 листопада 2026/)).toBeInTheDocument();
    expect(screen.getByText(uk.operator.competingTransfer)).toBeInTheDocument();
  });

  it('lists no competitors for a заявка that has none', async () => {
    await renderInRouter(<ReviewPage applicationId={appId} />);
    await loaded();
    expect(screen.queryByText(uk.operator.competingConnection)).not.toBeInTheDocument();
    expect(screen.queryByRole('link', { name: /Конкурент/ })).not.toBeInTheDocument();
  });
});

// A refiled заявка as the server returns it: filings newest first, checks oldest first (review.sql).
const oldFiling = {
  ...base.filings[0]!,
  id: filingId,
  applicant_full_name: 'Попередній Заявник',
  extract: { ...base.filings[0]!.extract, id: 'e1', original_name: 'old.pdf' },
  created_at: '2026-11-10T09:00:00Z',
};
const newFiling = {
  ...base.filings[0]!,
  id: refiledId,
  applicant_full_name: 'Новий Заявник',
  extract: { ...base.filings[0]!.extract, id: 'e2', original_name: 'new.pdf' },
  created_at: '2026-11-12T09:00:00Z',
};
const checkOf = (id: string, filing: string, signer: string, checkedAt: string) => ({
  id,
  filing_id: filing,
  signature_valid: true,
  signer_full_name: signer,
  signed_on: '2026-11-08',
  head_full_name: 'Керівник за витягом',
  extract_edrpou: '12345678',
  extract_dated_on: '2026-11-05',
  checked_at: checkedAt,
});
const refiled: ApplicationReview = {
  ...base,
  filings: [newFiling, oldFiling],
  checks: [
    checkOf('c1', filingId, 'Підписант першої заяви', '2026-11-11T08:05:00Z'),
    checkOf('c2', refiledId, 'Перший підписант нової заяви', '2026-11-13T09:00:00Z'),
    checkOf('c3', refiledId, 'Останній підписант нової заяви', '2026-11-14T10:00:00Z'),
  ],
};

describe('ReviewPage after a refiling', () => {
  beforeEach(() => {
    review = refiled;
  });

  it('presents the newest filing with the latest check of that filing (AC-06, AC-15)', async () => {
    await renderInRouter(<ReviewPage applicationId={appId} />);
    expect(await screen.findByText('Новий Заявник')).toBeInTheDocument();
    const signer = rowOf(uk.operator.signer);
    expect(within(signer).getByText('Останній підписант нової заяви')).toBeInTheDocument();
    expect(within(signer).getByText(sourceAt(uk.operator.fromSignature, '2026-11-14T10:00:00Z'))).toBeInTheDocument();
    expect(screen.queryByText('Перший підписант нової заяви')).not.toBeInTheDocument();
  });

  it('lists the earlier filing as history and offers the check form once (AC-06)', async () => {
    await renderInRouter(<ReviewPage applicationId={appId} />);
    expect(await screen.findByText('Новий Заявник')).toBeInTheDocument();
    expect(screen.getByText('Попередній Заявник')).toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: /Записати перевірку/ })).toHaveLength(1);
    expect(screen.getAllByRole('button', { name: 'Схвалити' })).toHaveLength(1);
  });

  it('links the extract of each filing through the generated download URL (AC-03, AC-08)', async () => {
    await renderInRouter(<ReviewPage applicationId={appId} />);
    await screen.findByText('Новий Заявник');
    const hrefs = screen.getAllByRole('link', { name: /витяг/i }).map((a) => a.getAttribute('href'));
    expect(hrefs).toHaveLength(2);
    expect(hrefs).toContain(getDownloadExtractForReviewUrl(appId, refiledId));
    expect(hrefs).toContain(getDownloadExtractForReviewUrl(appId, filingId));
  });

  it('links the extract of a заявка with a single filing (AC-03)', async () => {
    review = base;
    await renderInRouter(<ReviewPage applicationId={appId} />);
    await screen.findAllByText(/введено заявником/);
    const links = screen.getAllByRole('link', { name: /витяг/i });
    expect(links).toHaveLength(1);
    expect(links[0]).toHaveAttribute('href', getDownloadExtractForReviewUrl(appId, filingId));
  });

  it('records the check against the newest filing (AC-03, AC-15)', async () => {
    await renderInRouter(<ReviewPage applicationId={appId} />);
    await screen.findByText('Новий Заявник');
    fireEvent.change(screen.getByLabelText(/Підписант/), { target: { value: 'Підписант' } });
    fireEvent.change(screen.getByLabelText(/Дата підпису/), { target: { value: '2026-11-08' } });
    fireEvent.change(screen.getByLabelText(/Керівник/), { target: { value: 'Керівник' } });
    fireEvent.change(screen.getByLabelText(/ЄДРПОУ/), { target: { value: '12345678' } });
    fireEvent.change(screen.getByLabelText(/Дата витягу/), { target: { value: '2026-11-05' } });
    fireEvent.click(screen.getByRole('button', { name: /Записати перевірку/ }));
    await waitFor(() => expect(sent).toHaveLength(1));
    expect(sent[0]?.path).toMatch(/\/checks$/);
    expect(sent[0]?.body).toMatchObject({ filing_id: refiledId });
  });

  it('approves the newest filing (AC-08)', async () => {
    await renderInRouter(<ReviewPage applicationId={appId} />);
    await screen.findByText('Новий Заявник');
    fireEvent.click(screen.getByRole('button', { name: 'Схвалити' }));
    await waitFor(() => expect(sent).toHaveLength(1));
    expect(sent[0]?.path).toMatch(/\/approval$/);
    expect(sent[0]?.body).toMatchObject({ filing_id: refiledId });
  });

  it('rejects the newest filing (AC-08)', async () => {
    await renderInRouter(<ReviewPage applicationId={appId} />);
    await screen.findByText('Новий Заявник');
    fireEvent.click(screen.getByRole('button', { name: 'Відхилити' }));
    const dialog = await screen.findByRole('dialog');
    fireEvent.change(within(dialog).getByLabelText(/Причина/), { target: { value: 'Витяг нечитабельний' } });
    fireEvent.click(within(dialog).getByRole('button', { name: /Відхилити/ }));
    await waitFor(() => expect(sent).toHaveLength(1));
    expect(sent[0]?.path).toMatch(/\/rejection$/);
    expect(sent[0]?.body).toMatchObject({ filing_id: refiledId });
  });
});

describe('OperatorQueue', () => {
  it('lists the applications for the оператор (AC-16)', async () => {
    await renderInRouter(<OperatorQueue />);
    expect(await screen.findByText(/Тестовий Будинок/)).toBeInTheDocument();
  });

  it('shows no queue to a person who is not the оператор, and leaves the refusal to the section guard (AC-16)', async () => {
    queueStatus = 403;
    await renderInRouter(<OperatorQueue />);
    await waitFor(() => expect(urls.some((path) => path.endsWith('/identity/operator/applications'))).toBe(true));
    expect(screen.queryByText(/Тестовий Будинок/)).not.toBeInTheDocument();
    expect(screen.queryByText('Немає доступу')).not.toBeInTheDocument();
  });

  // The server marks every item over the whole queue (AC-04c, AC-19b); the page never works it out from its rows.
  const queueItem = (id: string, name: string, kind: 'connection' | 'transfer', competing: boolean, state = 'pending') =>
    ({ ...application, id, association_name: name, kind, state, competing }) as QueueItem;
  const queueRow = (name: string) => screen.getByRole('link', { name }).closest('li') as HTMLElement;

  it('marks the rows the server marked as competing заявки, and no other row (AC-04c)', async () => {
    queueItems = [
      queueItem(appId, 'ОСББ Перше', 'connection', true),
      queueItem(otherIds[0]!, 'ОСББ Друге', 'connection', true),
      queueItem(otherIds[1]!, 'ОСББ Окреме', 'connection', false),
    ];
    await renderInRouter(<OperatorQueue />);
    await screen.findByRole('link', { name: 'ОСББ Перше' });
    expect(within(queueRow('ОСББ Перше')).getByText(uk.operator.competingConnection)).toBeInTheDocument();
    expect(within(queueRow('ОСББ Друге')).getByText(uk.operator.competingConnection)).toBeInTheDocument();
    expect(within(queueRow('ОСББ Окреме')).queryByText(uk.operator.competingConnection)).not.toBeInTheDocument();
  });

  it('marks a row whose competitor is not on the page, because the server said so (AC-04c)', async () => {
    queueItems = [queueItem(appId, 'ОСББ Перше', 'connection', true)];
    await renderInRouter(<OperatorQueue />);
    await screen.findByRole('link', { name: 'ОСББ Перше' });
    expect(within(queueRow('ОСББ Перше')).getByText(uk.operator.competingConnection)).toBeInTheDocument();
  });

  it('does not mark a row the server did not mark, whatever else the page holds for its ЄДРПОУ (AC-04c)', async () => {
    queueItems = [
      queueItem(appId, 'ОСББ Перше', 'connection', false),
      queueItem(otherIds[0]!, 'ОСББ Друге', 'connection', false),
    ];
    await renderInRouter(<OperatorQueue />);
    await screen.findByRole('link', { name: 'ОСББ Перше' });
    expect(within(queueRow('ОСББ Перше')).queryByText(uk.operator.competingConnection)).not.toBeInTheDocument();
  });

  it('marks both rows of two competing запити as запити (AC-19b)', async () => {
    queueItems = [queueItem(appId, 'ОСББ Перше', 'transfer', true), queueItem(otherIds[0]!, 'ОСББ Друге', 'transfer', true)];
    await renderInRouter(<OperatorQueue />);
    await screen.findByRole('link', { name: 'ОСББ Перше' });
    expect(within(queueRow('ОСББ Перше')).getByText(uk.operator.competingTransfer)).toBeInTheDocument();
    expect(within(queueRow('ОСББ Друге')).getByText(uk.operator.competingTransfer)).toBeInTheDocument();
  });

  it('shows the fallback with a retry that reads the queue again when the read fails (not a blank page)', async () => {
    queueStatus = 500;
    await renderInRouter(<OperatorQueue />);
    expect(await screen.findByText(uk.association.requestFailed)).toBeInTheDocument();
    const reads = () => urls.filter((path) => path.endsWith('/identity/operator/applications')).length;
    const before = reads();
    queueStatus = 200;
    fireEvent.click(screen.getByRole('button', { name: uk.association.retry }));
    expect(await screen.findByRole('link', { name: 'ОСББ «Тестовий Будинок»' })).toBeInTheDocument();
    expect(reads()).toBeGreaterThan(before);
  });
});
