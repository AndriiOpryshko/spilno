import type { Meta, StoryObj } from '@storybook/react-vite';

import { LimitNote } from './limit-note';

// Figma: 02 Components / LimitNote.
const meta = {
  title: 'Governance/LimitNote',
  component: LimitNote,
  decorators: [(Story) => <div className="max-w-lg">{Story()}</div>],
} satisfies Meta<typeof LimitNote>;

export default meta;
type Story = StoryObj<typeof meta>;

const ruleRef = { title: 'Статут, п. 9.4', fragment: 'Правління ухвалює рішення щодо договорів до 150 000 грн' };

export const Blocking: Story = {
  args: {
    warning: { id: 'w-1', tier: 3, slotId: 's-amount', ruleRef, title: 'Рішення за загальними зборами', message: 'Ліміт правління — 150 000 грн (Статут, п. 9.4). Винесіть питання на збори або зменште суму.' },
  },
};

export const Advisory: Story = {
  args: {
    warning: { id: 'w-2', tier: 2, slotId: 's-contractor', ruleRef, title: 'Перевірте назву підрядника', message: 'Назва в пропозиції та в договорі різниться.' },
  },
};
