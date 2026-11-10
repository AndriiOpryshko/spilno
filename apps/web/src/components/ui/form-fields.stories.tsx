import type { Meta, StoryObj } from '@storybook/react-vite';

import { DateField } from './date-field';
import { FileField } from './file-field';
import { Label } from './label';

function Fields() {
  return (
    <div className="grid max-w-sm gap-4">
      <div className="grid gap-1.5">
        <Label htmlFor="d">Дата</Label>
        <DateField id="d" />
      </div>
      <div className="grid gap-1.5">
        <Label htmlFor="d-bad">Дата з помилкою</Label>
        <DateField id="d-bad" aria-invalid />
      </div>
      <div className="grid gap-1.5">
        <Label htmlFor="f">Файл</Label>
        <FileField id="f" />
      </div>
    </div>
  );
}

const meta = { title: 'UI/FormFields', component: Fields } satisfies Meta<typeof Fields>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
