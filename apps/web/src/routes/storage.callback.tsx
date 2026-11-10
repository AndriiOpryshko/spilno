import { createFileRoute } from '@tanstack/react-router';

import { StorageCallback, storageFlowKey } from '@/features/identity/storage/storage-line';

// The association id is stashed in sessionStorage by whoever starts the consent flow.
export const Route = createFileRoute('/storage/callback')({
  validateSearch: (s: Record<string, unknown>) => ({ code: typeof s.code === 'string' ? s.code : '', state: typeof s.state === 'string' ? s.state : '' }),
  component: CallbackPage,
});

function CallbackPage() {
  const { code, state } = Route.useSearch();
  const associationId = sessionStorage.getItem(storageFlowKey);
  if (!associationId || !code || !state) return null;
  return <StorageCallback associationId={associationId} code={code} state={state} />;
}
