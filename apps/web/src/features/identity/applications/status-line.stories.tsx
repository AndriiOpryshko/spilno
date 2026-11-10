import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';

import type { Application } from '@/api/gen/model';

import { StatusLine } from './status-line';

const base: Application = {
  id: '00000000-0000-7000-8000-0000000000a1',
  kind: 'connection',
  edrpou: '12345678',
  association_id: null,
  association_name: 'ОСББ «Тест»',
  state: 'pending',
  filed_at: '2026-11-10',
  rejection_reason: null,
  created_at: '2026-11-10T10:00:00Z',
};

const meta = { title: 'Identity/StatusLine', component: StatusLine } satisfies Meta<typeof StatusLine>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Pending: Story = {
  args: { application: base },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText(/очікує розгляду з 10\.11\.2026/)).toBeInTheDocument();
  },
};

export const Approved: Story = {
  args: { application: { ...base, state: 'approved', association_id: '00000000-0000-7000-8000-0000000000b1' } },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('link', { name: /ОСББ/ })).toBeInTheDocument();
  },
};

export const Rejected: Story = {
  args: { application: { ...base, state: 'rejected', rejection_reason: 'Підпис КЕП недійсний' } },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText(/Підпис КЕП недійсний/)).toBeInTheDocument();
  },
};
