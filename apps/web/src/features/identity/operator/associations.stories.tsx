import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';

import { OperatorAssociations } from './associations';

const items = [
  { id: '00000000-0000-7000-8000-000000000001', name: 'ОСББ «Липки»', edrpou: '12345678', state: 'connected', connected_by_hand: false, created_at: '2026-01-01T00:00:00Z' },
  { id: '00000000-0000-7000-8000-000000000002', name: 'ОСББ «Сонячний»', edrpou: null, state: 'disabled', connected_by_hand: true, created_at: '2026-01-02T00:00:00Z' },
];

const meta = {
  title: 'Identity/OperatorAssociations',
  component: OperatorAssociations,
  decorators: [
    (Story) => {
      globalThis.fetch = () =>
        Promise.resolve(new Response(JSON.stringify({ items, next_cursor: null }), { status: 200, headers: { 'Content-Type': 'application/json' } }));
      return (
        <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
          <Story />
        </QueryClientProvider>
      );
    },
  ],
} satisfies Meta<typeof OperatorAssociations>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  play: async ({ canvasElement }) => {
    await expect(await within(canvasElement).findByText(/Липки/)).toBeInTheDocument();
  },
};
