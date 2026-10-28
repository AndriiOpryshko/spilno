// Governance read models for the cases list and the board protocol draft.
// Hand-written until the governance context lands in openapi.yaml; then replace with
// orval models from @/api/gen/model.

export type CaseStatus = 'draft' | 'review' | 'ready' | 'printed';

export type DocumentKind = 'offer' | 'invoice' | 'contract' | 'act';

export interface CaseDocument {
  id: string;
  kind: DocumentKind;
  title: string;
}

export interface CaseSummary {
  id: string;
  title: string;
  createdAt: string;
  status: CaseStatus;
  amountUah: number | null;
  documents: CaseDocument[];
}

export type SlotOrigin = 'system' | 'ai' | 'human';

export type SlotStatus = 'verified' | 'needs_review' | 'not_found';

export interface SourceRef {
  documentId: string;
  documentTitle: string;
  fragment: string;
}

// What the documents say a slot should contain; drives the as-you-type hints.
export interface SlotExpectation {
  value: string;
  reason: string;
}

export interface ProtocolSlot {
  id: string;
  label: string;
  origin: SlotOrigin;
  value: string;
  status: SlotStatus;
  editable: boolean;
  sourceRef: SourceRef | null;
  amount?: { uah: number; inWords: string };
  expected?: SlotExpectation;
  examples?: string[];
}

export interface RuleRef {
  title: string;
  fragment: string;
}

// Tier 2: AI applied rules. Tier 3: deterministic config check (a blocking check in Stage 2).
// In Stage 1 neither blocks printing; both are advisory (PRD.ao-prd.md §3, AC-06).
export interface ProtocolWarning {
  id: string;
  tier: 2 | 3;
  title: string;
  message: string;
  slotId: string | null;
  ruleRef: RuleRef;
}

// Lifecycle states of a protocol draft (CONTEXT.md, PRD.ao-prd.md §10).
export type ProtocolState = 'draft' | 'ready_for_vote' | 'ready_for_signing' | 'signed' | 'withdrawn';

// Tally of one agenda item in a попереднє голосування round.
export interface PollItem {
  id: string;
  title: string;
  for: number;
  against: number;
  abstain: number;
}

// The current раунд опитування; informal, never the meeting's vote (AC-46).
export interface PollRound {
  round: number;
  voted: number;
  boardSize: number;
  items: PollItem[];
}

export interface Withdrawal {
  reason: string;
  date: string;
}

export interface ProtocolDraft {
  caseId: string;
  state: ProtocolState;
  caseTitle: string;
  number: string;
  date: string;
  board: string[];
  documents: CaseDocument[];
  slots: ProtocolSlot[];
  warnings: ProtocolWarning[];
  // Latest round; kept after a return to чернетка for history, shown only while готовий до голосування.
  poll?: PollRound;
  withdrawal?: Withdrawal;
}
