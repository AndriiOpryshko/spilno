import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';

import { ProtocolStateBadge } from './protocol-state-badge';

// Figma: 05 Lifecycle / ProtocolStateBadge (node 31-19), one variant per CONTEXT.md state.
const meta = {
  title: 'Governance/ProtocolStateBadge',
  component: ProtocolStateBadge,
  args: { state: 'draft' },
} satisfies Meta<typeof ProtocolStateBadge>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Draft: Story = {
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText('чернетка')).toBeVisible();
  },
};
export const ReadyForVote: Story = { args: { state: 'ready_for_vote' } };
export const ReadyForSigning: Story = { args: { state: 'ready_for_signing' } };
export const Signed: Story = { args: { state: 'signed' } };
export const Withdrawn: Story = { args: { state: 'withdrawn' } };

export const All: Story = {
  render: () => (
    <div className="flex flex-wrap gap-3">
      {(['draft', 'ready_for_vote', 'ready_for_signing', 'signed', 'withdrawn'] as const).map((s) => (
        <ProtocolStateBadge key={s} state={s} />
      ))}
    </div>
  ),
};
