import { cn } from 'cn';

import { Button } from '@/components/ui/button';
import { uk } from '@/i18n/uk';

import { getReadiness } from './readiness';
import { type GlyphState, StatusGlyph } from './status-glyph';
import type { ProtocolDraft } from './types';

// Print gate (ReadinessCard in Figma). One segment per slot.
export function ReadinessCard({ draft }: { draft: ProtocolDraft }) {
  const readiness = getReadiness(draft);
  const blocked = new Set(draft.warnings.filter((w) => w.tier === 3).map((w) => w.slotId));
  const items: { state: GlyphState; text: string }[] = [];
  for (const s of draft.slots) {
    if (blocked.has(s.id)) items.push({ state: 'blocked', text: uk.protocol.needsMeeting(s.label) });
    else if (s.status === 'needs_review') items.push({ state: 'attention', text: uk.protocol.needsReview(s.label) });
    else if (s.status === 'not_found') items.push({ state: 'empty', text: uk.protocol.missing(s.label) });
  }

  // Grouped like a progress bar: done, then needs a human, then blocked.
  const rank = (id: string, verified: boolean) => (blocked.has(id) ? 2 : verified ? 0 : 1);
  const segments = draft.slots
    .map((s) => rank(s.id, s.status === 'verified'))
    .sort((a, b) => a - b)
    .map((r) => ['bg-success', 'bg-warning', 'bg-destructive'][r]);

  return (
    <section aria-label={uk.protocol.readiness} className="space-y-3 rounded-lg border bg-card p-4 shadow-[0_1px_3px_rgba(22,22,40,0.06),0_8px_24px_-4px_rgba(22,22,40,0.05)]">
      <div className="flex items-center justify-between text-sm font-semibold">
        <h2>{uk.protocol.readiness}</h2>
        <span className="text-muted-foreground">
          {readiness.verified} / {readiness.total}
        </span>
      </div>
      <div className="flex gap-1">
        {segments.map((tone, i) => (
          <span key={i} className={cn('h-1 flex-1 rounded-full', tone)} />
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
      <Button className="w-full" disabled={!readiness.ready} onClick={() => window.print()}>
        {uk.protocol.print}
      </Button>
    </section>
  );
}
