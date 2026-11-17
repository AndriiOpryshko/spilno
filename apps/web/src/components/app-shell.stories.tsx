import type { Decorator, Meta, StoryObj } from '@storybook/react-vite';
import { createMemoryHistory, createRootRoute, createRouter, RouterProvider } from '@tanstack/react-router';
import { expect, within } from 'storybook/test';

import { Sidebar } from './app-shell';

const associationId = '00000000-0000-7000-8000-000000000001';

// Links need a router; the story itself is the root route's component.
const withRouter: Decorator = (Story) => {
  const router = createRouter({
    routeTree: createRootRoute({ component: () => <Story /> }),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  });
  return <RouterProvider router={router} />;
};

const meta = {
  title: 'Shell/Sidebar',
  component: Sidebar,
  decorators: [withRouter],
  args: { associationId, associationName: 'ОСББ «Липки»', isOperator: false, casesCount: 4 },
} satisfies Meta<typeof Sidebar>;

export default meta;
type Story = StoryObj<typeof meta>;

// The card of the association is one click from every page (AC-27, AC-29).
export const WithAssociationCard: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(await canvas.findByRole('link', { name: 'ОСББ' })).toHaveAttribute('href', `/o/${associationId}/association`);
    await expect(canvas.getByRole('link', { name: /Справи/ })).toBeInTheDocument();
  },
};

export const WithOperatorEntry: Story = {
  args: { isOperator: true },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(await canvas.findByRole('link', { name: 'Оператор' })).toHaveAttribute('href', '/operator');
    await expect(canvas.getByRole('link', { name: 'ОСББ' })).toBeInTheDocument();
  },
};
