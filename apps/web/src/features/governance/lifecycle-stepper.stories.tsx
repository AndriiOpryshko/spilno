import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';

import { LifecycleStepper } from './lifecycle-stepper';

// Figma: 05 Lifecycle / LifecycleStepper (node 31-144): Step 1–4, Withdrawn, Compact 1–3.
const meta = {
  title: 'Governance/LifecycleStepper',
  component: LifecycleStepper,
  args: { state: 'draft' },
} satisfies Meta<typeof LifecycleStepper>;

export default meta;
type Story = StoryObj<typeof meta>;

const currentStep = (canvasElement: HTMLElement) => within(canvasElement).getByRole('list').querySelector('[aria-current="step"]');

export const Step1Draft: Story = {
  play: async ({ canvasElement }) => {
    await expect(currentStep(canvasElement)).toHaveTextContent('чернетка');
  },
};

export const Step2ReadyForVote: Story = {
  args: { state: 'ready_for_vote' },
  play: async ({ canvasElement }) => {
    await expect(currentStep(canvasElement)).toHaveTextContent('готовий до голосування');
  },
};

export const Step3ReadyForSigning: Story = { args: { state: 'ready_for_signing' } };
export const Step4Signed: Story = { args: { state: 'signed' } };

export const Withdrawn: Story = {
  args: { state: 'withdrawn' },
  play: async ({ canvasElement }) => {
    await expect(currentStep(canvasElement)).toHaveTextContent('знятий з розгляду');
  },
};

// Below 768 px: "Крок N з 4 · <state>" over a 4-segment bar; the ol stays for screen readers.
export const Compact: Story = {
  args: { state: 'ready_for_vote', compact: true },
  decorators: [(Story) => <div className="w-[358px]">{Story()}</div>],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByText('Крок 2 з 4 · готовий до голосування')).toBeVisible();
    await expect(currentStep(canvasElement)).toHaveTextContent('готовий до голосування');
  },
};
