import { applyTransition } from './lifecycle';
import { cases, protocols } from './mock-data';
import type { CaseSummary, ProtocolDraft, ProtocolState } from './types';

// In-memory stand-in for the governance API. Replace with generated hooks from
// @/api/gen/endpoints/governance/governance once the contract exists.

const LATENCY_MS = 350;

const delay = <T,>(value: T): Promise<T> =>
  new Promise((resolve) => setTimeout(() => resolve(structuredClone(value)), LATENCY_MS));

const fail = (error: Error): Promise<never> => new Promise((_, reject) => setTimeout(() => reject(error), LATENCY_MS));

export class NotFoundError extends Error {}

export function listCases(): Promise<CaseSummary[]> {
  return delay(cases);
}

export function getProtocol(caseId: string): Promise<ProtocolDraft> {
  const draft = protocols[caseId];
  if (!draft) return fail(new NotFoundError(caseId));
  return delay(draft);
}

// Only чернетка is editable; every other state is frozen (CONTEXT.md).
export function updateSlot(caseId: string, slotId: string, value: string): Promise<ProtocolDraft> {
  const draft = protocols[caseId];
  const slot = draft?.slots.find((s) => s.id === slotId);
  if (!draft || draft.state !== 'draft' || !slot || !slot.editable) {
    return Promise.reject(new Error(`Slot ${slotId} is not editable`));
  }
  slot.value = value;
  slot.origin = 'human';
  slot.status = value.trim() ? 'verified' : 'not_found';
  return delay(draft);
}

// Moves the protocol along its lifecycle; a refused transition rejects and keeps the state.
export function transitionProtocol(caseId: string, to: ProtocolState): Promise<ProtocolDraft> {
  const draft = protocols[caseId];
  if (!draft) return fail(new NotFoundError(caseId));
  try {
    protocols[caseId] = applyTransition(draft, to);
  } catch (error) {
    return fail(error as Error);
  }
  return delay(protocols[caseId]);
}
