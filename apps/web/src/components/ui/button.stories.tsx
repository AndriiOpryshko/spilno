import type { Meta, StoryObj } from '@storybook/react-vite';

import { Button } from './button';

// Figma: 02 Components / Button (Primary, Tonal, Outline × Default, Disabled).
const meta = {
  title: 'UI/Button',
  component: Button,
  args: { children: 'Надрукувати протокол' },
} satisfies Meta<typeof Button>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Primary: Story = {};
export const Tonal: Story = { args: { variant: 'secondary', children: 'Підставити' } };
export const Outline: Story = { args: { variant: 'outline', children: 'Відкрити' } };

export const Disabled: Story = {
  render: () => (
    <div className="flex gap-2">
      <Button disabled>Надрукувати протокол</Button>
      <Button variant="secondary" disabled>
        Підставити
      </Button>
      <Button variant="outline" disabled>
        Відкрити
      </Button>
    </div>
  ),
};
