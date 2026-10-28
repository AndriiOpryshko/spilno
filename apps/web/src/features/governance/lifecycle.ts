import { getReadiness } from './readiness';
import type { PollItem, ProtocolDraft, ProtocolState } from './types';

// The four steps of the main path; знятий з розгляду is an exit, not a step.
export const lifecycleSteps = ['draft', 'ready_for_vote', 'ready_for_signing', 'signed'] as const satisfies readonly ProtocolState[];

// Stage 1 transitions a member can make on this screen (PRD.ao-prd.md §10). Signing and withdrawing
// need the scan upload and the withdraw dialog, which come in the next iteration.
const allowed: Record<ProtocolState, ProtocolState[]> = {
  draft: ['ready_for_vote', 'ready_for_signing'],
  ready_for_vote: ['draft', 'ready_for_signing'],
  ready_for_signing: ['draft'],
  signed: [],
  withdrawn: [],
};

export type TransitionErrorCode = 'governance.transition_not_allowed' | 'governance.unconfirmed_values';

export class TransitionError extends Error {
  constructor(
    readonly code: TransitionErrorCode,
    message: string,
  ) {
    super(message);
  }
}

export const canTransition = (from: ProtocolState, to: ProtocolState) => allowed[from].includes(to);

// Returns why the transition is refused, or null when it is allowed.
export function transitionError(draft: ProtocolDraft, to: ProtocolState): TransitionError | null {
  if (!canTransition(draft.state, to)) {
    return new TransitionError('governance.transition_not_allowed', `Transition ${draft.state} → ${to} is not allowed`);
  }
  if (to === 'ready_for_signing' && !getReadiness(draft).ready) {
    return new TransitionError('governance.unconfirmed_values', 'Unconfirmed values proposed by the assistant block printing');
  }
  return null;
}

// Poll items of a new round: the previous round's items with empty tallies, or one item per decision.
function pollItems(draft: ProtocolDraft): PollItem[] {
  const zero = { for: 0, against: 0, abstain: 0 };
  if (draft.poll) return draft.poll.items.map((i) => ({ ...i, ...zero }));
  const funding = draft.slots.find((s) => s.id === 's-funding')?.value.trim();
  return [
    { id: 'p-1', title: draft.caseTitle, ...zero },
    ...(funding ? [{ id: 'p-2', title: `Джерело коштів: ${funding}`, ...zero }] : []),
  ];
}

// Pure transition: throws TransitionError when refused. Entering готовий до голосування opens the next
// round; leaving it closes the round (kept for history, AC-17). The number always stays (AC-26).
export function applyTransition(draft: ProtocolDraft, to: ProtocolState): ProtocolDraft {
  const error = transitionError(draft, to);
  if (error) throw error;
  if (to !== 'ready_for_vote') return { ...draft, state: to };
  return {
    ...draft,
    state: to,
    poll: { round: (draft.poll?.round ?? 0) + 1, voted: 0, boardSize: draft.board.length, items: pollItems(draft) },
  };
}
