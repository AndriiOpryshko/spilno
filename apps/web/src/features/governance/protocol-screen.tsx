import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { useState } from 'react';

import { TopBar } from '@/components/app-shell';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { formatDate, uk } from '@/i18n/uk';

import { getProtocol } from './api';
import { getReadiness } from './readiness';
import { ReadinessCard } from './readiness-card';
import { SlotRow } from './slot-row';
import { StatusGlyph } from './status-glyph';
import type { ProtocolDraft } from './types';

export function ProtocolScreen({ caseId }: { caseId: string }) {
  const query = useQuery({ queryKey: ['protocol', caseId], queryFn: () => getProtocol(caseId), retry: false });
  const [savedJustNow, setSavedJustNow] = useState(false);
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
      <div className="px-4 py-6 lg:px-8 lg:py-8">
        {query.isPending && <ProtocolSkeleton />}
        {query.isError && <Unavailable caseId={caseId} />}
        {draft && <ProtocolBody draft={draft} onSaved={() => setSavedJustNow(true)} />}
      </div>
    </>
  );
}

function ProtocolBody({ draft, onSaved }: { draft: ProtocolDraft; onSaved: () => void }) {
  const readiness = getReadiness(draft);
  const toCheck = readiness.total - readiness.verified;
  const warningsFor = (slotId: string) => draft.warnings.filter((w) => w.slotId === slotId);

  return (
    <div className="grid items-start gap-4 lg:grid-cols-[1fr_320px] lg:gap-8">
      <div className="space-y-5">
        <header className="space-y-1">
          <p className="text-xs font-semibold tracking-wide text-muted-foreground uppercase">{uk.protocol.eyebrow}</p>
          <h1 className="text-[22px] leading-7 font-semibold lg:text-[28px] lg:leading-9">{uk.protocol.title(draft.number, draft.caseTitle)}</h1>
          <p className="text-sm text-muted-foreground">{uk.protocol.intro(toCheck)}</p>
        </header>
        <div className="lg:hidden">
          <ReadinessCard draft={draft} />
        </div>
        <section className="overflow-hidden rounded-lg border bg-card shadow-[0_1px_3px_rgba(22,22,40,0.05)]">
          {draft.slots.map((s) => (
            <SlotRow key={s.id} caseId={draft.caseId} slot={s} warnings={warningsFor(s.id)} onSaved={onSaved} />
          ))}
        </section>
        <div className="lg:hidden">
          <Properties draft={draft} ready={readiness.ready} />
        </div>
      </div>
      <aside className="sticky top-6 hidden space-y-4 lg:block">
        <ReadinessCard draft={draft} />
        <Properties draft={draft} ready={readiness.ready} />
      </aside>
    </div>
  );
}

function Properties({ draft, ready }: { draft: ProtocolDraft; ready: boolean }) {
  const rows: [string, string][] = [
    [uk.protocol.props.status, ready ? uk.caseStatus.ready : uk.caseStatus.review],
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
    <div role="alert" className="flex flex-wrap items-start gap-3 rounded-lg bg-destructive-subtle px-4 py-3">
      <StatusGlyph state="neutral" className="mt-1" />
      <div className="min-w-0 flex-1">
        <p className="font-semibold">{uk.protocol.unavailable}</p>
        <p className="text-[13px] text-muted-foreground">{uk.protocol.unavailableHint(caseId)}</p>
      </div>
      <Button variant="outline" asChild>
        <Link to="/cases">{uk.protocol.backToCases}</Link>
      </Button>
    </div>
  );
}
