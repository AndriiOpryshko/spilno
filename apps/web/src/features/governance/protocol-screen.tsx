import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { useEffect, useRef, useState } from 'react';

import { TopBar } from '@/components/app-shell';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { formatDate, uk } from '@/i18n/uk';

import { getProtocol, transitionProtocol } from './api';
import { InlineAlert } from './inline-alert';
import { LifecycleStepper } from './lifecycle-stepper';
import { PollRoundPanel } from './poll-round-panel';
import { ProtocolStateBadge } from './protocol-state-badge';
import { getReadiness } from './readiness';
import { ReadinessCard } from './readiness-card';
import { SlotRow } from './slot-row';
import { StatusGlyph } from './status-glyph';
import { TransitionActions } from './transition-actions';
import type { ProtocolDraft, ProtocolState } from './types';

export function ProtocolScreen({ caseId }: { caseId: string }) {
  const query = useQuery({ queryKey: ['protocol', caseId], queryFn: () => getProtocol(caseId), retry: false });
  const [savedJustNow, setSavedJustNow] = useState(false);
  const [announcement, setAnnouncement] = useState('');
  const draft = query.data;

  const crumbs = [{ label: uk.nav.cases, to: '/cases' }, ...(draft ? [{ label: draft.caseTitle }, { label: uk.protocol.breadcrumb(draft.number) }] : [])];
  const saved = draft && (
    <span className="flex items-center gap-1.5 text-xs text-muted-foreground">
      <StatusGlyph state="done" /> {savedJustNow ? uk.protocol.savedJustNow : uk.protocol.saved}
    </span>
  );

  return (
    <>
      <TopBar crumbs={crumbs} end={saved} />
      {/* Polite live region: present before any change so screen readers pick up the new state. */}
      <p aria-live="polite" className="sr-only">
        {announcement}
      </p>
      <div className="px-4 py-6 lg:px-8 lg:py-8">
        {query.isPending && <ProtocolSkeleton />}
        {query.isError && <Unavailable caseId={caseId} />}
        {draft && (
          <ProtocolBody
            draft={draft}
            onSaved={() => setSavedJustNow(true)}
            onTransitioned={(state) => {
              setSavedJustNow(true);
              setAnnouncement(uk.protocol.stateAnnounce(uk.protocol.state[state]));
            }}
          />
        )}
      </div>
    </>
  );
}

interface ProtocolBodyProps {
  draft: ProtocolDraft;
  onSaved: () => void;
  onTransitioned: (state: ProtocolState) => void;
}

function ProtocolBody({ draft, onSaved, onTransitioned }: ProtocolBodyProps) {
  const queryClient = useQueryClient();
  const actionsRef = useRef<HTMLDivElement>(null);
  const focusActions = useRef(false);
  const readiness = getReadiness(draft);
  const toCheck = readiness.total - readiness.verified;
  const warningsFor = (slotId: string) => draft.warnings.filter((w) => w.slotId === slotId);
  const isDraft = draft.state === 'draft';

  const transition = useMutation({
    mutationFn: (to: ProtocolState) => transitionProtocol(draft.caseId, to),
    onSuccess: (next) => {
      queryClient.setQueryData(['protocol', draft.caseId], next);
      onTransitioned(next.state);
      focusActions.current = true;
    },
  });

  // After a transition the clicked button is gone; move focus to the first action of the new state.
  useEffect(() => {
    if (!focusActions.current) return;
    focusActions.current = false;
    actionsRef.current?.querySelector<HTMLButtonElement>('button:not(:disabled)')?.focus();
  }, [draft.state]);

  const intro =
    draft.state === 'draft'
      ? uk.protocol.intro(toCheck)
      : draft.state === 'withdrawn'
        ? uk.protocol.stateIntro.withdrawn(draft.withdrawal?.reason ?? '', draft.withdrawal ? formatDate(draft.withdrawal.date) : '')
        : uk.protocol.stateIntro[draft.state];
  const poll = draft.state === 'ready_for_vote' && draft.poll ? <PollRoundPanel poll={draft.poll} /> : null;

  return (
    <div className="grid items-start gap-4 lg:grid-cols-[1fr_320px] lg:gap-8">
      <div className="min-w-0 space-y-5">
        <header className="space-y-2">
          <p className="text-xs font-semibold tracking-wide text-muted-foreground uppercase">{uk.protocol.eyebrow}</p>
          <ProtocolStateBadge state={draft.state} />
          <h1 className="text-[22px] leading-7 font-semibold lg:text-[28px] lg:leading-9">{uk.protocol.title(draft.number, draft.caseTitle)}</h1>
          <p className="text-sm text-muted-foreground">{intro}</p>
          <LifecycleStepper state={draft.state} className="pt-1" />
        </header>
        <div className="lg:hidden">
          <ReadinessCard draft={draft} />
        </div>
        <TransitionActions
          ref={actionsRef}
          state={draft.state}
          number={draft.number}
          blocking={readiness.blockingSlots.map((s) => s.label)}
          pending={transition.isPending ? (transition.variables ?? null) : null}
          error={transition.isError}
          onTransition={(to) => transition.mutate(to)}
          onPrint={() => window.print()}
        />
        <section className="overflow-hidden rounded-lg border bg-card shadow-[0_1px_3px_rgba(22,22,40,0.05)]">
          {draft.slots.map((s) => (
            <SlotRow key={s.id} caseId={draft.caseId} slot={s} warnings={warningsFor(s.id)} onSaved={onSaved} readOnly={!isDraft} />
          ))}
        </section>
        <div className="space-y-4 lg:hidden">
          {poll}
          <Properties draft={draft} />
        </div>
      </div>
      <aside className="sticky top-6 hidden space-y-4 lg:block">
        <ReadinessCard draft={draft} />
        {poll}
        <Properties draft={draft} />
      </aside>
    </div>
  );
}

function Properties({ draft }: { draft: ProtocolDraft }) {
  const rows: [string, string][] = [
    [uk.protocol.props.status, uk.protocol.state[draft.state]],
    [uk.protocol.props.date, formatDate(draft.date)],
    [uk.protocol.props.board, draft.board.join(', ')],
    [uk.protocol.props.documents, draft.documents.map((d) => d.title).join(' · ')],
  ];
  return (
    <dl className="overflow-hidden rounded-lg border bg-card text-[13px]">
      {rows.map(([k, v]) => (
        <div key={k} className="grid grid-cols-[90px_1fr] gap-3 border-b px-4 py-3 last:border-b-0">
          <dt className="text-muted-foreground">{k}</dt>
          <dd className="font-medium">{v}</dd>
        </div>
      ))}
    </dl>
  );
}

function ProtocolSkeleton() {
  return (
    <div className="grid gap-8 lg:grid-cols-[1fr_320px]" aria-busy>
      <div className="space-y-4">
        <Skeleton className="h-8 w-2/3" />
        <Skeleton className="h-64 w-full rounded-lg" />
      </div>
      <Skeleton className="hidden h-56 rounded-lg lg:block" />
    </div>
  );
}

function Unavailable({ caseId }: { caseId: string }) {
  return (
    <InlineAlert
      tone="neutral"
      title={uk.protocol.unavailable}
      action={
        <Button variant="outline" asChild>
          <Link to="/cases">{uk.protocol.backToCases}</Link>
        </Button>
      }
    >
      {uk.protocol.unavailableHint(caseId)}
    </InlineAlert>
  );
}
