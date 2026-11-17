import { createMemoryHistory, createRootRoute, createRouter, RouterProvider } from '@tanstack/react-router';
import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import type { Application } from '@/api/gen/model';

import { uk } from '@/i18n/uk';

import { StatusLine } from './status-line';

const application = (over: Partial<Application>): Application => ({
  id: '00000000-0000-7000-8000-0000000000a1',
  kind: 'connection',
  edrpou: '12345678',
  association_id: null,
  association_name: 'ОСББ «Тест»',
  state: 'pending',
  filed_at: '2026-11-10',
  rejection_reason: null,
  created_at: '2026-11-10T10:00:00Z',
  ...over,
});

afterEach(() => vi.unstubAllEnvs());

describe('StatusLine', () => {
  it('shows a pending заявка with the filing date and the оператор contact address (AC-05)', () => {
    vi.stubEnv('VITE_OPERATOR_CONTACT_EMAIL', 'operator@example.test');
    render(<StatusLine application={application({ state: 'pending' })} />);
    expect(screen.getByText(/очікує розгляду з 10\.11\.2026/)).toBeInTheDocument();
    expect(screen.getByText(/operator@example\.test/)).toBeInTheDocument();
  });

  it.each([['unset', undefined], ['empty', '']])(
    'shows a visible fallback instead of an empty contact when the address is %s (AC-05)',
    (_name, value) => {
      vi.stubEnv('VITE_OPERATOR_CONTACT_EMAIL', value as string);
      render(<StatusLine application={application({ state: 'pending' })} />);
      expect(screen.getByText(uk.applications.status.contactMissing)).toBeInTheDocument();
      expect(screen.queryByText(/Питання щодо заявки/)).not.toBeInTheDocument();
    },
  );

  it('shows nothing of the association for a pending заявка (AC-05)', () => {
    render(<StatusLine application={application({ state: 'pending', association_id: null })} />);
    expect(screen.queryByRole('link', { name: /ОСББ|асоціац/i })).not.toBeInTheDocument();
  });

  it('shows an approved заявка with a link into its association', async () => {
    const approved = application({ state: 'approved', association_id: '00000000-0000-7000-8000-0000000000b1' });
    const router = createRouter({
      routeTree: createRootRoute({ component: () => <StatusLine application={approved} /> }),
      history: createMemoryHistory({ initialEntries: ['/'] }),
    });
    render(<RouterProvider router={router} />);
    expect(await screen.findByText(/схвалена/i)).toBeInTheDocument();
    expect(screen.queryByText(/очікує розгляду/)).not.toBeInTheDocument();
    expect(screen.getByRole('link', { name: /ОСББ/ })).toHaveAttribute(
      'href',
      expect.stringContaining('00000000-0000-7000-8000-0000000000b1'),
    );
  });

  it('shows a rejected заявка with the reason (AC-14)', () => {
    render(
      <StatusLine application={application({ state: 'rejected', rejection_reason: 'Підпис КЕП недійсний' })} />,
    );
    expect(screen.getByText(/відхилена/i)).toBeInTheDocument();
    expect(screen.getByText(/Підпис КЕП недійсний/)).toBeInTheDocument();
  });
});
