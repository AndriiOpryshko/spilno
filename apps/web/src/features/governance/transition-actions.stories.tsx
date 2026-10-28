import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import { TransitionActions } from './transition-actions';

// Figma: 05 Lifecycle / TransitionActions (node 31-207): Draft, Ready for vote, Ready for signing,
// Pending, Disabled reason; plus the transition error (node 32-2543).
const meta = {
  title: 'Governance/TransitionActions',
  component: TransitionActions,
  args: { state: 'draft', number: '17/2026', blocking: [], pending: null, error: false, onTransition: fn(), onPrint: fn() },
  decorators: [(Story) => <div className="max-w-3xl">{Story()}</div>],
} satisfies Meta<typeof TransitionActions>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Draft: Story = {
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByRole('button', { name: 'Подати на попереднє голосування' }));
    await expect(args.onTransition).toHaveBeenCalledWith('ready_for_vote');
    await userEvent.click(canvas.getByRole('button', { name: 'Позначити готовим до підпису' }));
    await expect(args.onTransition).toHaveBeenCalledWith('ready_for_signing');
  },
};

export const DisabledReason: Story = {
  args: { blocking: ['Сума за договором'] },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const mark = canvas.getByRole('button', { name: 'Позначити готовим до підпису' });
    await expect(mark).toBeDisabled();
    await expect(mark).toHaveAccessibleDescription('Підтвердіть: Сума за договором');
    await expect(canvas.getByText('Підтвердіть: Сума за договором')).toBeVisible();
  },
};

export const ReadyForVote: Story = {
  args: { state: 'ready_for_vote' },
  play: async ({ args, canvasElement }) => {
    await userEvent.click(within(canvasElement).getByRole('button', { name: 'Повернути в чернетку' }));
    await expect(args.onTransition).toHaveBeenCalledWith('draft');
  },
};

export const ReadyForSigning: Story = {
  args: { state: 'ready_for_signing' },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByRole('button', { name: 'Надрукувати протокол' }));
    await expect(args.onPrint).toHaveBeenCalled();
    await expect(canvas.getByRole('button', { name: 'Повернути в чернетку' })).toHaveAccessibleDescription(
      'Надрукована копія стане недійсною. Номер № 17/2026 залишиться за протоколом.',
    );
  },
};

export const Pending: Story = {
  args: { pending: 'ready_for_signing' },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByRole('button', { name: 'Позначення…' })).toBeDisabled();
    await expect(canvas.getByRole('button', { name: 'Подати на попереднє голосування' })).toBeDisabled();
  },
};

export const Error: Story = {
  args: { state: 'ready_for_vote', error: true },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('alert')).toHaveTextContent('Не вдалося змінити стан протоколу');
  },
};

export const SignedHasNoActions: Story = {
  args: { state: 'signed' },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).queryByRole('button')).not.toBeInTheDocument();
  },
};
