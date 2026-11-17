import { useInfiniteQuery } from '@tanstack/react-query';
import { useMemo, useState } from 'react';

import { getListBoardMembersQueryKey, listBoardMembers } from '@/api/gen/endpoints/governance/governance';
import { useGetAssociationCard, useListMemberships } from '@/api/gen/endpoints/identity/identity';
import type { BoardMember } from '@/api/gen/model';
import { TopBar } from '@/components/app-shell';
import { Button } from '@/components/ui/button';
import { RosterDecisionMark } from '@/features/identity/association/roster-decision-mark';
import { PageSkeleton, RequestFallback } from '@/features/identity/shell/request-fallback';
import { formatDate, uk } from '@/i18n/uk';

import { RosterForm } from './roster-form';

const t = uk.association.roster;

type Editing = { member?: BoardMember } | null;

export function RosterScreen({ associationId }: { associationId: string }) {
  const [editing, setEditing] = useState<Editing>(null);
  const roster = useInfiniteQuery({
    // Pages are a different cache shape from the plain list read, so the key gets a suffix.
    queryKey: [...getListBoardMembersQueryKey(associationId), 'pages'],
    queryFn: ({ pageParam }) => listBoardMembers(associationId, pageParam ? { cursor: pageParam } : undefined),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (last) => last.data.next_cursor ?? undefined,
    retry: false,
  });
  const members = useMemo(() => (roster.data?.pages ?? []).flatMap((page) => page.data.items), [roster.data]);
  // The decision history of the card feeds the mark on the entries a decision touched.
  const card = useGetAssociationCard(associationId, { query: { retry: false } });
  // The layout already loaded the roster of memberships; this reads the same cache entry.
  const { data: memberships } = useListMemberships(undefined, { query: { refetchOnMount: false } });
  const associationName = memberships?.data.items.find((m) => m.association_id === associationId)?.association_name;

  if (roster.error) {
    const error = roster.error as { status?: number; info?: { code?: string; detail?: string } };
    // The layout owns the suspended notice of a вимкнене ОСББ.
    if (error.status === 403 && error.info?.code === 'identity.association_disabled') return null;
    return <RequestFallback error={error} onRetry={() => void roster.refetch()} />;
  }
  if (!roster.data) return <PageSkeleton />;

  const history = card.data?.data.history ?? [];
  return (
    <>
      <TopBar crumbs={[{ label: uk.nav.roster }]} associationName={associationName} />
      <section className="mx-auto max-w-2xl space-y-6 px-4 py-8">
        <header className="flex items-center justify-between gap-3">
          <h1 className="text-2xl font-semibold">{t.title}</h1>
          <Button onClick={() => setEditing({})}>{t.add}</Button>
        </header>
        {members.length === 0 ? (
          <p className="text-muted-foreground">{t.empty}</p>
        ) : (
          <ul className="space-y-3">
            {members.map((m) => (
              <li key={m.id} className="space-y-1 rounded-md border p-3">
                <div className="flex items-start justify-between gap-3">
                  <p className="font-medium">{m.full_name}</p>
                  <Button variant="outline" size="sm" aria-label={t.editEntry(m.full_name)} onClick={() => setEditing({ member: m })}>
                    {t.edit}
                  </Button>
                </div>
                <p className="text-sm">{t.position[m.position]}</p>
                <p className="text-sm text-muted-foreground">{t.term(formatDate(m.term_start), m.term_end ? formatDate(m.term_end) : undefined)}</p>
                {m.decision_id && <RosterDecisionMark kind={m.decision_effect === 'ended' ? 'ended' : 'appointed'} decisionId={m.decision_id} history={history} associationId={associationId} />}
              </li>
            ))}
          </ul>
        )}
        {roster.hasNextPage && (
          <Button variant="outline" disabled={roster.isFetchingNextPage} onClick={() => void roster.fetchNextPage()}>
            {roster.isFetchingNextPage ? t.loadingMore : t.more}
          </Button>
        )}
        {editing && <RosterForm associationId={associationId} member={editing.member} onClose={() => setEditing(null)} />}
      </section>
    </>
  );
}
