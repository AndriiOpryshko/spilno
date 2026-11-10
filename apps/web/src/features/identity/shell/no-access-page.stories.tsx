import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';

import { NoAccessPage } from './no-access-page';

const meta = { title: 'Identity/NoAccessPage', component: NoAccessPage } satisfies Meta<typeof NoAccessPage>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('heading', { name: 'Немає доступу' })).toBeInTheDocument();
  },
};
