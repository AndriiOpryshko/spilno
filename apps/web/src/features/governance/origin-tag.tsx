import { cn } from 'cn';

import { uk } from '@/i18n/uk';

import type { SlotOrigin } from './types';

const look: Record<SlotOrigin | 'editing', string> = {
  system: 'bg-origin-system-subtle text-origin-system',
  ai: 'bg-origin-ai-subtle text-origin-ai',
  human: 'bg-origin-human-subtle text-origin-human',
  editing: 'bg-origin-human-subtle text-origin-human',
};

export function OriginTag({ origin, className }: { origin: SlotOrigin | 'editing'; className?: string }) {
  return <span className={cn('inline-flex h-5 items-center rounded-sm px-2 text-xs font-semibold', look[origin], className)}>{uk.origin[origin]}</span>;
}
