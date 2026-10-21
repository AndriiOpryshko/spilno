import type { Meta, StoryObj } from '@storybook/react-vite';

import { OriginTag } from './origin-tag';
import { StatusGlyph } from './status-glyph';

// Figma: 02 Components / StatusGlyph and OriginTag.
const meta = { title: 'Governance/Badges' } satisfies Meta;

export default meta;
type Story = StoryObj<typeof meta>;

export const StatusGlyphs: Story = {
  render: () => (
    <ul className="space-y-2 text-sm">
      {(['done', 'attention', 'empty', 'blocked', 'neutral'] as const).map((s) => (
        <li key={s} className="flex items-center gap-2">
          <StatusGlyph state={s} label={s} /> {s}
        </li>
      ))}
    </ul>
  ),
};

export const OriginTags: Story = {
  render: () => (
    <div className="flex gap-2">
      <OriginTag origin="system" />
      <OriginTag origin="ai" />
      <OriginTag origin="human" />
      <OriginTag origin="editing" />
    </div>
  ),
};
