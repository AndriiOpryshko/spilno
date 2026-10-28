import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { cn } from 'cn';
import { useEffect, useMemo, useRef, useState } from 'react';

import { TopBar } from '@/components/app-shell';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Kbd } from '@/components/ui/kbd';
import { Skeleton } from '@/components/ui/skeleton';
import { formatDate, formatUah, uk } from '@/i18n/uk';

import { listCases } from './api';
import { InlineAlert } from './inline-alert';
import { type GlyphState, StatusGlyph } from './status-glyph';
import type { CaseStatus, CaseSummary } from './types';

const filters: Array<CaseStatus | 'all'> = ['all', 'draft', 'review', 'ready', 'printed'];
const glyphFor: Record<CaseStatus, GlyphState> = { draft: 'neutral', review: 'attention', ready: 'done', printed: 'done' };

const docs = (c: CaseSummary) => {
  const kinds = c.documents.map((d) => uk.documentKind[d.kind].toLowerCase());
  const text = kinds.join(', ');
  return text.charAt(0).toUpperCase() + text.slice(1);
};
const amount = (c: CaseSummary) => (c.amountUah === null ? uk.cases.amountNotFound : formatUah(c.amountUah));

export function CasesScreen() {
  const query = useQuery({ queryKey: ['cases'], queryFn: listCases });
  const [search, setSearch] = useState('');
  const [status, setStatus] = useState<CaseStatus | 'all'>('all');
  const searchRef = useRef<HTMLInputElement>(null);

  // "/" focuses search, as the hint in the field says.
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === '/' && document.activeElement?.tagName !== 'INPUT' && document.activeElement?.tagName !== 'TEXTAREA') {
        e.preventDefault();
        searchRef.current?.focus();
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, []);

  const all = useMemo(() => query.data ?? [], [query.data]);
  const visible = useMemo(() => {
    const needle = search.trim().toLowerCase();
    return all.filter((c) => (status === 'all' || c.status === status) && (!needle || c.title.toLowerCase().includes(needle)));
  }, [all, search, status]);
  const count = (f: CaseStatus | 'all') => (f === 'all' ? all.length : all.filter((c) => c.status === f).length);
  const attention = all.filter((c) => c.status === 'draft' || c.status === 'review').length;
  const reset = () => {
    setSearch('');
    setStatus('all');
  };

  return (
    <>
      <TopBar crumbs={[{ label: uk.nav.cases }]} end={<span className="text-[13px] text-muted-foreground lg:hidden">{uk.app.osbbShort}</span>} />
      <div className="space-y-6 px-4 py-6 lg:px-8 lg:py-8">
        <header className="space-y-1">
          <p className="text-xs font-semibold tracking-wide text-muted-foreground uppercase">{uk.cases.eyebrow}</p>
          <h1 className="text-[26px] font-semibold lg:text-[28px]">{uk.cases.title}</h1>
          {query.isSuccess && <p className="text-sm text-muted-foreground">{uk.cases.summary(all.length, attention)}</p>}
        </header>

        <div className="flex flex-col gap-4 lg:flex-row lg:items-center">
          <div className="relative lg:w-80">
            <Input ref={searchRef} aria-label={uk.cases.search} placeholder={uk.cases.search} value={search} onChange={(e) => setSearch(e.target.value)} className="bg-card pr-9" />
            <Kbd className="absolute top-2 right-2">/</Kbd>
          </div>
          <div className="-mx-4 flex gap-1 overflow-x-auto px-4 lg:mx-0 lg:px-0" role="group" aria-label={uk.cases.columns.status}>
            {filters.map((f) => (
              <Button key={f} size="sm" variant={status === f ? 'secondary' : 'outline'} aria-pressed={status === f} onClick={() => setStatus(f)}>
                {f === 'all' ? uk.cases.all : uk.caseStatus[f]}
                {query.isSuccess && <span className="opacity-60">{count(f)}</span>}
              </Button>
            ))}
          </div>
        </div>

        {query.isPending && <CasesSkeleton />}
        {query.isError && (
          <InlineAlert
            title={uk.cases.loadError}
            action={
              <Button variant="outline" onClick={() => query.refetch()}>
                {uk.cases.retry}
              </Button>
            }
          >
            {uk.cases.loadErrorHint}
          </InlineAlert>
        )}
        {query.isSuccess && visible.length === 0 && (
          <div className="flex flex-col items-center gap-2 rounded-lg border bg-card px-6 py-8 text-center">
            <StatusGlyph state="neutral" />
            <p className="font-semibold">{uk.cases.empty}</p>
            <p className="text-[13px] text-muted-foreground">{uk.cases.emptyHint}</p>
            <Button variant="secondary" onClick={reset}>
              {uk.cases.clear}
            </Button>
          </div>
        )}
        {query.isSuccess && visible.length > 0 && (
          <>
            <CasesTable cases={visible} />
            <ul className="space-y-3 md:hidden">
              {visible.map((c) => (
                <li key={c.id}>
                  <CaseCard c={c} />
                </li>
              ))}
            </ul>
          </>
        )}
      </div>
    </>
  );
}

