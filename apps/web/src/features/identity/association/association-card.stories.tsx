import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';

import type { AssociationCard as Card } from '@/api/gen/model';

import { AssociationCard } from './association-card';

const base: Card = {
  id: '00000000-0000-7000-8000-000000000002',
  name: 'ОСББ «Тестовий Будинок»',
  edrpou: '12345678',
  address: 'м. Київ, вул. Тестова, 1',
  state: 'connected',
  head_full_name: 'Тестовий Голова Правління',
  connected_by_hand: false,
  history: [
    {
      kind: 'connected',
      occurred_at: '2026-11-11T08:05:00Z',
      decision_id: '00000000-0000-7000-8000-00000000f002',
      filed_at: '2026-11-10T09:00:00Z',
      reason: null,
      head_full_name: 'Тестовий Голова Правління',
      signer_full_name: 'Підписант Тестовий',
      extract_id: '00000000-0000-7000-8000-00000000e001',
      extract_dated_on: '2026-11-08',
    },
  ],
};

const meta = { title: 'Identity/AssociationCard', component: AssociationCard } satisfies Meta<typeof AssociationCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const WithExtract: Story = {
  args: { card: base },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('heading', { name: 'ОСББ «Тестовий Будинок»' })).toBeInTheDocument();
  },
};

export const ByHand: Story = {
  args: { card: { ...base, edrpou: null, head_full_name: null, connected_by_hand: true, history: [] } },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('link', { name: 'Подати запит на передачу' })).toBeInTheDocument();
  },
};
