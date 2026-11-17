import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createMemoryHistory, createRouter, RouterProvider } from '@tanstack/react-router';
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { getAddCaseDocumentUrl, getDownloadFileUrl, getGetCaseUrl, getListCasesUrl } from '@/api/gen/endpoints/governance/governance';
import { formatDate } from '@/i18n/uk';
import { routeTree } from '@/routeTree.gen';

// The справи list and the case screen through the real route tree, on the generated client
// (T80, AC-25, AC-28; board-protocol T36 scope): the list from listCases, the case with its document
// versions from getCase, an upload through addCaseDocument, the refusals shown inline, and no way to change
// anything in a вимкнене ОСББ. The fake answers like the server (problem+json, codes of the contract).
//
// Contract for the implementer (labels are matched by regex, so the exact copy is free):
// - /o/<id>/cases lists every справа as a link named by its title to /o/<id>/cases/<case id>, with its created date;
// - /o/<id>/cases/<case id> (a new route, cases.$caseId.index.tsx) has an h1 with the title, a link back to the
//   list, a link to the protocol draft /o/<id>/cases/<case id>/protocol, and the documents with their versions;
// - every document links its file through the generated getDownloadFileUrl;
// - the upload is a file input labelled «Файл…» (FileField) and a button named «Додати|Завантажити|Прикріпити»
//   that calls addCaseDocument; the new document shows after the case is read again;
// - a file over 20 MB or not PDF/JPEG/PNG is refused in the browser (no request) and by the server
//   (governance.file_invalid); both are an inline role=alert naming the rule («PDF», «20 МБ»);
// - governance.drive_not_connected renders AssociationProblemNotice (storage message, who can connect, link to the card);
// - identity.association_disabled on a change renders «Доступ призупинено: <причина>» and nothing is attached.

const lypky = '00000000-0000-7000-8000-000000000001';
const principal = { account_id: '00000000-0000-7000-8000-00000000a001', email: 'user-1@example.test', is_operator: false };
const memberId = '00000000-0000-7000-8000-00000000b001';
const reason = 'спір щодо складу правління';
const roofId = '00000000-0000-7000-8000-000000000c01';
const liftId = '00000000-0000-7000-8000-000000000c02';
const roofTitle = 'Гідроізоляція покрівлі, будинок 1';
const liftTitle = 'Ремонт сходових клітин';

const fileRef = (id: string, name: string) => ({
  id,
  original_name: name,
  media_type: 'application/pdf',
  size_bytes: 482113,
  sha256: '9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08',
  uploaded_by_member_id: memberId,
  created_at: '2026-10-20T09:10:00Z',
});
const estimateV1 = '00000000-0000-7000-8000-000000000d01';
const estimateV2 = '00000000-0000-7000-8000-000000000d02';
const newDoc = '00000000-0000-7000-8000-000000000d03';
const doc = (id: string, fileId: string, name: string, previous: string | null, date: string) => ({
  id,
  case_id: roofId,
  file: fileRef(fileId, name),
  previous_version_id: previous,
  document_date: date,
  created_by_member_id: memberId,
  created_at: '2026-10-20T09:10:00Z',
});
const v1File = '00000000-0000-7000-8000-000000000701';
const v2File = '00000000-0000-7000-8000-000000000702';
const newFile = '00000000-0000-7000-8000-000000000703';

const cases = [
  { id: roofId, title: roofTitle, created_by_member_id: memberId, created_at: '2026-10-20T09:00:00Z' },
  { id: liftId, title: liftTitle, created_by_member_id: memberId, created_at: '2026-10-09T09:00:00Z' },
];

let disabledFromStart: boolean;
let documents: ReturnType<typeof doc>[];
let uploadAnswer: () => Response;
let calls: string[];
let uploads: Array<{ path: string; file: File | null; previous: string | null }>;

const json = (status: number, body: unknown) =>
  new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': status >= 400 ? 'application/problem+json' : 'application/json' },
  });