function CasesTable({ cases }: { cases: CaseSummary[] }) {
  const head = 'text-xs font-semibold tracking-wide text-muted-foreground uppercase';
  return (
    <div className="overflow-hidden rounded-lg border bg-card shadow-[0_1px_3px_rgba(22,22,40,0.05)] max-md:hidden">
      <div className="grid grid-cols-[14px_1fr_180px_140px_130px_88px] items-center gap-4 border-b bg-muted px-4 py-2">
        <span />
        <span className={head}>{uk.cases.columns.case}</span>
        <span className={head}>{uk.cases.columns.documents}</span>
        <span className={cn(head, 'text-right')}>{uk.cases.columns.amount}</span>
        <span className={head}>{uk.cases.columns.status}</span>
        <span />
      </div>
      {cases.map((c) => (
        <div key={c.id} className="grid grid-cols-[14px_1fr_180px_140px_130px_88px] items-center gap-4 border-b px-4 py-3 last:border-b-0">
          <StatusGlyph state={glyphFor[c.status]} label={uk.caseStatus[c.status]} />
          <div className="min-w-0">
            <p className="truncate text-sm font-medium">{c.title}</p>
            <p className="text-xs text-muted-foreground">{uk.cases.created(formatDate(c.createdAt))}</p>
          </div>
          <span className="text-[13px] text-muted-foreground">{docs(c)}</span>
          <span className="text-right text-sm font-semibold tabular-nums">{amount(c)}</span>
          <span className="text-[13px] text-muted-foreground">{uk.caseStatus[c.status]}</span>
          <Button size="sm" variant="outline" asChild>
            <Link to="/cases/$caseId/protocol" params={{ caseId: c.id }}>
              {uk.cases.open}
            </Link>
          </Button>
        </div>
      ))}
    </div>
  );
}

function CaseCard({ c }: { c: CaseSummary }) {
  return (
    <Link
      to="/cases/$caseId/protocol"
      params={{ caseId: c.id }}
      className="block space-y-2 rounded-lg border bg-card p-4 outline-none focus-visible:ring-2 focus-visible:ring-ring"
    >
      <div className="flex items-start gap-2">
        <StatusGlyph state={glyphFor[c.status]} className="mt-1" />
        <p className="flex-1 font-semibold">{c.title}</p>
      </div>
      <p className="text-xs text-muted-foreground">
        {uk.cases.created(formatDate(c.createdAt))} · {docs(c)}
      </p>
      <div className="flex items-center justify-between">
        <span className="font-semibold tabular-nums">{amount(c)}</span>
        <span className="text-[13px] text-muted-foreground">{uk.caseStatus[c.status]}</span>
      </div>
    </Link>
  );
}

function CasesSkeleton() {
  return (
    <div className="space-y-2" aria-busy>
      {[0, 1, 2].map((i) => (
        <Skeleton key={i} className="h-14 w-full rounded-lg" />
      ))}
    </div>
  );
}
