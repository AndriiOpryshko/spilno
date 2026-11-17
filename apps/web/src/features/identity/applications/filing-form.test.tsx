import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createMemoryHistory, createRootRoute, createRouter, RouterProvider } from '@tanstack/react-router';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { FilingForm } from './filing-form';

type Kind = 'connection' | 'transfer';

const problem = (status: number, code: string, detail: string, errors?: { pointer: string; detail: string }[]) =>
  new Response(JSON.stringify({ type: 'about:blank', title: code, status, code, detail, errors }), {
    status,
    headers: { 'Content-Type': 'application/problem+json' },
  });

const filed = (kind: Kind) =>
  new Response(
    JSON.stringify({
      id: '00000000-0000-7000-8000-0000000000a1',
      kind,
      edrpou: '12345678',
      association_id: null,
      association_name: 'ОСББ «Тест»',
      state: 'pending',
      filed_at: '2026-11-10',
      rejection_reason: null,
      created_at: '2026-11-10T10:00:00Z',
    }),
    { status: 201, headers: { 'Content-Type': 'application/json' } },
  );

// The API may still answer 501: the form is tested against a stubbed fetch.
function stubApi(response: () => Response) {
  const fetchMock = vi.fn(async () => response());
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}

async function renderForm(kind: Kind) {
  const router = createRouter({
    routeTree: createRootRoute({
      component: () => (
        <QueryClientProvider client={new QueryClient({ defaultOptions: { mutations: { retry: false } } })}>
          <FilingForm kind={kind} />
        </QueryClientProvider>
      ),
    }),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  });
  render(<RouterProvider router={router} />);
  await screen.findByRole('button', { name: /Подати/ });
}

function fillValid() {
  fireEvent.change(screen.getByLabelText(/ЄДРПОУ/), { target: { value: '12345678' } });
  fireEvent.change(screen.getByLabelText(/Назва ОСББ/), { target: { value: 'ОСББ «Тест»' } });
  fireEvent.change(screen.getByLabelText(/Адреса/), { target: { value: 'Київ, вул. Тестова, 1' } });
  fireEvent.change(screen.getByLabelText(/ПІБ/), { target: { value: 'Тестовий Голова' } });
  fireEvent.change(screen.getByLabelText(/Початок строку/), { target: { value: '2025-01-01' } });
  const file = new File(['%PDF-1.7'], 'extract.pdf', { type: 'application/pdf' });
  fireEvent.change(screen.getByLabelText(/Витяг/), { target: { files: [file] } });
}

const submit = () => fireEvent.click(screen.getByRole('button', { name: /Подати/ }));

afterEach(() => vi.unstubAllGlobals());

