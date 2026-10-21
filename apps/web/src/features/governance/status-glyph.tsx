import { cn } from 'cn';

export type GlyphState = 'done' | 'attention' | 'empty' | 'blocked' | 'neutral';

const tone: Record<GlyphState, string> = {
  done: 'text-success',
  attention: 'text-warning',
  empty: 'text-warning',
  blocked: 'text-destructive',
  neutral: 'text-muted-foreground',
};

// Ring + fill encodes progress, color encodes meaning (StatusGlyph in Figma).
export function StatusGlyph({ state, label, className }: { state: GlyphState; label?: string; className?: string }) {
  const dashed = state === 'empty' || state === 'neutral';
  return (
    <svg viewBox="0 0 14 14" className={cn('size-3.5 shrink-0', tone[state], className)} role={label ? 'img' : undefined} aria-label={label} aria-hidden={label ? undefined : true}>
      <circle cx="7" cy="7" r="6.25" fill="none" stroke="currentColor" strokeWidth="1.5" strokeDasharray={dashed ? '2 2' : undefined} />
      {state === 'done' && <circle cx="7" cy="7" r="7" fill="currentColor" />}
      {(state === 'attention' || state === 'blocked') && <circle cx="7" cy="7" r="3" fill="currentColor" />}
    </svg>
  );
}
