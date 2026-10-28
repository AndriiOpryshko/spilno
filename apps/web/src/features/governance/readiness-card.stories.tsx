import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';

import { protocols } from './mock-data';
import { ReadinessCard } from './readiness-card';

// Figma: 02 Components / ReadinessCard (Ready=No, Ready=Yes). Only unconfirmed AI values block (BR-11).
const blocked = structuredClone(protocols['c-101']!);
const confirmed = { ...structuredClone(blocked), slots: blocked.slots.map((s) => ({ ...s, status: 'verified' as const })) };
const ready = { ...confirmed, warnings: [] };

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
    await expect(canvas.getByText('3 / 5')).toBeVisible();
    await expect(canvas.getByText('Підрядник: уточніть значення')).toBeVisible();
    await expect(canvas.getByText('Джерело коштів: не вказано')).toBeVisible();
    await expect(canvas.getByText(/Сума договору: потрібне рішення зборів/)).toBeVisible();
  },
};

// A tier-3 charter check alone is advisory: the AI values are confirmed, nothing blocks.
export const AdvisoryOnly: Story = {
  args: { draft: confirmed },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByText('Усі значення помічника підтверджено')).toBeVisible();
    await expect(canvas.getByText(/потрібне рішення зборів \(порада\)/)).toBeVisible();
  },
};

export const Ready: Story = {
  args: { draft: ready },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByText('Усі поля перевірено')).toBeVisible();
    await expect(canvas.queryByRole('button')).not.toBeInTheDocument();
  },
};
