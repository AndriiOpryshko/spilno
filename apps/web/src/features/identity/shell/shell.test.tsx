import { createMemoryHistory, createRootRoute, createRoute, createRouter, RouterProvider } from '@tanstack/react-router';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import type { ReactNode } from 'react';
import { describe, expect, it } from 'vitest';

import { getStartSignInUrl } from '@/api/gen/endpoints/identity/identity';
import { Sidebar } from '@/components/app-shell';

import { isNoAccessProblem, resolveEntry } from './entry';
import { AssociationPicker } from './association-picker';
import { NoAccessPage } from './no-access-page';
import { SignInPage } from './sign-in-page';

const lypky = '00000000-0000-7000-8000-000000000001';
const sonyachnyi = '00000000-0000-7000-8000-000000000002';

const membership = (association_id: string, association_name: string) => ({
  association_id,
  association_name,
  association_state: 'connected' as const,
  board_member_id: '00000000-0000-7000-8000-00000000b001',
  position: 'member' as const,
});
const person = { account_id: '00000000-0000-7000-8000-00000000a001', email: 'user-1@example.test', is_operator: false };
const operator = { ...person, is_operator: true };

// Links need a router; the page under test is the root route's component.
async function renderInRouter(ui: ReactNode) {
  const router = createRouter({
    routeTree: createRootRoute({ component: () => <>{ui}</> }),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  });
  render(<RouterProvider router={router} />);
  await waitFor(() => expect(document.body.textContent).not.toBe(''));
}

// The same, with the оператор's section as a second route, to follow the link.
async function renderWithOperatorSection(ui: ReactNode) {
  const root = createRootRoute();
  const home = createRoute({ getParentRoute: () => root, path: '/', component: () => <>{ui}</> });
  const section = createRoute({ getParentRoute: () => root, path: '/operator', component: () => <p>Розділ оператора</p> });
  const router = createRouter({ routeTree: root.addChildren([home, section]), history: createMemoryHistory({ initialEntries: ['/'] }) });
  render(<RouterProvider router={router} />);
  await screen.findByRole('link', { name: 'Оператор' });
  return router;
}

describe('resolveEntry (AC-04d)', () => {
  it('sends a visitor without a session to sign-in', () => {
    expect(resolveEntry({ session: undefined, memberships: [] })).toEqual({ kind: 'sign-in' });
  });

  it('sends a person on no roster and not the operator to their applications', () => {
    expect(resolveEntry({ session: person, memberships: [] })).toEqual({ kind: 'applications' });
  });

  it('keeps the оператор on no roster at the picker, where the operator entry is', () => {
    expect(resolveEntry({ session: operator, memberships: [] })).toEqual({ kind: 'picker' });
  });

  it('goes straight into the only association, with its id in the address', () => {
    expect(resolveEntry({ session: person, memberships: [membership(lypky, 'ОСББ «Липки»')] })).toEqual({
      kind: 'association',
      associationId: lypky,
    });
  });

  it('asks a person on two rosters to pick', () => {
    const memberships = [membership(lypky, 'ОСББ «Липки»'), membership(sonyachnyi, 'ОСББ «Сонячний»')];
    expect(resolveEntry({ session: person, memberships })).toEqual({ kind: 'picker' });
  });
});

