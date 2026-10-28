import { cn } from 'cn';

import { uk } from '@/i18n/uk';

import { lifecycleSteps } from './lifecycle';
import { ProtocolStateBadge } from './protocol-state-badge';
import type { ProtocolState } from './types';

interface LifecycleStepperProps {
  state: ProtocolState;
  // Force the compact "Крок N з 4" view at any width (below 768 px it is always compact).
  compact?: boolean;
  className?: string;
}

// чернетка → готовий до голосування → готовий до підпису → підписаний; знятий з розгляду is an exit
// shown after the muted path (LifecycleStepper in Figma). Read-only, never focusable.
export function LifecycleStepper({ state, compact = false, className }: LifecycleStepperProps) {
  const withdrawn = state === 'withdrawn';
  const current = withdrawn ? -1 : lifecycleSteps.indexOf(state);
  const total = lifecycleSteps.length;

  return (
    <div className={className}>
      <ol aria-label={uk.protocol.lifecycle} className={cn('flex flex-wrap items-center gap-x-2 gap-y-1 text-[13px]', compact ? 'sr-only' : 'max-md:sr-only')}>
        {lifecycleSteps.map((step, i) => {
          const done = i < current;
          const active = i === current;
          return (
            <li key={step} aria-current={active ? 'step' : undefined} className="flex items-center gap-2">
              {i > 0 && <span aria-hidden className={cn('h-px w-6', i <= current ? 'bg-primary' : 'bg-border')} />}
              <StepMark n={i + 1} done={done} active={active} />
              <span className={cn(active ? 'font-semibold text-foreground' : done ? 'text-foreground' : 'text-muted-foreground')}>
                {uk.protocol.state[step]}
                {done && <span className="sr-only"> — {uk.protocol.stepDone}</span>}
              </span>
            </li>
          );
        })}
        {withdrawn && (
          <li aria-current="step" className="flex items-center">
            <ProtocolStateBadge state="withdrawn" />
          </li>
        )}
      </ol>

      <div aria-hidden className={cn('space-y-2', compact ? '' : 'md:hidden')}>
        <p className="text-[13px] font-semibold">
          {withdrawn ? uk.protocol.state.withdrawn : `${uk.protocol.stepOf(current + 1, total)} · ${uk.protocol.state[lifecycleSteps[current]!]}`}
        </p>
        <div className="flex gap-1">
          {lifecycleSteps.map((step, i) => (
            <span key={step} className={cn('h-1 flex-1 rounded-full', i <= current ? 'bg-primary' : 'bg-border')} />
          ))}
        </div>
      </div>
    </div>
  );
}

function StepMark({ n, done, active }: { n: number; done: boolean; active: boolean }) {
  if (done) {
    return (
      <svg aria-hidden viewBox="0 0 20 20" className="size-5 shrink-0 text-primary">
        <circle cx="10" cy="10" r="10" fill="currentColor" />
        <path d="M6 10.5l2.5 2.5L14 7.5" fill="none" stroke="var(--primary-foreground)" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" />
      </svg>
    );
  }
  return (
    <span
      aria-hidden
      className={cn(
        'flex size-5 shrink-0 items-center justify-center rounded-full text-[11px] font-semibold',
        active ? 'bg-primary text-primary-foreground' : 'border border-border text-muted-foreground',
      )}
    >
      {n}
    </span>
  );
}
