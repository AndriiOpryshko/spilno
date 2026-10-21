import { describe, expect, it } from 'vitest';

import { protocols } from './mock-data';
import { getReadiness } from './readiness';
import { validateSlot } from './validation';

const draft = protocols['c-101']!;
const slot = (id: string) => draft.slots.find((s) => s.id === id)!;

describe('getReadiness', () => {
  it('counts tier-3 checks and unverified slots', () => {
    expect(getReadiness(draft)).toEqual({ ready: false, total: 5, verified: 2, blockingWarnings: 1, unresolvedSlots: 2 });
  });

  it('is ready once nothing blocks', () => {
    const clean = { ...draft, warnings: [], slots: draft.slots.map((s) => ({ ...s, status: 'verified' as const })) };
    expect(getReadiness(clean).ready).toBe(true);
  });
});

describe('validateSlot', () => {
  it('suggests the contract name while the value differs', () => {
    expect(validateSlot(slot('s-contractor'), 'Дах-Сервіс')).toEqual({
      kind: 'suggestion',
      message: expect.stringContaining('ТОВ «Дах-Сервіс»'),
      fix: 'ТОВ «Дах-Сервіс»',
    });
  });

  it('accepts the expected value regardless of spacing and case', () => {
    expect(validateSlot(slot('s-contractor'), '  тов  «дах-сервіс» ')).toBeNull();
  });

  it('offers examples for an empty slot that has them', () => {
    expect(validateSlot(slot('s-funding'), '')).toMatchObject({ kind: 'examples', examples: ['Фонд ремонту', 'Внески співвласників', 'Залишок 2025 року'] });
  });

  it('asks for a value when an empty slot has no examples', () => {
    expect(validateSlot(slot('s-decision'), ' ')).toEqual({ kind: 'required', message: 'Поле потрібне для протоколу' });
  });
});