const problem = (status: number, code: string, detail: string, errors?: unknown[]) =>
  json(status, { type: 'about:blank', title: 'Problem', status, code, detail, ...(errors ? { errors } : {}) });

function fakeApi(input: RequestInfo | URL, init?: RequestInit) {
  const url = new URL(String(input), 'http://localhost');
  const path = url.pathname;
  calls.push(`${init?.method ?? 'GET'} ${path}${url.search}`);
  if (path.endsWith('/identity/session')) return Promise.resolve(json(200, principal));
  if (path.endsWith('/identity/memberships')) {
    return Promise.resolve(
      json(200, {
        items: [{ association_id: lypky, association_name: 'ОСББ «Липки»', association_state: disabledFromStart ? 'disabled' : 'connected', board_member_id: memberId, position: 'member' }],
        next_cursor: null,
      }),
    );
  }
  if (path.includes('/identity/associations/')) {
    if (disabledFromStart) return Promise.resolve(problem(403, 'identity.association_disabled', reason));
    return Promise.resolve(json(404, {}));
  }
  if (path.includes('/governance/associations/')) {
    // A вимкнене ОСББ refuses every scoped read and change outside the registry (ADR-0007).
    if (disabledFromStart) return Promise.resolve(problem(403, 'identity.association_disabled', reason));
    if (init?.method === 'POST' && path === getAddCaseDocumentUrl(lypky, roofId)) {
      const body = init.body;
      const form = body instanceof FormData ? body : new FormData();
      const file = form.get('file');
      uploads.push({ path, file: file instanceof Blob ? (file as File) : null, previous: (form.get('previous_version_id') as string | null) ?? null });
      return Promise.resolve(uploadAnswer());
    }
    if (path === getListCasesUrl(lypky)) return Promise.resolve(json(200, { items: cases, next_cursor: null }));
    if (path === getGetCaseUrl(lypky, roofId)) return Promise.resolve(json(200, { ...cases[0], documents }));
    return Promise.resolve(problem(404, 'governance.not_found', 'not found'));
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

const listPath = `/o/${lypky}/cases`;
const casePath = `/o/${lypky}/cases/${roofId}`;
const protocolPath = `/o/${lypky}/cases/c-101/protocol`;
const main = () => screen.getByRole('main');
const settle = (ms = 600) => new Promise((resolve) => setTimeout(resolve, ms));
const alertWith = (text: RegExp) =>
  waitFor(() => {
    const found = screen.getAllByRole('alert').find((a) => text.test(a.textContent ?? ''));
    expect(found, `an alert matching ${text}`).toBeDefined();
    return found as HTMLElement;
  });

const pdf = (name = 'kosztorys.pdf') => new File(['%PDF-1.7 test'], name, { type: 'application/pdf' });
const attach = (file: File) => fireEvent.change(screen.getByLabelText(/Файл/), { target: { files: [file] } });
const send = () => fireEvent.click(screen.getByRole('button', { name: /Додати|Завантажити|Прикріпити/ }));
const sendIfOffered = () => {
  if (screen.queryByRole('button', { name: /Додати|Завантажити|Прикріпити/ })) send();
};
const openCase = async () => {
  open(casePath);
  await screen.findByRole('heading', { level: 1, name: roofTitle });
};

beforeEach(() => {
  disabledFromStart = false;
  calls = [];
  uploads = [];
  documents = [doc(estimateV1, v1File, 'kosztorys-v1.pdf', null, '2026-10-01'), doc(estimateV2, v2File, 'kosztorys-v2.pdf', estimateV1, '2026-10-18')];
  uploadAnswer = () => {
    documents = [...documents, doc(newDoc, newFile, 'kosztorys.pdf', null, '2026-10-19')];
    return json(201, documents.at(-1));
  };
  vi.stubGlobal('fetch', vi.fn(fakeApi));
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe('the справи list on the API (AC-25, F15)', () => {
  it('lists the справи read through listCases, each opening its case with the created date', async () => {
    open(listPath);
    expect(await screen.findByRole('heading', { level: 1, name: /Справи/ })).toBeInTheDocument();

    const roof = await screen.findByRole('link', { name: new RegExp(roofTitle) });
    expect(roof.getAttribute('href')).toBe(casePath);
    expect(screen.getByRole('link', { name: new RegExp(liftTitle) }).getAttribute('href')).toBe(`/o/${lypky}/cases/${liftId}`);
    expect(main()).toHaveTextContent(formatDate('2026-10-20T09:00:00Z'));
    expect(calls).toContain(`GET ${getListCasesUrl(lypky)}`);
  });

  it('shows nothing of the in-memory mock', async () => {
    open(listPath);
    await screen.findByRole('link', { name: new RegExp(roofTitle) });
    expect(screen.queryByText(/під’їзд 2/)).toBeNull();
    expect(screen.queryByText(/Щорічний огляд ліфта/)).toBeNull();
    expect(main().textContent).not.toMatch(/Комерційна пропозиція|184\s?500/);
  });
});

describe('the case screen with its documents (AC-09, T36)', () => {
  it('shows the title, the documents read through getCase and the way back', async () => {
    await openCase();
    expect(calls).toContain(`GET ${getGetCaseUrl(lypky, roofId)}`);
    expect(await screen.findByText('kosztorys-v1.pdf')).toBeInTheDocument();
    expect(screen.getByText('kosztorys-v2.pdf')).toBeInTheDocument();
    expect(main()).toHaveTextContent(formatDate('2026-10-18'));

    const hrefs = within(main()).getAllByRole('link').map((a) => a.getAttribute('href'));
    expect(hrefs).toContain(listPath);
    expect(hrefs).toContain(`${casePath}/protocol`);
  });

  it('links every document through the generated download URL and marks the newer version', async () => {
    await openCase();
    await screen.findByText('kosztorys-v2.pdf');

    const links = within(main()).getAllByRole('link');
    for (const [id, name] of [
      [v1File, 'kosztorys-v1.pdf'],
      [v2File, 'kosztorys-v2.pdf'],
    ] as const) {
      const link = links.find((a) => a.getAttribute('href') === getDownloadFileUrl(lypky, id));
      expect(link, `a link to ${name}`).toBeDefined();
      expect(link).toHaveAccessibleName(new RegExp(`Завантажити|${name.replace('.', '\\.')}`));
    }
    const row = screen.getByText('kosztorys-v2.pdf').closest<HTMLElement>('li, tr, article, [role="row"], [role="listitem"]');
    expect(row, 'the document row').not.toBeNull();
    expect(row?.textContent).toMatch(/версі/i);
  });
});

describe('uploading a document (AC-09, AC-28)', () => {
  it('sends the chosen file through addCaseDocument and shows the new document', async () => {
    await openCase();
    await screen.findByText('kosztorys-v2.pdf');

    attach(pdf());
    send();

    expect(await screen.findByText('kosztorys.pdf')).toBeInTheDocument();
    expect(uploads).toHaveLength(1);
    expect(uploads[0]?.path).toBe(getAddCaseDocumentUrl(lypky, roofId));
    expect(uploads[0]?.file?.name).toBe('kosztorys.pdf');
    expect(screen.queryAllByRole('alert')).toEqual([]);
  });

  it('refuses a file over 20 MB in the browser, names the rule and sends nothing', async () => {
    await openCase();
    await screen.findByText('kosztorys-v2.pdf');

    const big = pdf('big.pdf');
    Object.defineProperty(big, 'size', { value: 20 * 1024 * 1024 + 1 });
    attach(big);
    sendIfOffered();

    await alertWith(/20\s?МБ/);
    expect(uploads).toEqual([]);
    expect(screen.queryByText('big.pdf')).toBeNull();
  });

  it('refuses a file that is not PDF, JPEG or PNG in the browser and sends nothing', async () => {
    await openCase();
    await screen.findByText('kosztorys-v2.pdf');

    attach(new File(['notes'], 'notes.txt', { type: 'text/plain' }));
    sendIfOffered();

    await alertWith(/PDF/);
    expect(uploads).toEqual([]);
  });

  it('shows governance.file_invalid inline and attaches nothing', async () => {
    uploadAnswer = () => problem(422, 'governance.file_invalid', 'Файл має бути PDF, JPEG або PNG і не більший за 20 МБ');
    await openCase();
    await screen.findByText('kosztorys-v2.pdf');

    attach(pdf());
    send();

    const alert = await alertWith(/PDF/);
    expect(alert.textContent).toMatch(/20/);
    expect(uploads).toHaveLength(1);
    expect(screen.queryByText('kosztorys.pdf')).toBeNull();
    // The documents already there stay, and the member can choose another file.
    expect(screen.getByText('kosztorys-v2.pdf')).toBeInTheDocument();
    expect(screen.getByLabelText(/Файл/)).toBeEnabled();
  });

  it('says to connect the storage first, that any член правління can, and links the card on governance.drive_not_connected (AC-28)', async () => {
    uploadAnswer = () => problem(409, 'governance.drive_not_connected', 'Спершу підключіть сховище документів ОСББ');
    await openCase();
    await screen.findByText('kosztorys-v2.pdf');

    attach(pdf());
    send();

    const alert = await alertWith(/Спершу підключіть сховище документів ОСББ/);
    expect(alert.textContent).toMatch(/будь-який член правління/);
    expect(within(alert).getByRole('link').getAttribute('href')).toBe(`/o/${lypky}/association`);
    expect(screen.queryByText('kosztorys.pdf')).toBeNull();
  });

  it('answers «Доступ призупинено» with the reason when the association was disabled meanwhile, and attaches nothing (AC-25)', async () => {
    uploadAnswer = () => problem(403, 'identity.association_disabled', reason);
    await openCase();
    await screen.findByText('kosztorys-v2.pdf');

    attach(pdf());
    send();

    await alertWith(new RegExp(`Доступ призупинено: ${reason}`));
    expect(uploads).toHaveLength(1);
    expect(screen.queryByText('kosztorys.pdf')).toBeNull();
  });
});

describe('справи in a вимкнене ОСББ (AC-25)', () => {
  beforeEach(() => {
    disabledFromStart = true;
  });

  it('shows the suspended notice with the reason, no справи and no load error on the list', async () => {
    open(listPath);
    await alertWith(new RegExp(`Доступ призупинено: ${reason}`));
    await settle();
    expect(screen.queryByRole('link', { name: new RegExp(roofTitle) })).toBeNull();
    expect(screen.queryByText(/Не вдалося завантажити справи/)).toBeNull();
    expect(screen.queryAllByText(/під’їзд 2/)).toEqual([]);
  });

  it('offers no upload on the case screen', async () => {
    open(casePath);
    await alertWith(new RegExp(`Доступ призупинено: ${reason}`));
    await settle();
    // The screen itself is there (a way back to the list), only the change controls are not.
    const hrefs = within(main()).getAllByRole('link').map((a) => a.getAttribute('href'));
    expect(hrefs).toContain(listPath);
    expect(screen.queryByLabelText(/Файл/)).toBeNull();
    expect(screen.queryByRole('button', { name: /Додати|Завантажити|Прикріпити/ })).toBeNull();
    expect(uploads).toEqual([]);
  });

  it('does not open the protocol draft: no mock draft, no slots, no print', async () => {
    open(protocolPath);
    await alertWith(new RegExp(`Доступ призупинено: ${reason}`));
    await settle();
    expect(screen.queryAllByText(/Гідроізоляція покрівлі, під’їзд 2/)).toEqual([]);
    expect(screen.queryByRole('button', { name: /Надрукувати протокол/ })).toBeNull();
    expect(within(main()).queryAllByRole('textbox')).toEqual([]);
  });
});

describe('the protocol draft screen keeps its mock (T80)', () => {
  it('says on screen that the draft is a mock until board-protocol builds drafts', async () => {
    open(protocolPath);
    await screen.findByRole('heading', { level: 1, name: /Гідроізоляція покрівлі, під’їзд 2/ });
    expect(main().textContent).toMatch(/макет|демонстраційн/i);
  });
});
