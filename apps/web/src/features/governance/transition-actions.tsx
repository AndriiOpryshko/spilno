import { cn } from 'cn';
import { useId } from 'react';

import { Button } from '@/components/ui/button';
import { uk } from '@/i18n/uk';

import { InlineAlert } from './inline-alert';
import { StatusGlyph } from './status-glyph';
import type { ProtocolState } from './types';

interface TransitionActionsProps {
  state: ProtocolState;
  number: string;
  // Labels of AI-proposed values nobody confirmed yet; non-empty disables marking готовий до підпису.
  blocking: string[];
  // Target of the transition in flight: its button shows a pending label, all buttons are disabled.
  pending: ProtocolState | null;
  error: boolean;
  onTransition: (to: ProtocolState) => void;
  onPrint: () => void;
  ref?: React.Ref<HTMLDivElement>;
  className?: string;
}

const t = uk.protocol.actions;

// Lifecycle actions per state (TransitionActions in Figma). Signed and withdrawn protocols have none.
export function TransitionActions({ state, number, blocking, pending, error, onTransition, onPrint, ref, className }: TransitionActionsProps) {
  const reasonId = useId();
  const busy = pending !== null;
  if (state === 'signed' || state === 'withdrawn') return null;

  const move = (to: ProtocolState, label: string, pendingLabel: string, variant: 'default' | 'outline', extra?: { disabled?: boolean; describedBy?: string }) => (
    <Button
      key={to}
      variant={variant}
      className="max-md:w-full"
      disabled={busy || extra?.disabled}
      aria-describedby={extra?.describedBy}
      onClick={() => onTransition(to)}
    >
      {pending === to ? pendingLabel : label}
    </Button>
  );
  const blocked = blocking.length > 0;
  const toSigning = move('ready_for_signing', t.toSigning, t.toSigningPending, 'default', { disabled: blocked, describedBy: blocked ? reasonId : undefined });

  const buttons =
    state === 'draft'
      ? [move('ready_for_vote', t.toVote, t.toVotePending, 'outline'), toSigning]
      : state === 'ready_for_vote'
        ? [move('draft', t.toDraft, t.toDraftPending, 'outline'), toSigning]
        : [
            <Button key="print" className="max-md:w-full" disabled={busy} onClick={onPrint}>
              {t.print}
            </Button>,
            move('draft', t.toDraft, t.toDraftPending, 'outline', { describedBy: reasonId }),
          ];

  return (
    <div ref={ref} role="group" aria-label={t.label} className={cn('space-y-2', className)}>
      <div className="flex flex-col gap-2 md:flex-row md:flex-wrap md:items-center">{buttons}</div>
      {state !== 'ready_for_signing' && blocked && (
        <p id={reasonId} className="flex items-center gap-2 text-[13px]">
          <StatusGlyph state="attention" /> {t.confirmFirst(blocking)}
        </p>
      )}
      {state === 'ready_for_signing' && (
        <p id={reasonId} className="flex items-center gap-2 text-xs text-muted-foreground">
          <StatusGlyph state="neutral" /> {t.voidCopy(number)}
        </p>
      )}
      {error && <InlineAlert title={t.failed}>{t.failedHint}</InlineAlert>}
    </div>
  );
}
