import { cn } from 'cn';

import { StatusGlyph } from './status-glyph';
import type { ProtocolWarning } from './types';

// Tier 3 blocks printing and says what to do; tier 2 is advisory.
export function LimitNote({ warning }: { warning: ProtocolWarning }) {
  const blocking = warning.tier === 3;
  return (
    <div className={cn('flex items-start gap-2 rounded-md px-3 py-2 text-[13px] leading-[18px]', blocking ? 'bg-destructive-subtle' : 'bg-warning-subtle')}>
      <StatusGlyph state={blocking ? 'blocked' : 'attention'} className="mt-0.5" />
      <div>
        <p className={cn('font-semibold', blocking ? 'text-destructive' : 'text-warning')}>{warning.title}</p>
        <p>{warning.message}</p>
      </div>
    </div>
  );
}
