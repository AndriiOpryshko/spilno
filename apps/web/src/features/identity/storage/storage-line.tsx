import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link, useNavigate } from '@tanstack/react-router';
import { useEffect, useRef } from 'react';

import { completeStorageConnection, getStartStorageConnectionUrl, getStorageConnection } from '@/api/gen/endpoints/identity/identity';
import { Skeleton } from '@/components/ui/skeleton';
import { InlineAlert } from '@/features/governance/inline-alert';
import { RequestFallback } from '@/features/identity/shell/request-fallback';
import { uk } from '@/i18n/uk';

export const storageFlowKey = 'storage-flow-association-id';

export function StorageLine({ associationId }: { associationId: string }) {
  const connection = useQuery({ queryKey: ['storage-connection', associationId], queryFn: () => getStorageConnection(associationId) });
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
      await queryClient.invalidateQueries({ queryKey: ['storage-connection', associationId] });
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
  if ((complete.error as { info?: { code?: string } } | null)?.info?.code !== 'identity.storage_account_personal') {
    return <RequestFallback error={complete.error} onRetry={() => mutate()} />;
  }
  return (
    <div className="mx-auto max-w-2xl space-y-4 px-4 py-8">
      <p role="alert">{uk.storage.personalAccount}</p>
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
            <Link className="underline" to="/">
              {uk.storage.callbackBack}
            </Link>
          )
        }
      />
    </div>
  );
}
