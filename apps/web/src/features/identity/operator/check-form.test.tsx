import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { uk } from '@/i18n/uk';

import { CheckForm } from './check-form';

const appId = '00000000-0000-7000-8000-0000000000a1';
const filingId = '00000000-0000-7000-8000-0000000000f1';

const problem = (status: number, body: Record<string, unknown>) =>
  new Response(JSON.stringify({ type: 'about:blank', title: 'x', status, ...body }), {
    status,
    headers: { 'Content-Type': 'application/problem+json' },
  });

function renderForm() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  render(
    <QueryClientProvider client={client}>
      <CheckForm applicationId={appId} filingId={filingId} />
    </QueryClientProvider>,
  );
}

beforeEach(() => {
  vi.stubGlobal(
    'fetch',
    vi.fn(() =>
      Promise.resolve(
        problem(422, {
          errors: [
            { pointer: '/extract_edrpou', detail: 'must be 8 digits' },
            { pointer: '/extract_dated_on', detail: 'must be a date' },
          ],
        }),
      ),
    ),
  );
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe('CheckForm refusals (AC-13)', () => {
  it('names every broken field at once when the ЄДРПОУ and the date are both refused', async () => {
    renderForm();
    fireEvent.change(screen.getByLabelText(uk.operator.edrpou), { target: { value: '123' } });
    fireEvent.change(screen.getByLabelText(uk.operator.extractDatedOn), { target: { value: '2026-13-45' } });
    fireEvent.click(screen.getByRole('button', { name: uk.operator.recordCheck }));

    const alert = await screen.findByRole('alert');
    await waitFor(() => {
      expect(alert).toHaveTextContent(uk.operator.edrpou);
      expect(alert).toHaveTextContent(uk.operator.extractDatedOn);
    });
    expect(alert).not.toHaveTextContent(uk.operator.signer);
    expect(screen.getByLabelText(uk.operator.edrpou)).toHaveAttribute('aria-invalid', 'true');
    expect(screen.getByLabelText(uk.operator.extractDatedOn)).toHaveAttribute('aria-invalid', 'true');
    expect(screen.getByLabelText(uk.operator.signer)).not.toHaveAttribute('aria-invalid', 'true');
  });
});
