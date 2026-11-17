import { useInfiniteQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { useMemo } from 'react';

import { getListApplicationsForReviewQueryKey, listApplicationsForReview } from '@/api/gen/endpoints/identity/identity';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { RequestFallback } from '@/features/identity/shell/request-fallback';
import { uk } from '@/i18n/uk';

export function OperatorQueue() {
  const query = useInfiniteQuery({
    // Pages are a different cache shape from the plain list read, so the key gets a suffix.
    queryKey: [...getListApplicationsForReviewQueryKey(), 'pages'],
    queryFn: ({ pageParam }) => listApplicationsForReview(pageParam ? { cursor: pageParam } : undefined),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (last) => last.data.next_cursor ?? undefined,
  });
  const items = useMemo(() => (query.data?.pages ?? []).flatMap((page) => page.data.items), [query.data]);
  // A 403 is left to the section guard; any other failed read shows the fallback, never a blank page.
  if (query.isError) {
    if ((query.error as { status?: number } | null)?.status === 403) return null;
    return <RequestFallback error={query.error} onRetry={() => void query.refetch()} />;
  }
  if (!query.data) return null;
  return (
    <section className="mx-auto mt-8 flex max-w-2xl flex-col gap-3">
      <h1 className="text-xl font-semibold">{uk.operator.queueTitle}</h1>
      <ul className="flex flex-col gap-2">
        {items.map((a) => (
          <li key={a.id} className="flex items-center justify-between gap-3 rounded-md border px-3 py-2">
            <Link to="/operator/applications/$applicationId" params={{ applicationId: a.id }} className="font-medium">
              {a.association_name}
            </Link>
            <div className="flex items-center gap-2">
              {a.competing && (
                <Badge variant="secondary">{a.kind === 'transfer' ? uk.operator.competingTransfer : uk.operator.competingConnection}</Badge>
              )}
              <Badge variant="secondary">{a.kind === 'transfer' ? uk.operator.kindTransfer : uk.operator.kindConnection}</Badge>
            </div>
          </li>
        ))}
      </ul>
      {query.hasNextPage && (
        <Button variant="outline" disabled={query.isFetchingNextPage} onClick={() => void query.fetchNextPage()}>
          {uk.operator.more}
        </Button>
      )}
    </section>
  );
}
