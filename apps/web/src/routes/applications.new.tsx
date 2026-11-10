import { createFileRoute } from '@tanstack/react-router';

import { FilingForm } from '@/features/identity/applications/filing-form';

interface Search {
  kind: 'connection' | 'transfer';
  association_id?: string;
}

function FilingRoute() {
  const { kind, association_id } = Route.useSearch();
  return <FilingForm kind={kind} associationId={association_id} />;
}

export const Route = createFileRoute('/applications/new')({
  validateSearch: (s: Record<string, unknown>): Search => ({
    kind: s.kind === 'transfer' ? 'transfer' : 'connection',
    association_id: typeof s.association_id === 'string' ? s.association_id : undefined,
  }),
  component: FilingRoute,
});
