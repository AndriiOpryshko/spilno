import type { Decorator, Meta, StoryObj } from '@storybook/react-vite';
import { createMemoryHistory, createRootRoute, createRouter, RouterProvider } from '@tanstack/react-router';
import { expect, within } from 'storybook/test';

import { AssociationPicker } from './association-picker';

const membership = (id: string, name: string) => ({
  association_id: id,
  association_name: name,
  association_state: 'connected' as const,
  board_member_id: '00000000-0000-7000-8000-00000000b001',
  position: 'member' as const,
});

// Links need a router; the story itself is the root route's component.
const withRouter: Decorator = (Story) => {
  const router = createRouter({
    routeTree: createRootRoute({ component: () => <Story /> }),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  });
  return <RouterProvider router={router} />;
};

const meta = {
  title: 'Identity/AssociationPicker',
  component: AssociationPicker,
  decorators: [withRouter],
  args: {
    memberships: [membership('00000000-0000-7000-8000-000000000001', 'ОСББ «Липки»'), membership('00000000-0000-7000-8000-000000000002', 'ОСББ «Сонячний»')],
    isOperator: false,
  },
} satisfies Meta<typeof AssociationPicker>;

export default meta;
type Story = StoryObj<typeof meta>;

export const TwoAssociations: Story = {
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getAllByRole('link')).toHaveLength(2);
  },
};

export const WithOperatorEntry: Story = {
  args: { isOperator: true },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('link', { name: 'Оператор' })).toBeInTheDocument();
  },
};
