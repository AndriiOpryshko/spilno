import type { ProtocolSlot } from './types';

export type SlotHint =
  | { kind: 'suggestion'; message: string; fix: string }
  | { kind: 'examples'; message: string; examples: string[] }
  | { kind: 'required'; message: string };

const normalize = (s: string) => s.replace(/\s+/g, ' ').trim().toLowerCase();

// Runs on every keystroke. Hints guide the chair; they never block saving.
export function validateSlot(slot: ProtocolSlot, value: string): SlotHint | null {
  if (!value.trim()) {
    if (slot.examples?.length) {
      return { kind: 'examples', message: 'Оберіть варіант або впишіть свій', examples: slot.examples };
    }
    return { kind: 'required', message: 'Поле потрібне для протоколу' };
  }
  if (slot.expected && normalize(value) !== normalize(slot.expected.value)) {
    return { kind: 'suggestion', message: slot.expected.reason, fix: slot.expected.value };
  }
  return null;
}
