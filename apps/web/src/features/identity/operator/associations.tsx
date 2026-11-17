import { useInfiniteQuery, useQueryClient } from '@tanstack/react-query';
import { useMemo, useState } from 'react';

import {
  getListAssociationsForOperatorQueryKey,
  useDisableAssociation,
  useEnableAssociation,
  listAssociationsForOperator,
} from '@/api/gen/endpoints/identity/identity';
import type { OperatorAssociation } from '@/api/gen/model';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { InlineAlert } from '@/features/governance/inline-alert';
import { PageSkeleton, RequestFallback } from '@/features/identity/shell/request-fallback';
import { uk } from '@/i18n/uk';

const messageFor = (err: unknown) => {
  const code = (err as { info?: { code?: string } }).info?.code;
  if (code === 'identity.reason_required') return uk.operator.reasonRequired;
  if (code === 'identity.association_state_unchanged') return uk.association.stateUnchanged;
  return uk.operator.actionFailed;
};

function AssociationRow({ association, onChanged }: { association: OperatorAssociation; onChanged: () => void }) {
  const [asking, setAsking] = useState(false);
  const [reason, setReason] = useState('');
  const [error, setError] = useState<string | null>(null);
  const disable = useDisableAssociation();
  const enable = useEnableAssociation();
  const connected = association.state === 'connected';

  const settle = {
    onSuccess: () => {
      setAsking(false);
      setReason('');
      setError(null);
      onChanged();
    },
    onError: (err: unknown) => {
      setError(messageFor(err));
      onChanged();
    },
  };

  return (
    <li>
      <div className="flex flex-wrap items-start justify-between gap-3 rounded-md border px-3 py-2">
        <div className="flex flex-col">
          <p className="font-medium">{association.name}</p>
          {association.edrpou && <span className="text-sm text-muted-foreground">{association.edrpou}</span>}
        </div>
        <Badge variant={connected ? 'default' : 'secondary'}>{connected ? uk.operator.stateConnected : uk.operator.stateDisabled}</Badge>
        {connected ? (
          asking ? (
            <form
              className="flex w-full flex-col gap-2"
              onSubmit={(e) => {
                e.preventDefault();
                disable.mutate({ associationId: association.id, data: { reason } }, settle);
              }}
            >
              <Textarea aria-label={uk.operator.reasonLabel} value={reason} onChange={(e) => setReason(e.target.value)} />
              <Button type="submit" variant="destructive" disabled={disable.isPending}>
                {uk.operator.disable}
              </Button>
            </form>
          ) : (
            <Button variant="outline" onClick={() => setAsking(true)}>
              {uk.operator.disable}
            </Button>
          )
        ) : (
          <Button variant="outline" disabled={enable.isPending} onClick={() => enable.mutate({ associationId: association.id }, settle)}>
            {uk.operator.enable}
          </Button>
        )}
        {error && (
          <InlineAlert title={error} className="w-full" />
        )}
      </div>
    </li>
  );
}

export function OperatorAssociations() {
  const client = useQueryClient();
  const query = useInfiniteQuery({
    // Pages are a different cache shape from the plain list read, so the key gets a suffix.
    queryKey: [...getListAssociationsForOperatorQueryKey(), 'pages'],
    queryFn: ({ pageParam }) => listAssociationsForOperator(pageParam ? { cursor: pageParam } : undefined),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (last) => last.data.next_cursor ?? undefined,
  });
  const items = useMemo(() => (query.data?.pages ?? []).flatMap((page) => page.data.items), [query.data]);
  if (query.isError) return <RequestFallback error={query.error} onRetry={() => void query.refetch()} />;
  if (!query.data) return <PageSkeleton />;
  const refresh = () => client.invalidateQueries({ queryKey: getListAssociationsForOperatorQueryKey() });

  return (
    <section className="mx-auto mt-8 flex max-w-2xl flex-col gap-3">
      <h1 className="text-xl font-semibold">{uk.operator.associationsTitle}</h1>
      <ul className="flex flex-col gap-2">
        {items.map((a) => (
          <AssociationRow key={a.id} association={a} onChanged={refresh} />
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
