import { useInfiniteQuery } from '@tanstack/react-query';
import { Link, useParams } from '@tanstack/react-router';
import { useMemo } from 'react';

import { getListCasesQueryKey, listCases, useGetCase } from '@/api/gen/endpoints/governance/governance';
import { useAllMemberships } from '@/features/identity/shell/use-all-memberships';
import type { Case } from '@/api/gen/model';
import { TopBar } from '@/components/app-shell';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { RequestFallback } from '@/features/identity/shell/request-fallback';
import { formatDate, uk } from '@/i18n/uk';

import { CaseDocuments } from './case-documents';
import { StatusGlyph } from './status-glyph';

// A вимкнене ОСББ refuses every governance read (ADR-0007); the layout shows the suspended notice,
// so the screens below show no справи and no load error for it.
const isDisabledProblem = (error: unknown) => {
  const { status, info } = (error ?? {}) as { status?: number; info?: { code?: string } };
  return status === 403 && info?.code === 'identity.association_disabled';
};

function useAssociationName(associationId: string) {
  // The layout already loaded the roster; this reads the same cache entry.
  const { data: roster } = useAllMemberships({ refetchOnMount: false });
  return roster?.data.items.find((m) => m.association_id === associationId)?.association_name;
}

export function CasesScreen() {
  const { associationId } = useParams({ from: '/o/$associationId' });
  const associationName = useAssociationName(associationId);
  // Paged by next_cursor, like the реєстр: «показати ще» reads the next page.
  const query = useInfiniteQuery({
    // Pages are a different cache shape from the layout's plain list read, so the key gets a suffix.
    queryKey: [...getListCasesQueryKey(associationId), 'pages'],
    queryFn: ({ pageParam }) => listCases(associationId, pageParam ? { cursor: pageParam } : undefined),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (last) => last.data.next_cursor ?? undefined,
    retry: false,
  });
  const items = useMemo(() => (query.data?.pages ?? []).flatMap((page) => page.data.items), [query.data]);

  return (
    <>
      <TopBar crumbs={[{ label: uk.nav.cases }]} associationName={associationName} />
      <div className="space-y-6 px-4 py-6 lg:px-8 lg:py-8">
        <header className="space-y-1">
          <p className="text-xs font-semibold tracking-wide text-muted-foreground uppercase">{uk.cases.eyebrow}</p>
          <h1 className="text-[26px] font-semibold lg:text-[28px]">{uk.cases.title}</h1>
        </header>

        {query.isPending && <CasesSkeleton />}
        {query.isError && !isDisabledProblem(query.error) && (
          <RequestFallback error={query.error} onRetry={() => void query.refetch()} />
        )}
        {query.isSuccess && items.length === 0 && (
          <div className="flex flex-col items-center gap-2 rounded-lg border bg-card px-6 py-8 text-center">
            <StatusGlyph state="neutral" />
            <p className="font-semibold">{uk.cases.emptyList}</p>
            <p className="text-[13px] text-muted-foreground">{uk.cases.emptyListHint}</p>
          </div>
        )}
        {query.isSuccess && items.length > 0 && (
          <ul className="overflow-hidden rounded-lg border bg-card shadow-[0_1px_3px_rgba(22,22,40,0.05)]">
            {items.map((c) => (
              <li key={c.id} className="border-b last:border-b-0">
                <CaseRow c={c} associationId={associationId} />
              </li>
            ))}
          </ul>
        )}
        {query.isSuccess && query.hasNextPage && (
          <Button variant="outline" disabled={query.isFetchingNextPage} onClick={() => query.fetchNextPage()}>
            {query.isFetchingNextPage ? uk.cases.loadingMore : uk.cases.more}
          </Button>
        )}
      </div>
    </>
  );
}

function CaseRow({ c, associationId }: { c: Case; associationId: string }) {
  return (
    <Link
      to="/o/$associationId/cases/$caseId"
      params={{ associationId, caseId: c.id }}
      className="block space-y-1 px-4 py-3 outline-none hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring"
    >
      <span className="block text-sm font-medium">{c.title}</span>
      <span className="block text-xs text-muted-foreground">{uk.cases.created(formatDate(c.created_at))}</span>
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

// One справа with its documents and their versions (AC-09); the way on to its protocol draft.
export function CaseScreen({ caseId }: { caseId: string }) {
  const { associationId } = useParams({ from: '/o/$associationId' });
  const associationName = useAssociationName(associationId);
  const query = useGetCase(associationId, caseId, { query: { retry: false } });
  const detail = query.data?.data;
  const disabled = isDisabledProblem(query.error);

  return (
    <>
      <TopBar
        crumbs={[{ label: uk.nav.cases, to: `/o/${associationId}/cases` }, ...(detail ? [{ label: detail.title }] : [])]}
        associationName={associationName}
      />
      <div className="space-y-6 px-4 py-6 lg:px-8 lg:py-8">
        <header className="space-y-1">
          {detail && <h1 className="text-[26px] font-semibold lg:text-[28px]">{detail.title}</h1>}
          {detail && <p className="text-sm text-muted-foreground">{uk.cases.created(formatDate(detail.created_at))}</p>}
          <p className="flex flex-wrap gap-4 text-sm">
            <Link className="underline" to="/o/$associationId/cases" params={{ associationId }}>
              {uk.cases.backToList}
            </Link>
            {detail && (
              <Link className="underline" to="/o/$associationId/cases/$caseId/protocol" params={{ associationId, caseId }}>
                {uk.cases.openProtocol}
              </Link>
            )}
          </p>
        </header>

        {query.isPending && <CasesSkeleton />}
        {query.isError && !disabled && (
          <RequestFallback
            error={query.error}
            onRetry={() => void query.refetch()}
            action={
              <Button variant="outline" asChild>
                <Link to="/o/$associationId/cases" params={{ associationId }}>
                  {uk.cases.backToList}
                </Link>
              </Button>
            }
          />
        )}
        {detail && <CaseDocuments associationId={associationId} caseId={caseId} documents={detail.documents} />}
      </div>
    </>
  );
}
