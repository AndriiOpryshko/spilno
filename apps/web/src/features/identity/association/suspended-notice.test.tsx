import { createMemoryHistory, createRootRoute, createRouter, RouterProvider } from '@tanstack/react-router';
import { render, screen, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { describe, expect, it } from 'vitest';

import { AssociationProblemNotice, SuspendedNotice } from './suspended-notice';

const associationId = '00000000-0000-7000-8000-000000000002';

async function renderInRouter(ui: ReactNode) {
  const router = createRouter({
    routeTree: createRootRoute({ component: () => <>{ui}</> }),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  });
  render(<RouterProvider router={router} />);
  await waitFor(() => expect(document.body.textContent).not.toBe(''));
}

describe('SuspendedNotice (AC-21, AC-25)', () => {
  it('shows «Доступ призупинено» with the reason', async () => {
    await renderInRouter(<SuspendedNotice reason="спір щодо складу правління" />);
    expect(screen.getByText('Доступ призупинено: спір щодо складу правління')).toBeInTheDocument();
  });
});

describe('AssociationProblemNotice', () => {
  it('shows the suspended notice with the reason from the detail of a 403 identity.association_disabled (AC-25)', async () => {
    const error = { status: 403, info: { code: 'identity.association_disabled', detail: 'спір щодо складу правління' } };
    await renderInRouter(<AssociationProblemNotice error={error} associationId={associationId} />);
    expect(screen.getByText('Доступ призупинено: спір щодо складу правління')).toBeInTheDocument();
  });

  it('asks to connect the storage first, with a link to the card and the note that any член правління can (AC-28)', async () => {
    const error = { status: 409, info: { code: 'governance.drive_not_connected', detail: 'x' } };
    await renderInRouter(<AssociationProblemNotice error={error} associationId={associationId} />);
    expect(screen.getByText(/Спершу підключіть сховище документів ОСББ/)).toBeInTheDocument();
    expect(document.body.textContent).toMatch(/будь-який член правління/);
    const link = screen.getByRole('link');
    expect(link.getAttribute('href')).toContain(`/o/${associationId}/association`);
  });

  it('renders nothing for another problem', async () => {
    const error = { status: 500, info: { code: 'x.y', detail: 'boom' } };
    await renderInRouter(
      <>
        <span>ok</span>
        <AssociationProblemNotice error={error} associationId={associationId} />
      </>,
    );
    expect(screen.queryByRole('link')).not.toBeInTheDocument();
    expect(document.body.textContent).toBe('ok');
  });
});
