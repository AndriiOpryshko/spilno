import type { Decorator, Meta, StoryObj } from '@storybook/react-vite';
import { RouterProvider, createMemoryHistory, createRootRoute, createRouter } from '@tanstack/react-router';
import { expect, within } from 'storybook/test';
import { expect as vitestExpect } from 'vitest';
import { page } from 'vitest/browser';

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
      ended_head_full_name: null,
      signer_full_name: 'Підписант Тестовий',
      extract_id: '00000000-0000-7000-8000-00000000e001',
      extract_dated_on: '2026-11-08',
    },
  ],
};

const withRouter: Decorator = (Story) => {
  const router = createRouter({
    routeTree: createRootRoute({ component: () => <Story /> }),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  });
  return <RouterProvider router={router} />;
};

const meta = { title: 'Identity/AssociationCard', component: AssociationCard, decorators: [withRouter] } satisfies Meta<typeof AssociationCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const WithExtract: Story = {
  args: { card: base },
  play: async ({ canvasElement }) => {
    await expect(await within(canvasElement).findByRole('heading', { name: 'ОСББ «Тестовий Будинок»' })).toBeInTheDocument();
    await document.fonts.ready;
    await vitestExpect(page.elementLocator(canvasElement)).toMatchScreenshot('association-card');
  },
};

export const ByHand: Story = {
  args: { card: { ...base, edrpou: null, head_full_name: null, connected_by_hand: true, history: [] } },
  play: async ({ canvasElement }) => {
    await expect(await within(canvasElement).findByRole('link', { name: 'Подати запит на передачу' })).toBeInTheDocument();
  },
};

const transfer = {
  kind: 'transfer_approved' as const,
  occurred_at: '2026-12-05T08:05:00Z',
  decision_id: '00000000-0000-7000-8000-00000000f003',
  filed_at: '2026-12-03T09:00:00Z',
  reason: null,
  head_full_name: 'Новий Голова Правління',
  signer_full_name: 'Підписант Тестовий',
  extract_id: '00000000-0000-7000-8000-00000000e002',
  extract_dated_on: '2026-12-01',
};

export const TransferEndedAHead: Story = {
  args: { card: { ...base, head_full_name: 'Новий Голова Правління', history: [{ ...transfer, ended_head_full_name: 'Колишній Голова Правління' }] } },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(await canvas.findByText(/повноваження завершено/)).toHaveTextContent('Колишній Голова Правління');
    await expect(canvas.getByText(/призначено/)).toHaveTextContent('Новий Голова Правління');
  },
};

// AC-19c: the association set up by hand had no голова entry, so the transfer ends no term.
export const TransferOnHandMadeAssociation: Story = {
  args: { card: { ...base, edrpou: null, head_full_name: 'Новий Голова Правління', history: [{ ...transfer, ended_head_full_name: null }] } },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(await canvas.findByText(/призначено/)).toHaveTextContent('Новий Голова Правління');
    await expect(canvas.queryByText(/повноваження завершено/)).toBeNull();
  },
};
