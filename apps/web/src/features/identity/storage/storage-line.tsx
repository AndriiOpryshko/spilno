import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Link, useNavigate } from '@tanstack/react-router';
import { useEffect, useRef } from 'react';

import { completeStorageConnection, getGetStorageConnectionQueryKey, getStartStorageConnectionUrl, useGetStorageConnection, useListMemberships } from '@/api/gen/endpoints/identity/identity';
import { Skeleton } from '@/components/ui/skeleton';
import { InlineAlert } from '@/features/governance/inline-alert';
import { isNoAccessProblem } from '@/features/identity/shell/entry';
import { RequestFallback } from '@/features/identity/shell/request-fallback';
import { uk } from '@/i18n/uk';

export const storageFlowKey = 'storage-flow-association-id';

export function StorageLine({ associationId }: { associationId: string }) {
  const connection = useGetStorageConnection(associationId);
  if (connection.isError) return <RequestFallback error={connection.error} onRetry={() => void connection.refetch()} />;
  if (!connection.data) return <Skeleton className="h-5 w-48" />;
  const { connected, google_account_email: email } = connection.data.data;
  if (connected) return <p>{uk.storage.line(email ?? '')}</p>;
  return (
    <p className="space-x-2">
      <span>{uk.storage.lineNotConnected}</span>
      <a className="underline" href={getStartStorageConnectionUrl(associationId)} onClick={() => sessionStorage.setItem(storageFlowKey, associationId)}>
        {uk.storage.connect}
      </a>
    </p>
  );
}

export function StorageCallback({ associationId, code, state }: { associationId: string; code: string; state: string }) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const sent = useRef(false);
  const complete = useMutation({
    mutationFn: () => completeStorageConnection(associationId, { code, state }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: getGetStorageConnectionQueryKey(associationId) });
      await navigate({ to: '/o/$associationId/association', params: { associationId } });
    },
  });
  const { mutate } = complete;
  useEffect(() => {
    if (sent.current) return;
    sent.current = true;
    mutate();
  }, [mutate]);

  if (!complete.isError) return <Skeleton className="m-8 h-24 max-w-md" />;
  const { status, info } = complete.error as { status?: number; info?: { code?: string } };
  if (info?.code !== 'identity.storage_account_personal') {
    // A refusal of the request itself (a bad or replayed state, a refused code, a вимкнене ОСББ): the code is
    // single-use, so reading it again cannot help; the way out is to start over from the card.
    const refused = status !== undefined && status >= 400 && status < 500 && status !== 401 && !isNoAccessProblem(complete.error);
    if (!refused) return <RequestFallback error={complete.error} onRetry={() => mutate()} />;
    return (
      <div className="mx-auto max-w-2xl px-4 py-8">
        <InlineAlert
          title={uk.storage.callbackNotCompleted}
          action={
            <Link className="underline" to="/o/$associationId/association" params={{ associationId }}>
              {uk.storage.callbackToCard}
            </Link>
          }
        />
      </div>
    );
  }
  return (
    <div className="mx-auto max-w-2xl space-y-4 px-4 py-8">
      <InlineAlert title={uk.storage.personalAccount} />
      <StorageLine associationId={associationId} />
    </div>
  );
}

// The consent page was opened without what it needs to finish: say so and point the way back.
export function StorageCallbackLost({ associationId }: { associationId: string | null }) {
  return (
    <div className="mx-auto max-w-2xl px-4 py-8">
      <InlineAlert
        title={uk.storage.callbackLost}
        action={
          associationId ? (
            <Link className="underline" to="/o/$associationId/association" params={{ associationId }}>
              {uk.storage.callbackBack}
            </Link>
          ) : (
            <CardLinks />
          )
        }
      />
    </div>
  );
}

// Without the association id, the way back is the card of each association the person belongs to.
function CardLinks() {
  const memberships = useListMemberships();
  if (memberships.isPending) return null;
  const items = memberships.data?.data.items ?? [];
  if (items.length === 0) {
    return (
      <Link className="underline" to="/">
        {uk.storage.callbackBack}
      </Link>
    );
  }
  return (
    <span className="flex flex-col gap-1">
      <span>{uk.storage.callbackPickCard}</span>
      {items.map((m) => (
        <Link key={m.association_id} className="underline" to="/o/$associationId/association" params={{ associationId: m.association_id }}>
          {m.association_name}
        </Link>
      ))}
    </span>
  );
}
