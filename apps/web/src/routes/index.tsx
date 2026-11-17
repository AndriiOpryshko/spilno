import { createFileRoute, Navigate } from '@tanstack/react-router';

import { useGetSession } from '@/api/gen/endpoints/identity/identity';
import { AssociationPicker } from '@/features/identity/shell/association-picker';
import { resolveEntry } from '@/features/identity/shell/entry';
import { PageSkeleton, RequestFallback } from '@/features/identity/shell/request-fallback';
import { useAllMemberships } from '@/features/identity/shell/use-all-memberships';

export const Route = createFileRoute('/')({
  component: EntryPage,
});

function EntryPage() {
  const session = useGetSession({ query: { retry: false } });
  const memberships = useAllMemberships({ enabled: session.isSuccess });

  if (session.isError) return <RequestFallback error={session.error} onRetry={() => void session.refetch()} />;
  if (memberships.isError) return <RequestFallback error={memberships.error} onRetry={() => void memberships.refetch()} />;
  if (!session.data || !memberships.data) return <PageSkeleton />;

  const principal = session.data.data;
  const items = memberships.data.data.items;
  const entry = resolveEntry({ session: principal, memberships: items });
  switch (entry.kind) {
    case 'sign-in':
      return <Navigate to="/sign-in" replace />;
    case 'applications':
      return <Navigate to="/applications" replace />;
    case 'association':
      return <Navigate to="/o/$associationId/cases" params={{ associationId: entry.associationId }} replace />;
    case 'picker':
      return <AssociationPicker memberships={items} isOperator={principal.is_operator} />;
  }
}
