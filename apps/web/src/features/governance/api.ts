import { cases, protocols } from './mock-data';
import type { CaseSummary, ProtocolDraft } from './types';

// In-memory stand-in for the governance API. Replace with generated hooks from
// @/api/gen/endpoints/governance/governance once the contract exists.

const LATENCY_MS = 350;

const delay = <T,>(value: T): Promise<T> =>
  new Promise((resolve) => setTimeout(() => resolve(structuredClone(value)), LATENCY_MS));

export class NotFoundError extends Error {}

export function listCases(): Promise<CaseSummary[]> {
  return delay(cases);
}

export function getProtocol(caseId: string): Promise<ProtocolDraft> {
  const draft = protocols[caseId];
  if (!draft) {
    return new Promise((_, reject) => setTimeout(() => reject(new NotFoundError(caseId)), LATENCY_MS));
  }
  return delay(draft);
}

export function updateSlot(caseId: string, slotId: string, value: string): Promise<ProtocolDraft> {
  const draft = protocols[caseId];
  const slot = draft?.slots.find((s) => s.id === slotId);
  if (!draft || !slot || !slot.editable) {
    return Promise.reject(new Error(`Slot ${slotId} is not editable`));
  }
  slot.value = value;
  slot.origin = 'human';
  slot.status = value.trim() ? 'verified' : 'not_found';
  return delay(draft);
}
