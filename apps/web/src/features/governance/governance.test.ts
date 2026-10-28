import { describe, expect, it } from 'vitest';

import { transitionProtocol, updateSlot } from './api';
import { applyTransition, canTransition, lifecycleSteps, transitionError } from './lifecycle';
import { protocols } from './mock-data';
import { getReadiness } from './readiness';
import type { ProtocolDraft, ProtocolState } from './types';
import { validateSlot } from './validation';

const draft = protocols['c-101']!;
const slot = (id: string) => draft.slots.find((s) => s.id === id)!;
const confirmAi = (d: ProtocolDraft): ProtocolDraft => ({
  ...d,
  slots: d.slots.map((s) => (s.origin === 'ai' ? { ...s, status: 'verified' as const } : s)),
});

describe('getReadiness', () => {
  it('blocks only on AI-proposed values nobody confirmed', () => {
    const r = getReadiness(draft);
    expect(r).toMatchObject({ ready: false, total: 5, verified: 3 });
    expect(r.blockingSlots.map((s) => s.id)).toEqual(['s-contractor', 's-funding']);
  });

  it('keeps tier-3 checks as advisory: they alone do not block', () => {
    const r = getReadiness(confirmAi(draft));
    expect(r.ready).toBe(true);
    expect(r.advisoryWarnings.map((w) => w.id)).toEqual(['w-1']);
  });

  it('does not block on an unverified value a person entered', () => {
    const humanEmpty = confirmAi({ ...draft, slots: draft.slots.map((s) => (s.id === 's-decision' ? { ...s, value: '', status: 'not_found' as const } : s)) });
    expect(getReadiness(humanEmpty)).toMatchObject({ ready: true, verified: 4 });
  });

  it('unblocks once a person saves the value (origin becomes human)', () => {
    const corrected = { ...draft, slots: draft.slots.map((s) => (s.origin === 'ai' && s.status !== 'verified' ? { ...s, origin: 'human' as const, status: 'not_found' as const } : s)) };
    expect(getReadiness(corrected).ready).toBe(true);
  });
});

describe('lifecycle transitions', () => {
  const states: ProtocolState[] = [...lifecycleSteps, 'withdrawn'];
  const allowed = new Set(['draft>ready_for_vote', 'draft>ready_for_signing', 'ready_for_vote>draft', 'ready_for_vote>ready_for_signing', 'ready_for_signing>draft']);

  it.each(states.flatMap((from) => states.map((to) => [from, to] as const)))('%s → %s follows PRD §10', (from, to) => {
    expect(canTransition(from, to)).toBe(allowed.has(`${from}>${to}`));
  });

  it('refuses готовий до підпису while an AI value is unconfirmed', () => {
    expect(transitionError(draft, 'ready_for_signing')?.code).toBe('governance.unconfirmed_values');
    expect(transitionError(confirmAi(draft), 'ready_for_signing')).toBeNull();
  });

  it('refuses any move out of a final state', () => {
    expect(() => applyTransition({ ...draft, state: 'signed' }, 'draft')).toThrow(expect.objectContaining({ code: 'governance.transition_not_allowed' }));
    expect(() => applyTransition({ ...draft, state: 'withdrawn' }, 'ready_for_vote')).toThrow();
  });

  it('opens the next poll round with empty tallies and keeps the number', () => {
    const vote = protocols['c-103']!;
    const back = applyTransition(vote, 'draft');
    expect(back).toMatchObject({ state: 'draft', number: vote.number, poll: { round: 2 } });
    const again = applyTransition(back, 'ready_for_vote');
    expect(again.poll).toMatchObject({ round: 3, voted: 0, boardSize: 4 });
    expect(again.poll!.items.every((i) => i.for + i.against + i.abstain === 0)).toBe(true);
  });

  it('opens round 1 from a draft that was never polled', () => {
    expect(applyTransition(draft, 'ready_for_vote').poll).toMatchObject({ round: 1, items: [{ title: draft.caseTitle }] });
  });
});

describe('in-memory api', () => {
  it('rejects a slot edit outside чернетка', async () => {
    await expect(updateSlot('c-103', 's-decision', 'x')).rejects.toThrow('not editable');
  });

  it('rejects a refused transition and keeps the state', async () => {
    await expect(transitionProtocol('c-104', 'ready_for_vote')).rejects.toThrow('not allowed');
    expect(protocols['c-104']!.state).toBe('ready_for_signing');
  });

  it('moves готовий до підпису back to чернетка and keeps the number', async () => {
    const before = protocols['c-104']!;
    const after = await transitionProtocol('c-104', 'draft');
    expect(after).toMatchObject({ state: 'draft', number: before.number });
    protocols['c-104'] = before;
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
