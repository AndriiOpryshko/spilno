import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from '@tanstack/react-router';
import { useEffect, useRef } from 'react';

import { completeStorageConnection, getStorageConnection } from '@/api/gen/endpoints/identity/identity';
import { uk } from '@/i18n/uk';

export const storageFlowKey = 'storage-flow-association-id';

const authorizationHref = (associationId: string) => `/api/identity/associations/${associationId}/storage-connection/authorization`;

export function StorageLine({ associationId }: { associationId: string }) {
  const connection = useQuery({ queryKey: ['storage-connection', associationId], queryFn: () => getStorageConnection(associationId) });
  if (!connection.data) return null;
  const { connected, google_account_email: email } = connection.data.data;
  if (connected) return <p>{uk.storage.line(email ?? '')}</p>;
  return (
    <p className="space-x-2">
      <span>{uk.storage.lineNotConnected}</span>
      <a className="underline" href={authorizationHref(associationId)} onClick={() => sessionStorage.setItem(storageFlowKey, associationId)}>
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

  const code409 = (complete.error as { info?: { code?: string } } | null)?.info?.code;
  return (
    <div className="mx-auto max-w-2xl space-y-4 px-4 py-8">
      {code409 === 'identity.storage_account_personal' && <p role="alert">{uk.storage.personalAccount}</p>}
      {complete.isError && <StorageLine associationId={associationId} />}
    </div>
  );
}