describe('FilingForm', () => {
  it('shows the position as fixed text, with no control to change it (AC-01)', async () => {
    await renderForm('connection');
    expect(screen.getByText('голова правління')).toBeInTheDocument();
    expect(screen.queryByRole('combobox')).not.toBeInTheDocument();
    expect(screen.queryByRole('radio')).not.toBeInTheDocument();
    expect(screen.queryByLabelText(/Посада/)).not.toBeInTheDocument();
  });

  it('files without a term end and sends the position head (AC-01)', async () => {
    const fetchMock = stubApi(() => filed('connection'));
    await renderForm('connection');
    fillValid();
    submit();
    expect(await screen.findByText('Заявку подано, очікує розгляду оператором')).toBeInTheDocument();
    const calls = fetchMock.mock.calls as unknown as [string, RequestInit][];
    const body = calls[0]![1].body as FormData;
    expect(body.get('kind')).toBe('connection');
    expect(body.get('position')).toBe('head');
    expect(body.get('term_end')).toBeNull();
  });

  it('shows the success message of a запит на передачу (AC-17)', async () => {
    const fetchMock = stubApi(() => filed('transfer'));
    await renderForm('transfer');
    fillValid();
    submit();
    expect(await screen.findByText('Запит на передачу подано, очікує розгляду оператором')).toBeInTheDocument();
    const calls = fetchMock.mock.calls as unknown as [string, RequestInit][];
    expect((calls[0]![1].body as FormData).get('kind')).toBe('transfer');
  });

  // The payload the backend sends (applications.go fileProblem, domain/filing.go): English detail, pointers per field.
  const invalidFiling = () =>
    problem(422, 'identity.application_invalid', 'The application is not valid.', [
      { pointer: '/edrpou', detail: 'must be 8 digits' },
      { pointer: '/association_name', detail: 'must be 1 to 200 characters' },
      { pointer: '/association_address', detail: 'must be 1 to 500 characters' },
      { pointer: '/applicant_full_name', detail: 'must be 1 to 200 characters' },
      { pointer: '/term_start', detail: 'must be a date not later than the filing date' },
      { pointer: '/extract', detail: 'is required' },
    ]);

  const fieldLabels = [/ЄДРПОУ/, /Назва ОСББ/, /Адреса/, /ПІБ/, /Початок строку/, /Витяг/];

  const messageOf = (label: RegExp) => {
    const field = screen.getByLabelText(label);
    const id = field.getAttribute('aria-describedby');
    expect(id).toBeTruthy();
    return document.getElementById(id as string)?.textContent ?? '';
  };

  it('writes a Ukrainian message next to every broken field after one submit (AC-02)', async () => {
    stubApi(invalidFiling);
    await renderForm('connection');
    submit();
    await waitFor(() => {
      for (const label of fieldLabels) expect(screen.getByLabelText(label)).toHaveAttribute('aria-invalid', 'true');
    });
    const messages = fieldLabels.map(messageOf);
    for (const m of messages) {
      expect(m).toMatch(/[А-Яа-яІіЇїЄєҐґ]/);
      // A file format keeps its usual name; nothing else is in Latin letters.
      expect(m.replace(/\bPDF\b/g, '')).not.toMatch(/[A-Za-z]/);
    }
    expect(messages[0]).toMatch(/8/);
    expect(new Set(messages).size).toBeGreaterThan(2);
    expect(screen.queryByText(/must be|is required|not valid/)).not.toBeInTheDocument();
  });

  it('shows the «Перевірте поля заявки: …» summary naming the broken fields, not the server text (AC-02, PRD 5.1)', async () => {
    stubApi(() =>
      problem(422, 'identity.application_invalid', 'The application is not valid.', [
        { pointer: '/edrpou', detail: 'must be 8 digits' },
        { pointer: '/term_start', detail: 'must be a date not later than the filing date' },
        { pointer: '/extract', detail: 'is required' },
      ]),
    );
    await renderForm('connection');
    submit();
    const summary = await screen.findByRole('alert');
    expect(summary.textContent).toMatch(/^Перевірте поля заявки: /);
    expect(summary.textContent).toMatch(/ЄДРПОУ/);
    expect(summary.textContent).toMatch(/Початок строку/);
    expect(summary.textContent).toMatch(/Витяг/);
    expect(summary.textContent).not.toMatch(/not valid/);
    expect(screen.getByLabelText(/Назва ОСББ/)).not.toHaveAttribute('aria-invalid', 'true');
    expect(screen.getByLabelText(/Адреса/)).not.toHaveAttribute('aria-invalid', 'true');
  });

  it('marks the extract on the client when no file is chosen, though the server does not report it (AC-02)', async () => {
    // The server leaves /extract out when the multipart file is empty alongside other errors.
    stubApi(() =>
      problem(422, 'identity.application_invalid', 'The application is not valid.', [
        { pointer: '/edrpou', detail: 'must be 8 digits' },
        { pointer: '/association_name', detail: 'must be 1 to 200 characters' },
        { pointer: '/association_address', detail: 'must be 1 to 500 characters' },
        { pointer: '/applicant_full_name', detail: 'must be 1 to 200 characters' },
        { pointer: '/term_start', detail: 'must be a date not later than the filing date' },
      ]),
    );
    await renderForm('connection');
    submit();
    const summary = await screen.findByRole('alert');
    await waitFor(() => expect(screen.getByLabelText(/Витяг/)).toHaveAttribute('aria-invalid', 'true'));
    const message = document.getElementById('filing-extract-error');
    expect(message).not.toBeNull();
    expect(message?.textContent).toMatch(/[А-Яа-яІіЇїЄєҐґ]/);
    expect(screen.getByLabelText(/Витяг/)).toHaveAttribute('aria-describedby', 'filing-extract-error');
    expect(summary.textContent).toMatch(/Витяг/);
  });

  it('does not mark the extract when a file is chosen and the server does not report it (AC-02)', async () => {
    stubApi(() =>
      problem(422, 'identity.application_invalid', 'The application is not valid.', [
        { pointer: '/edrpou', detail: 'must be 8 digits' },
      ]),
    );
    await renderForm('connection');
    fillValid();
    submit();
    await screen.findByRole('alert');
    expect(screen.getByLabelText(/Витяг/)).not.toHaveAttribute('aria-invalid', 'true');
    expect(document.getElementById('filing-extract-error')).toBeNull();
  });

  it('names a field once in the summary when the server sends two errors for it (AC-02)', async () => {
    stubApi(() =>
      problem(422, 'identity.application_invalid', 'The application is not valid.', [
        { pointer: '/edrpou', detail: 'must be 8 digits' },
        { pointer: '/edrpou', detail: 'must not be empty' },
        { pointer: '/term_start', detail: 'must be a date not later than the filing date' },
      ]),
    );
    await renderForm('connection');
    fillValid();
    submit();
    const summary = await screen.findByRole('alert');
    expect(summary.textContent?.match(/ЄДРПОУ/g)).toHaveLength(1);
    expect(summary.textContent).toMatch(/Початок строку/);
  });

  it('leaves a pointer the form has no field for out of the summary, with no English key (AC-02)', async () => {
    // The shared request validator names /association_id and /kind.
    stubApi(() =>
      problem(422, 'identity.application_invalid', 'The application is not valid.', [
        { pointer: '/kind', detail: 'must be one of connection, transfer' },
        { pointer: '/association_id', detail: 'must be a UUID' },
        { pointer: '/edrpou', detail: 'must be 8 digits' },
      ]),
    );
    await renderForm('connection');
    fillValid();
    submit();
    const summary = await screen.findByRole('alert');
    expect(summary.textContent).toMatch(/^Перевірте поля заявки: ЄДРПОУ$/);
    expect(summary.textContent).not.toMatch(/kind|association_id/);
  });

  it.each([
    ['is required'],
    ['the extract file is larger than 5 MB'],
    ['the extract must be a PDF or a signed container'],
  ])('words the extract message so it is true for the server detail «%s» (AC-02)', async (detail) => {
    stubApi(() => problem(422, 'identity.application_invalid', 'The application is not valid.', [{ pointer: '/extract', detail }]));
    await renderForm('connection');
    fillValid();
    submit();
    await screen.findByRole('alert');
    const message = document.getElementById('filing-extract-error')?.textContent ?? '';
    // Missing, too large (5 MB) and wrong type (PDF or a signed container) all get the same text.
    expect(message).toMatch(/5\s*(МБ|Мб|MB)/);
    expect(message).toMatch(/PDF/);
  });

  const chooseExtract = (file: File) => fireEvent.change(screen.getByLabelText(/Витяг/), { target: { files: [file] } });

  const expectExtractRefused = () => {
    expect(screen.getByLabelText(/Витяг/)).toHaveAttribute('aria-invalid', 'true');
    const message = document.getElementById('filing-extract-error')?.textContent ?? '';
    expect(message).toMatch(/5\s*МБ/);
    expect(message).toMatch(/PDF/);
    expect(screen.queryByText(/Не вдалося подати/)).not.toBeInTheDocument();
  };

  it('refuses an extract over 5 MB under the field, without a request (AC-02)', async () => {
    const fetchMock = stubApi(() => filed('connection'));
    await renderForm('connection');
    fillValid();
    chooseExtract(new File([new Uint8Array(5 * 1024 * 1024 + 1)], 'big.pdf', { type: 'application/pdf' }));
    submit();
    await waitFor(expectExtractRefused);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it('refuses an extract of another type under the field, without a request (AC-02)', async () => {
    const fetchMock = stubApi(() => filed('connection'));
    await renderForm('connection');
    fillValid();
    chooseExtract(new File(['x'], 'extract.png', { type: 'image/png' }));
    submit();
    await waitFor(expectExtractRefused);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it('shows a 413 answer as the same extract field message (AC-02)', async () => {
    stubApi(() => problem(413, 'identity.payload_too_large', 'The request body is too large.'));
    await renderForm('connection');
    fillValid();
    submit();
    await waitFor(expectExtractRefused);
    expect(screen.queryByText(/too large/)).not.toBeInTheDocument();
  });

  it('shows «Не вдалося подати» for a code it does not know, not the English server detail', async () => {
    stubApi(() => problem(409, 'identity.something_new', 'The application cannot be filed.'));
    await renderForm('connection');
    fillValid();
    submit();
    expect(await screen.findByRole('alert')).toHaveTextContent('Не вдалося подати. Спробуйте ще раз');
    expect(screen.queryByText(/cannot be filed/)).not.toBeInTheDocument();
  });

  it('offers the запит form when the association is already connected (AC-04)', async () => {
    stubApi(() =>
      problem(409, 'identity.association_already_connected', 'The application cannot be filed.'),
    );
    await renderForm('connection');
    fillValid();
    submit();
    expect(
      await screen.findByText('Це ОСББ вже підключене — зверніться до голови вашого ОСББ або подайте запит на передачу'),
    ).toBeInTheDocument();
    const link = screen.getByRole('link', { name: /запит на передачу/i });
    expect(link.getAttribute('href')).toContain('/applications/new');
    expect(link.getAttribute('href')).toContain('kind=transfer');
  });

  it('shows the open заявка when one already exists (AC-04b)', async () => {
    stubApi(() => problem(409, 'identity.application_already_open', 'The application cannot be filed.'));
    await renderForm('connection');
    fillValid();
    submit();
    expect(await screen.findByText('У вас уже є заявка на це ОСББ')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /заявк/i }).getAttribute('href')).toBe('/applications');
  });

  it('offers the заявка form when the association is not connected (AC-18)', async () => {
    stubApi(() =>
      problem(409, 'identity.association_not_connected', 'The application cannot be filed.'),
    );
    await renderForm('transfer');
    fillValid();
    submit();
    expect(await screen.findByText('Це ОСББ ще не підключене — подайте заявку на підключення')).toBeInTheDocument();
    const link = screen.getByRole('link', { name: /заявку на підключення/i });
    expect(link.getAttribute('href')).toContain('kind=connection');
  });

  it('shows the refusal of an ended term, not the field summary or the English detail (AC-30)', async () => {
    // The payload the backend sends (applications.go fileProblem, domain/filing.go): the 409 code carries the
    // fields detail and the /term_end pointer.
    stubApi(() =>
      problem(409, 'identity.term_ended', 'The application is not valid.', [
        { pointer: '/term_end', detail: 'the term ended before the filing date' },
      ]),
    );
    await renderForm('connection');
    fillValid();
    fireEvent.change(screen.getByLabelText(/Кінець строку/), { target: { value: '2026-11-09' } });
    submit();
    const alert = await screen.findByRole('alert');
    expect(alert).toHaveTextContent('Строк повноважень голови вже завершився — подайте витяг після переобрання');
    expect(screen.getAllByRole('alert')).toHaveLength(1);
    expect(screen.queryByText(/Перевірте поля заявки/)).not.toBeInTheDocument();
    expect(screen.queryByText(/not valid|term ended/)).not.toBeInTheDocument();
    expect(screen.getByLabelText(/Кінець строку/)).not.toHaveAttribute('aria-invalid', 'true');
  });
});
