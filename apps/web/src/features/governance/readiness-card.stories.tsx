import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';

import { protocols } from './mock-data';
import { ReadinessCard } from './readiness-card';

// Figma: 02 Components / ReadinessCard (Ready=No, Ready=Yes).
const blocked = structuredClone(protocols['c-101']!);
const ready = { ...structuredClone(blocked), warnings: [], slots: blocked.slots.map((s) => ({ ...s, status: 'verified' as const })) };

const meta = {
  title: 'Governance/ReadinessCard',
  component: ReadinessCard,
  decorators: [(Story) => <div className="w-80">{Story()}</div>],
} satisfies Meta<typeof ReadinessCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Blocked: Story = {
  args: { draft: blocked },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByText('2 / 5')).toBeVisible();
    await expect(canvas.getByRole('button', { name: 'Надрукувати протокол' })).toBeDisabled();
  },
};

export const Ready: Story = {
  args: { draft: ready },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByText('Усі поля перевірено')).toBeVisible();
    await expect(canvas.getByRole('button', { name: 'Надрукувати протокол' })).toBeEnabled();
  },
};
