import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';

import { Button } from '@/components/ui/button';

import { InlineAlert } from './inline-alert';

// Figma: 05 Lifecycle / InlineAlert (node 31-161).
const meta = {
  title: 'Governance/InlineAlert',
  component: InlineAlert,
  args: { title: 'Не вдалося змінити стан протоколу', children: 'Стан протоколу не змінено. Оновіть сторінку й спробуйте ще раз — ваші правки не втрачено.' },
  decorators: [(Story) => <div className="max-w-xl">{Story()}</div>],
} satisfies Meta<typeof InlineAlert>;

export default meta;
type Story = StoryObj<typeof meta>;

export const ErrorTone: Story = {
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('alert')).toHaveTextContent('Не вдалося змінити стан протоколу');
  },
};

export const NeutralWithAction: Story = {
  args: {
    tone: 'neutral',
    title: 'Чернетки протоколу ще немає',
    children: 'Документи справи c-102 ще обробляються. Чернетка з’явиться тут автоматично.',
    action: <Button variant="outline">До справ</Button>,
  },
};
