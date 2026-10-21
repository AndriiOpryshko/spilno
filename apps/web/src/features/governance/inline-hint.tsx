import { uk } from '@/i18n/uk';

import { StatusGlyph } from './status-glyph';
import type { SlotHint } from './validation';

// Calm guidance under a field; appears while typing and never blocks saving.
export function InlineHint({ hint, onPick }: { hint: SlotHint; onPick: (value: string) => void }) {
  return (
    <div className="space-y-2" role="status" aria-live="polite">
      <div className="flex items-start gap-2 text-[13px] leading-[18px] text-warning">
        <StatusGlyph state={hint.kind === 'suggestion' ? 'attention' : 'empty'} className="mt-0.5" />
        <span className="flex-1">{hint.message}</span>
        {hint.kind === 'suggestion' && (
          <button
            type="button"
            onClick={() => onPick(hint.fix)}
            className="rounded-sm bg-warning-subtle px-2 py-0.5 text-xs font-semibold text-warning outline-none hover:brightness-95 focus-visible:ring-2 focus-visible:ring-ring"
          >
            {uk.protocol.fix}
          </button>
        )}
      </div>
      {hint.kind === 'examples' && (
        <div className="flex flex-wrap gap-2">
          {hint.examples.map((e) => (
            <button
              key={e}
              type="button"
              onClick={() => onPick(e)}
              className="rounded-full border bg-card px-3 py-1 text-xs font-medium text-muted-foreground outline-none hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
            >
              {e}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
