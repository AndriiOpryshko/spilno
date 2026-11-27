import type { Meta, StoryObj } from '@storybook/react-vite';
import { RouterProvider, createMemoryHistory, createRootRoute, createRouter } from '@tanstack/react-router';
import { expect, within } from 'storybook/test';

import { StorageCallback, StorageLine } from './storage-line';

const id = '00000000-0000-7000-8000-000000000002';

const meta = { title: 'Identity/StorageLine', component: StorageLine, args: { associationId: id } } satisfies Meta<typeof StorageLine>;

export default meta;
type Story = StoryObj<typeof meta>;

// Replaces the global fetch for one story; the returned cleanup restores it.
const stubStorage = (body: { connected: boolean; google_account_email: string | null }) => () => {
  const original = globalThis.fetch;
  globalThis.fetch = () => Promise.resolve(new Response(JSON.stringify(body), { status: 200, headers: { 'Content-Type': 'application/json' } }));
  return () => {
    globalThis.fetch = original;
  };
};

export const Connected: Story = {
  beforeEach: stubStorage({ connected: true, google_account_email: 'osbb@example.test' }),
  play: async ({ canvasElement }) => {
    await expect(await within(canvasElement).findByText(/osbb@example\.test/)).toBeInTheDocument();
  },
};

export const NotConnected: Story = {
  beforeEach: stubStorage({ connected: false, google_account_email: null }),
  play: async ({ canvasElement }) => {
    const link = await within(canvasElement).findByRole('link', { name: 'Підключити сховище' });
    await expect(link.getAttribute('href')).toContain(`/identity/associations/${id}/storage-connection/authorization`);
  },
};

// The callback page after the refusal of another account while the association has files (AC-27b).
export const CallbackRefusedHasFiles: Story = {
  beforeEach: () => {
    const original = globalThis.fetch;
    globalThis.fetch = (_input, init) =>
      Promise.resolve(
        init?.method === 'POST'
          ? new Response(JSON.stringify({ type: 'about:blank', title: 'Conflict', status: 409, code: 'identity.storage_has_files', detail: 'the storage already holds files' }), {
              status: 409,
              headers: { 'Content-Type': 'application/problem+json' },
            })
          : new Response(JSON.stringify({ connected: true, google_account_email: 'osbb@example.test' }), { status: 200, headers: { 'Content-Type': 'application/json' } }),
      );
    return () => {
      globalThis.fetch = original;
    };
  },
  render: (args) => {
    const router = createRouter({
      routeTree: createRootRoute({ component: () => <StorageCallback associationId={args.associationId} code="auth-code" state="auth-state" /> }),
      history: createMemoryHistory({ initialEntries: ['/'] }),
    });
    return <RouterProvider router={router} />;
  },
  play: async ({ canvasElement }) => {
    const alert = await within(canvasElement).findByRole('alert');
    await expect(alert).toHaveTextContent('osbb@example.test');
    await expect(alert).toHaveTextContent('файли залишаються');
    await expect(within(canvasElement).getByRole('link', { name: 'До картки ОСББ' })).toBeInTheDocument();
  },
};
