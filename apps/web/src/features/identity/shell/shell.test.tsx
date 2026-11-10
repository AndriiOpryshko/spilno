import { createMemoryHistory, createRootRoute, createRouter, RouterProvider } from '@tanstack/react-router';
import { render, screen, waitFor, within } from '@testing-library/react';
import type { ReactNode } from 'react';
import { describe, expect, it } from 'vitest';

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

describe('SignInPage', () => {
  it('offers the sign-in link that starts the identity flow', () => {
    render(<SignInPage />);
    expect(screen.getByRole('link')).toHaveAttribute('href', '/identity/sign-in');
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
    expect(screen.getByRole('link', { name: 'Оператор' }).getAttribute('href')).toContain('/operator');
    expect(screen.getAllByRole('link', { name: /Липки|Сонячний/ })).toHaveLength(2);
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
    expect(screen.getByRole('link', { name: 'Оператор' })).toBeInTheDocument();
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
