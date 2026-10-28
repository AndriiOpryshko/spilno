import { cn } from 'cn';

import { Badge } from '@/components/ui/badge';
import { uk } from '@/i18n/uk';

import type { ProtocolState } from './types';

// Pill + dot per lifecycle state (ProtocolStateBadge in Figma, 05 Lifecycle). The text is the CONTEXT.md term.
const look: Record<ProtocolState, { pill: string; dot: string }> = {
  draft: { pill: 'bg-muted text-foreground', dot: 'bg-muted-foreground' },
  ready_for_vote: { pill: 'bg-warning-subtle text-warning', dot: 'bg-warning' },
  ready_for_signing: { pill: 'bg-secondary text-secondary-foreground', dot: 'bg-secondary-foreground' },
  signed: { pill: 'bg-success-subtle text-success', dot: 'bg-success' },
  withdrawn: { pill: 'border-border bg-card text-muted-foreground line-through', dot: 'bg-muted-foreground' },
};

export function ProtocolStateBadge({ state, className }: { state: ProtocolState; className?: string }) {
  return (
    <Badge variant="ghost" data-state={state} className={cn('gap-1.5 px-2.5', look[state].pill, className)}>
      <span aria-hidden className={cn('size-1.5 rounded-full', look[state].dot)} />
      {uk.protocol.state[state]}
    </Badge>
  );
}
