import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import { protocols } from './mock-data';
import { SlotRow } from './slot-row';

// Figma: 02 Components / SlotRow (Verified, Attention, Editing, Empty, Locked).
const draft = protocols['c-101']!;
const slot = (id: string) => structuredClone(draft.slots.find((s) => s.id === id)!);

const meta = {
  title: 'Governance/SlotRow',
  component: SlotRow,
  args: { caseId: 'c-101', warnings: [], onSaved: fn(), slot: slot('s-agenda') },
  decorators: [(Story) => <div className="max-w-4xl overflow-hidden rounded-lg border bg-card">{Story()}</div>],
} satisfies Meta<typeof SlotRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Verified: Story = {};

export const Attention: Story = {
  args: { slot: slot('s-contractor') },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText(/У договорі повна назва/)).toBeVisible();
  },
};

export const Editing: Story = {
  args: { slot: slot('s-contractor') },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const field = canvas.getByLabelText('Підрядник');
    await userEvent.type(field, ' ТОВ');
    await expect(canvas.getAllByText('Редагується')[0]).toBeInTheDocument();
    await expect(canvas.getByRole('button', { name: 'Зберегти' })).toBeEnabled();
    await userEvent.click(canvas.getByRole('button', { name: 'Підставити' }));
    await expect(field).toHaveValue('ТОВ «Дах-Сервіс»');
    await expect(canvas.queryByText(/У договорі повна назва/)).not.toBeInTheDocument();
  },
};

export const Empty: Story = {
  args: { slot: slot('s-funding') },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByRole('button', { name: 'Фонд ремонту' }));
    await expect(canvas.getByLabelText('Джерело коштів')).toHaveValue('Фонд ремонту');
  },
};

export const LockedWithLimit: Story = {
  args: { slot: slot('s-amount'), warnings: draft.warnings.filter((w) => w.slotId === 's-amount') },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.queryByRole('textbox')).not.toBeInTheDocument();
    await expect(canvas.getByText('Рішення за загальними зборами')).toBeVisible();
  },
};
