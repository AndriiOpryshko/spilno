import type { Decorator, Meta, StoryObj } from '@storybook/react-vite';
import { createMemoryHistory, createRootRoute, createRouter, RouterProvider } from '@tanstack/react-router';
import { expect, within } from 'storybook/test';

import { SuspendedNotice } from '@/features/identity/association/suspended-notice';

import { AssociationShell } from './app-shell';

const associationId = '00000000-0000-7000-8000-000000000001';
const reason = 'спір щодо складу правління';

// Links need a router; the story itself is the root route's component.
const withRouter: Decorator = (Story) => {
  const router = createRouter({
    routeTree: createRootRoute({ component: () => <Story /> }),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  });
  return <RouterProvider router={router} />;
};

// The association layout (T68): the sidebar, the suspended notice above the page of a вимкнене ОСББ, and the page.
const meta = {
  title: 'Shell/AssociationShell',
  component: AssociationShell,
  decorators: [withRouter],
  args: {
    associationId,
    associationName: 'ОСББ «Липки»',
    isOperator: false,
    casesCount: 4,
    children: <h1>Справи</h1>,
  },
} satisfies Meta<typeof AssociationShell>;

export default meta;
type Story = StoryObj<typeof meta>;

export const WithoutNotice: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(await canvas.findByRole('heading', { name: 'Справи' })).toBeInTheDocument();
    await expect(canvas.getByRole('link', { name: 'ОСББ' })).toBeInTheDocument();
    await expect(canvas.queryByRole('alert')).toBeNull();
    await expect(canvas.queryByText(/Доступ призупинено/)).toBeNull();
  },
};

export const WithSuspendedNotice: Story = {
  args: { notice: <SuspendedNotice reason={reason} /> },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    // One notice, above the page, and the navigation stays.
    const notice = await canvas.findByRole('alert');
    await expect(notice).toHaveTextContent(`Доступ призупинено: ${reason}`);
    await expect(canvas.getAllByRole('alert')).toHaveLength(1);
    await expect(canvas.getByRole('heading', { name: 'Справи' })).toBeInTheDocument();
    await expect(notice.compareDocumentPosition(canvas.getByRole('heading', { name: 'Справи' })) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    await expect(canvas.getByRole('link', { name: 'ОСББ' })).toBeInTheDocument();
  },
};
