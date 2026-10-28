import { cn } from 'cn';

import { uk } from '@/i18n/uk';

import { getReadiness, isUnconfirmedAi } from './readiness';
import { type GlyphState, StatusGlyph } from './status-glyph';
import type { ProtocolDraft } from './types';

// Print gate (ReadinessCard in Figma). One segment per slot. Only unconfirmed AI-proposed values block
// (BR-11); tier-3 checks are listed as advisory attention. Printing itself lives in TransitionActions.
export function ReadinessCard({ draft }: { draft: ProtocolDraft }) {
  const readiness = getReadiness(draft);
  const label = (slotId: string | null) => draft.slots.find((s) => s.id === slotId)?.label;

  const items: { state: GlyphState; text: string }[] = readiness.blockingSlots.map((s) =>
    s.status === 'not_found' ? { state: 'empty', text: uk.protocol.missing(s.label) } : { state: 'attention', text: uk.protocol.needsReview(s.label) },
  );
  if (readiness.ready && readiness.advisoryWarnings.length > 0) items.push({ state: 'done', text: uk.protocol.allConfirmed });
  for (const w of readiness.advisoryWarnings) items.push({ state: 'attention', text: uk.protocol.needsMeeting(label(w.slotId) ?? w.title) });

  // Grouped like a progress bar: confirmed, then waiting for a person, then the rest.
  const rank = draft.slots.map((s) => (s.status === 'verified' ? 0 : isUnconfirmedAi(s) ? 1 : 2)).sort((a, b) => a - b);
  const tones = ['bg-success', 'bg-warning', 'bg-border'];

  return (
    <section aria-label={uk.protocol.readiness} className="space-y-3 rounded-lg border bg-card p-4 shadow-[0_1px_3px_rgba(22,22,40,0.06),0_8px_24px_-4px_rgba(22,22,40,0.05)]">
      <div className="flex items-center justify-between text-sm font-semibold">
        <h2>{uk.protocol.readiness}</h2>
        <span className="text-muted-foreground">
          {readiness.verified} / {readiness.total}
        </span>
      </div>
      <div className="flex gap-1">
        {rank.map((r, i) => (
          <span key={i} className={cn('h-1 flex-1 rounded-full', tones[r])} />
        ))}
      </div>
      <ul className="space-y-2 text-[13px]">
        {items.length === 0 ? (
          <li className="flex items-center gap-2">
            <StatusGlyph state="done" /> {uk.protocol.allVerified}
          </li>
        ) : (
          items.map((i) => (
            <li key={i.text} className="flex items-center gap-2">
              <StatusGlyph state={i.state} /> {i.text}
            </li>
          ))
        )}
      </ul>
    </section>
  );
}
