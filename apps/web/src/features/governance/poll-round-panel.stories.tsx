import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';

import { protocols } from './mock-data';
import { PollRoundPanel } from './poll-round-panel';

// Figma: 05 Lifecycle / PollRoundPanel (node 31-145). State and result only, no vote controls.
const meta = {
  title: 'Governance/PollRoundPanel',
  component: PollRoundPanel,
  args: { poll: structuredClone(protocols['c-103']!.poll!) },
  decorators: [(Story) => <div className="w-80">{Story()}</div>],
} satisfies Meta<typeof PollRoundPanel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Round2: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByRole('heading', { name: 'Попереднє голосування · раунд 2' })).toBeVisible();
    await expect(canvas.getByText('Неформальне опитування — не голосування засідання')).toBeVisible();
    await expect(canvas.getByText('3 з 4 членів правління проголосували')).toBeVisible();
    await expect(canvas.getByText('за 2 · проти 1 · утрим. 0')).toBeVisible();
    await expect(canvas.queryByRole('button')).not.toBeInTheDocument();
  },
};

export const NewRound: Story = {
  args: { poll: { round: 3, voted: 0, boardSize: 4, items: [{ id: 'p-1', title: 'Гідроізоляція покрівлі, під’їзд 2', for: 0, against: 0, abstain: 0 }] } },
};
