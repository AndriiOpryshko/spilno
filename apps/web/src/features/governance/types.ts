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

// Tier 2: AI applied rules, never blocks. Tier 3: deterministic config check, blocks printing.
export interface ProtocolWarning {
  id: string;
  tier: 2 | 3;
  title: string;
  message: string;
  slotId: string | null;
  ruleRef: RuleRef;
}

export interface ProtocolDraft {
  caseId: string;
  caseTitle: string;
  number: string;
  date: string;
  board: string[];
  documents: CaseDocument[];
  slots: ProtocolSlot[];
  warnings: ProtocolWarning[];
}
