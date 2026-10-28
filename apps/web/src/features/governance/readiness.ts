import type { ProtocolDraft, ProtocolSlot, ProtocolWarning } from './types';

export interface Readiness {
  ready: boolean;
  total: number;
  verified: number;
  // Values the помічник proposed that nobody confirmed yet: the only Stage 1 gate (AC-06).
  blockingSlots: ProtocolSlot[];
  // Tier-3 checks are Stage 2 blocking checks; in Stage 1 they stay visible and never block.
  advisoryWarnings: ProtocolWarning[];
}

export const isUnconfirmedAi = (s: ProtocolSlot) => s.origin === 'ai' && s.status !== 'verified';

// Marking готовий до підпису and printing wait only for unconfirmed AI-proposed values (BR-11).
export function getReadiness(draft: ProtocolDraft): Readiness {
  const blockingSlots = draft.slots.filter(isUnconfirmedAi);
  return {
    ready: blockingSlots.length === 0,
    total: draft.slots.length,
    verified: draft.slots.filter((s) => s.status === 'verified').length,
    blockingSlots,
    advisoryWarnings: draft.warnings.filter((w) => w.tier === 3),
  };
}