describe('AssociationPicker: filing for another association (AC-04d)', () => {
  it('offers the заявка на підключення link under the associations, to the connection form', async () => {
    await renderInRouter(
      <AssociationPicker memberships={[membership(lypky, 'ОСББ «Липки»'), membership(sonyachnyi, 'ОСББ «Сонячний»')]} isOperator={false} />,
    );
    const link = screen.getByRole('link', { name: 'Подати заявку на підключення' });
    expect(link).toHaveAttribute('href', '/applications/new?kind=connection');
    const last = screen.getAllByRole('link', { name: 'Картка ОСББ' }).at(-1)!;
    expect(last.compareDocumentPosition(link) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });
});

describe('AssociationPicker: the lasting link to the list of заявки (AC-05, AC-14)', () => {
  it.each([
    ['one roster', [membership(lypky, 'ОСББ «Липки»')]],
    ['several rosters', [membership(lypky, 'ОСББ «Липки»'), membership(sonyachnyi, 'ОСББ «Сонячний»')]],
  ])('offers «Мої заявки» to /applications for a person on %s', async (_label, memberships) => {
    await renderInRouter(<AssociationPicker memberships={memberships} isOperator={false} />);
    expect(screen.getByRole('link', { name: 'Мої заявки' })).toHaveAttribute('href', '/applications');
  });
});

describe('SignInPage', () => {
  it('offers the sign-in link that starts the identity flow', () => {
    render(<SignInPage />);
    expect(screen.getByRole('link')).toHaveAttribute('href', getStartSignInUrl());
  });
});

describe('AssociationPicker (AC-04d)', () => {
  const memberships = [membership(lypky, 'ОСББ «Липки»'), membership(sonyachnyi, 'ОСББ «Сонячний»')];

  it('lists both associations and carries each id in its link', async () => {
    await renderInRouter(<AssociationPicker memberships={memberships} isOperator={false} />);
    const lypkyLink = screen.getByRole('link', { name: /Липки/ });
    const sonyachnyiLink = screen.getByRole('link', { name: /Сонячний/ });
    expect(lypkyLink.getAttribute('href')).toContain(`/o/${lypky}/`);
    expect(sonyachnyiLink.getAttribute('href')).toContain(`/o/${sonyachnyi}/`);
  });

  it('shows no operator entry to a person who is not the operator (AC-16)', async () => {
    await renderInRouter(<AssociationPicker memberships={memberships} isOperator={false} />);
    expect(screen.queryByRole('link', { name: 'Оператор' })).toBeNull();
  });

  it('shows the operator entry next to the picker for an operator who is also a member', async () => {
    await renderInRouter(<AssociationPicker memberships={memberships} isOperator />);
    expect(screen.getByRole('link', { name: 'Оператор' }).getAttribute('href')).toBe('/operator');
    expect(screen.getAllByRole('link', { name: /Липки|Сонячний/ })).toHaveLength(2);
  });
});

describe('the operator entry navigates inside the app (AC-16)', () => {
  it('opens the section from the picker without leaving the page', async () => {
    const router = await renderWithOperatorSection(<AssociationPicker memberships={[]} isOperator />);
    fireEvent.click(screen.getByRole('link', { name: 'Оператор' }));
    expect(await screen.findByText('Розділ оператора')).toBeInTheDocument();
    expect(router.state.location.pathname).toBe('/operator');
  });

  it('opens the section from the sidebar without leaving the page', async () => {
    const router = await renderWithOperatorSection(<Sidebar associationId={lypky} associationName="ОСББ «Липки»" isOperator />);
    fireEvent.click(screen.getByRole('link', { name: 'Оператор' }));
    expect(await screen.findByText('Розділ оператора')).toBeInTheDocument();
    expect(router.state.location.pathname).toBe('/operator');
  });
});

describe('Sidebar', () => {
  it('names the association it belongs to and keeps the id in the cases link', async () => {
    await renderInRouter(<Sidebar associationId={lypky} associationName="ОСББ «Липки»" isOperator={false} />);
    const nav = screen.getByRole('complementary');
    expect(within(nav).getByText('ОСББ «Липки»')).toBeInTheDocument();
    expect(within(nav).getByRole('link', { name: /Справи/ }).getAttribute('href')).toContain(`/o/${lypky}/cases`);
    expect(within(nav).queryByRole('link', { name: 'Оператор' })).toBeNull();
  });

  it('shows the operator link only to the operator', async () => {
    await renderInRouter(<Sidebar associationId={lypky} associationName="ОСББ «Липки»" isOperator />);
    expect(screen.getByRole('link', { name: 'Оператор' }).getAttribute('href')).toBe('/operator');
  });
});

describe('no access (AC-07, AC-16, AC-16b)', () => {
  it.each(['identity.not_current_member', 'identity.not_operator'])('treats a 403 %s as no access', (code) => {
    expect(isNoAccessProblem({ status: 403, info: { code } })).toBe(true);
  });

  it('does not treat other failures as no access', () => {
    expect(isNoAccessProblem({ status: 500, info: { code: 'x' } })).toBe(false);
    expect(isNoAccessProblem({ status: 401, info: { code: 'identity.session_required' } })).toBe(false);
  });

  it('renders «Немає доступу» and nothing of the requested record', () => {
    const { container } = render(<NoAccessPage />);
    expect(screen.getByRole('heading', { name: 'Немає доступу' })).toBeInTheDocument();
    expect(container.textContent).not.toMatch(/Липки|протокол|справ/i);
  });
});
