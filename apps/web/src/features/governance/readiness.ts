import type { ProtocolDraft } from './types';

export interface Readiness {
  ready: boolean;
  total: number;
  verified: number;
  blockingWarnings: number;
  unresolvedSlots: number;
}

// Printing is blocked by tier-3 checks and by any slot a human has not confirmed yet.
// Tier-2 warnings are advisory and never block.
export function getReadiness(draft: ProtocolDraft): Readiness {
  const blockingWarnings = draft.warnings.filter((w) => w.tier === 3).length;
  const unresolvedSlots = draft.slots.filter((s) => s.status !== 'verified').length;
  const blockedSlots = new Set(draft.warnings.filter((w) => w.tier === 3).map((w) => w.slotId));
  const verified = draft.slots.filter((s) => s.status === 'verified' && !blockedSlots.has(s.id)).length;
  return {
    ready: blockingWarnings === 0 && unresolvedSlots === 0,
    total: draft.slots.length,
    verified,
    blockingWarnings,
    unresolvedSlots,
  };
}
