import { Link } from '@tanstack/react-router';

import type { RegistryDecision } from '@/api/gen/model';
import { Button } from '@/components/ui/button';
import { formatDate, formatUah, uk } from '@/i18n/uk';

import { StatusGlyph } from '../status-glyph';

export type RegistryRow = { kind: 'decision'; decision: RegistryDecision } | { kind: 'gap'; number: number };

// One page of the registry in number order, newest first: each withdrawn number sits where it falls
// between the decisions as a gap, so the sequence of numbers shows no silent hole.
export function pageRows(items: RegistryDecision[], withdrawnNumbers: number[]): RegistryRow[] {
  const rows: RegistryRow[] = [
    ...items.map((decision): RegistryRow => ({ kind: 'decision', decision })),
    ...withdrawnNumbers.map((number): RegistryRow => ({ kind: 'gap', number })),
  ];
  const numberOf = (row: RegistryRow) => (row.kind === 'gap' ? row.number : row.decision.number);
  return rows.sort((a, b) => numberOf(b) - numberOf(a));
}

interface RegistryResultsProps {
  associationId: string;
  rows: RegistryRow[];
  hasMore: boolean;
  loadingMore?: boolean;
  onMore: () => void;
}

export function RegistryResults({ associationId, rows, hasMore, loadingMore = false, onMore }: RegistryResultsProps) {
  return (
    <div className="space-y-4">
      <ul aria-label={uk.registry.list} className="overflow-hidden rounded-lg border bg-card">
        {rows.map((row) => (row.kind === 'gap' ? <GapRow key={`gap-${row.number}`} number={row.number} /> : <DecisionRow key={row.decision.agenda_item_id} associationId={associationId} decision={row.decision} />))}
      </ul>
      {hasMore && (
        <Button variant="outline" disabled={loadingMore} onClick={onMore}>
          {loadingMore ? uk.registry.loadingMore : uk.registry.more}
        </Button>
      )}
    </div>
  );
}

function DecisionRow({ associationId, decision }: { associationId: string; decision: RegistryDecision }) {
  return (
    <li className="flex flex-wrap items-baseline gap-x-4 gap-y-1 border-b px-4 py-3 last:border-b-0">
      <span className="w-16 text-sm font-semibold tabular-nums">{uk.registry.number(decision.number)}</span>
      <span className="w-24 text-[13px] text-muted-foreground tabular-nums">{formatDate(decision.meeting_date)}</span>
      <div className="min-w-0 flex-1 basis-64">
        <Link to="/o/$associationId/registry/$protocolId" params={{ associationId, protocolId: decision.protocol_id }} className="text-sm font-medium underline-offset-2 hover:underline">
          {decision.subject}
        </Link>
        {decision.counterparty && <p className="text-[13px] text-muted-foreground">{decision.counterparty}</p>}
      </div>
      {decision.amount_kopiyky !== null && <span className="text-sm font-semibold tabular-nums">{formatUah(decision.amount_kopiyky / 100)}</span>}
    </li>
  );
}

function GapRow({ number }: { number: number }) {
  return (
    <li className="flex items-center gap-3 border-b bg-muted px-4 py-2 text-[13px] text-foreground last:border-b-0">
      <StatusGlyph state="neutral" />
      <span>{uk.registry.gap(number)}</span>
      <span className="max-md:hidden">{uk.registry.gapHint}</span>
    </li>
  );
}
