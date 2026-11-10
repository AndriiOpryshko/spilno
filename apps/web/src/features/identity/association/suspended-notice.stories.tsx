import type { Decorator, Meta, StoryObj } from '@storybook/react-vite';
import { RouterProvider, createMemoryHistory, createRootRoute, createRouter } from '@tanstack/react-router';
import { expect, within } from 'storybook/test';

import { AssociationProblemNotice } from './suspended-notice';

const id = '00000000-0000-7000-8000-000000000002';

const withRouter: Decorator = (Story) => {
  const router = createRouter({
    routeTree: createRootRoute({ component: () => <Story /> }),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  });
  return <RouterProvider router={router} />;
};

const meta = {
  title: 'Identity/AssociationProblemNotice',
  component: AssociationProblemNotice,
  decorators: [withRouter],
  args: { associationId: id },
} satisfies Meta<typeof AssociationProblemNotice>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Suspended: Story = {
  args: { error: { status: 403, info: { code: 'identity.association_disabled', detail: 'Договір не оплачено' } } },
  play: async ({ canvasElement }) => {
    await expect(await within(canvasElement).findByText(/Договір не оплачено/)).toBeInTheDocument();
  },
};

export const ConnectStorageFirst: Story = {
  args: { error: { status: 409, info: { code: 'governance.drive_not_connected', detail: '' } } },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(await canvas.findByText(/будь-який член правління/)).toBeInTheDocument();
    await expect((await canvas.findByRole('link')).getAttribute('href')).toContain(`/o/${id}/association`);
  },
};
