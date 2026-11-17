import { useInfiniteQuery } from '@tanstack/react-query';
import { useParams } from '@tanstack/react-router';
import { useMemo, useState } from 'react';

import { getSearchRegistryQueryKey, searchRegistry } from '@/api/gen/endpoints/governance/governance';
import { useListMemberships } from '@/api/gen/endpoints/identity/identity';
import { TopBar } from '@/components/app-shell';
import { Skeleton } from '@/components/ui/skeleton';
import { RequestFallback } from '@/features/identity/shell/request-fallback';
import { uk } from '@/i18n/uk';

import { InlineAlert } from '../inline-alert';
import { StatusGlyph } from '../status-glyph';
import { emptyFilters, type FilterValues, invalidFields, isFiltered, toSearchParams } from './filters';
import { RegistryFilters } from './registry-filters';
import { pageRows, RegistryResults } from './registry-results';

// The реєстр протоколів (AC-26, AC-09): search over the decisions of signed protocols, paged by cursor.
// Governance reads stay open in a вимкнене ОСББ (ADR-0007); the layout shows the suspended notice.
export function RegistryScreen() {
  const { associationId } = useParams({ from: '/o/$associationId' });
  // The layout already loaded the roster; this reads the same cache entry.
  const { data: roster } = useListMemberships(undefined, { query: { refetchOnMount: false } });
  const association = roster?.data.items.find((m) => m.association_id === associationId);

  const [form, setForm] = useState<FilterValues>(emptyFilters);
  const [applied, setApplied] = useState<FilterValues>(emptyFilters);
  const params = useMemo(() => toSearchParams(applied), [applied]);

  const query = useInfiniteQuery({
    queryKey: getSearchRegistryQueryKey(associationId, params),
    queryFn: ({ pageParam }) => searchRegistry(associationId, { ...params, cursor: pageParam }),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (last) => last.data.next_cursor ?? undefined,
    retry: false,
  });

  const rows = useMemo(() => (query.data?.pages ?? []).flatMap((page) => pageRows(page.data.items, page.data.withdrawn_numbers)), [query.data]);
  const refused = invalidFields(query.error);

  return (
    <>
      <TopBar crumbs={[{ label: uk.nav.registry }]} associationName={association?.association_name} />
      <div className="space-y-6 px-4 py-6 lg:px-8 lg:py-8">
        <header className="space-y-1">
          <p className="text-xs font-semibold tracking-wide text-muted-foreground uppercase">{uk.registry.eyebrow}</p>
          <h1 className="text-[26px] font-semibold lg:text-[28px]">{uk.registry.title}</h1>
          <p className="text-sm text-muted-foreground">{uk.registry.intro}</p>
        </header>

        <RegistryFilters values={form} onChange={setForm} onSubmit={() => setApplied(form)} invalid={refused.fields} pending={query.isFetching && !query.isFetchingNextPage} />

        {query.isPending && (
          <div className="space-y-2" aria-busy>
            {[0, 1, 2].map((i) => (
              <Skeleton key={i} className="h-14 w-full rounded-lg" />
            ))}
          </div>
        )}
        {query.isError && refused.invalid && (
          <InlineAlert title={uk.registry.filters.invalidTitle}>{uk.registry.filters.invalidHint}</InlineAlert>
        )}
        {query.isError && !refused.invalid && (
          <RequestFallback error={query.error} onRetry={() => void query.refetch()} />
        )}
        {query.isSuccess && rows.length === 0 && <EmptyRegistry filtered={isFiltered(applied)} />}
        {query.isSuccess && rows.length > 0 && (
          <RegistryResults associationId={associationId} rows={rows} hasMore={query.hasNextPage} loadingMore={query.isFetchingNextPage} onMore={() => query.fetchNextPage()} />
        )}
      </div>
    </>
  );
}

// Nothing to list: either the registry has no signed protocol yet, or the filters match none.
export function EmptyRegistry({ filtered }: { filtered: boolean }) {
  return (
    <div className="flex flex-col items-center gap-2 rounded-lg border bg-card px-6 py-8 text-center">
      <StatusGlyph state="neutral" />
      <p className="font-semibold">{filtered ? uk.registry.noMatch : uk.registry.empty}</p>
      <p className="text-[13px] text-muted-foreground">{filtered ? uk.registry.noMatchHint : uk.registry.emptyHint}</p>
    </div>
  );
}
