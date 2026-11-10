import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, waitFor, within } from 'storybook/test';

import { Button } from './button';
import { Dialog, DialogContent, DialogTitle } from './dialog';

const meta = {
  title: 'UI/Dialog',
  component: Dialog,
  render: () => (
    <Dialog defaultOpen>
      <DialogContent aria-describedby={undefined}>
        <DialogTitle>Відхилити заявку</DialogTitle>
        <Button type="button">Перша дія</Button>
        <Button type="button" variant="outline">
          Друга дія
        </Button>
      </DialogContent>
    </Dialog>
  ),
} satisfies Meta<typeof Dialog>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Open: Story = {
  play: async ({ canvasElement }) => {
    // The dialog renders in a portal, outside the canvas element.
    const body = within(canvasElement.ownerDocument.body);
    const dialog = await body.findByRole('dialog', { name: 'Відхилити заявку' });
    await expect(dialog).toBeInTheDocument();

    // Focus trap: tabbing past the last control stays inside the dialog.
    for (let i = 0; i < 6; i++) {
      await userEvent.tab();
      await expect(dialog.contains(canvasElement.ownerDocument.activeElement)).toBe(true);
    }

    // Esc closes it.
    await userEvent.keyboard('{Escape}');
    await waitFor(() => expect(body.queryByRole('dialog')).not.toBeInTheDocument());
  },
};
