import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import { InlineHint } from './inline-hint';

// Figma: 02 Components / InlineHint (Kind=Suggestion, Kind=Examples).
const meta = {
  title: 'Governance/InlineHint',
  component: InlineHint,
  args: { onPick: fn() },
  decorators: [(Story) => <div className="max-w-lg">{Story()}</div>],
} satisfies Meta<typeof InlineHint>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Suggestion: Story = {
  args: { hint: { kind: 'suggestion', message: 'У договорі повна назва — ТОВ «Дах-Сервіс». Протокол має з нею збігатися.', fix: 'ТОВ «Дах-Сервіс»' } },
  play: async ({ args, canvasElement }) => {
    await userEvent.click(within(canvasElement).getByRole('button', { name: 'Підставити' }));
    await expect(args.onPick).toHaveBeenCalledWith('ТОВ «Дах-Сервіс»');
  },
};

export const Examples: Story = {
  args: { hint: { kind: 'examples', message: 'Оберіть варіант або впишіть свій', examples: ['Фонд ремонту', 'Внески співвласників', 'Залишок 2025 року'] } },
  play: async ({ args, canvasElement }) => {
    await userEvent.click(within(canvasElement).getByRole('button', { name: 'Внески співвласників' }));
    await expect(args.onPick).toHaveBeenCalledWith('Внески співвласників');
  },
};

export const Required: Story = {
  args: { hint: { kind: 'required', message: 'Поле потрібне для протоколу' } },
};
