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

  it('marks every broken field after one submit (AC-02)', async () => {
    stubApi(() =>
      problem(422, 'identity.application_invalid', 'Перевірте поля заявки: edrpou, term_start, extract', [
        { pointer: '/edrpou', detail: 'must be 8 digits' },
        { pointer: '/term_start', detail: 'must not be later than the filing date' },
        { pointer: '/extract', detail: 'is required' },
      ]),
    );
    await renderForm('connection');
    submit();
    await waitFor(() => {
      expect(screen.getByLabelText(/ЄДРПОУ/)).toHaveAttribute('aria-invalid', 'true');
      expect(screen.getByLabelText(/Початок строку/)).toHaveAttribute('aria-invalid', 'true');
      expect(screen.getByLabelText(/Витяг/)).toHaveAttribute('aria-invalid', 'true');
    });
    expect(screen.getByText('must be 8 digits')).toBeInTheDocument();
    expect(screen.getByText('must not be later than the filing date')).toBeInTheDocument();
    expect(screen.getByText('is required')).toBeInTheDocument();
    expect(screen.getByLabelText(/Назва ОСББ/)).not.toHaveAttribute('aria-invalid', 'true');
    expect(screen.getByLabelText(/Адреса/)).not.toHaveAttribute('aria-invalid', 'true');
  });

  it('offers the запит form when the association is already connected (AC-04)', async () => {
    stubApi(() =>
      problem(
        409,
        'identity.association_already_connected',
        'Це ОСББ вже підключене — зверніться до голови вашого ОСББ або подайте запит на передачу',
      ),
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
    stubApi(() => problem(409, 'identity.application_already_open', 'У вас уже є заявка на це ОСББ'));
    await renderForm('connection');
    fillValid();
    submit();
    expect(await screen.findByText('У вас уже є заявка на це ОСББ')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /заявк/i }).getAttribute('href')).toBe('/applications');
  });

  it('offers the заявка form when the association is not connected (AC-18)', async () => {
    stubApi(() =>
      problem(409, 'identity.association_not_connected', 'Це ОСББ ще не підключене — подайте заявку на підключення'),
    );
    await renderForm('transfer');
    fillValid();
    submit();
    expect(await screen.findByText('Це ОСББ ще не підключене — подайте заявку на підключення')).toBeInTheDocument();
    const link = screen.getByRole('link', { name: /заявку на підключення/i });
    expect(link.getAttribute('href')).toContain('kind=connection');
  });

  it('shows the refusal of an ended term (AC-30)', async () => {
    stubApi(() =>
      problem(
        409,
        'identity.term_ended',
        'Строк повноважень голови вже завершився — подайте витяг після переобрання',
      ),
    );
    await renderForm('connection');
    fillValid();
    fireEvent.change(screen.getByLabelText(/Кінець строку/), { target: { value: '2026-11-09' } });
    submit();
    expect(
      await screen.findByText('Строк повноважень голови вже завершився — подайте витяг після переобрання'),
    ).toBeInTheDocument();
  });
});
