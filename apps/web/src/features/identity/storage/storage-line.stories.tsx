import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';

import { StorageLine } from './storage-line';

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
