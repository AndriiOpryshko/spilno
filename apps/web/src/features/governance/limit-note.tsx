import { StatusGlyph } from './status-glyph';
import type { ProtocolWarning } from './types';

// A charter check on a slot. In Stage 1 every tier is advisory (BR-11): it says what to do, never blocks.
export function LimitNote({ warning }: { warning: ProtocolWarning }) {
  return (
    <div className="flex items-start gap-2 rounded-md bg-warning-subtle px-3 py-2 text-[13px] leading-[18px]">
      <StatusGlyph state="attention" className="mt-0.5" />
      <div>
        <p className="font-semibold text-warning">{warning.title}</p>
        <p>{warning.message}</p>
      </div>
    </div>
  );
}
