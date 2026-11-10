import { useQuery } from '@tanstack/react-query';
import { createFileRoute, Navigate } from '@tanstack/react-router';

import { getSession, listMemberships } from '@/api/gen/endpoints/identity/identity';
import { Skeleton } from '@/components/ui/skeleton';
import { AssociationPicker } from '@/features/identity/shell/association-picker';
import { resolveEntry } from '@/features/identity/shell/entry';

export const Route = createFileRoute('/')({
  component: EntryPage,
});

function EntryPage() {
  const session = useQuery({ queryKey: ['session'], queryFn: () => getSession(), retry: false });
  const memberships = useQuery({ queryKey: ['memberships'], queryFn: () => listMemberships(), enabled: session.isSuccess, retry: false });

  if (session.isError) return <Navigate to="/sign-in" replace />;
  if (!session.data || (!memberships.data && !memberships.isError)) return <Skeleton className="m-8 h-24 max-w-md" />;

  const principal = session.data.data;
  const items = memberships.data?.data.items ?? [];
  const entry = resolveEntry({ session: principal, memberships: items });
  switch (entry.kind) {
    case 'sign-in':
      return <Navigate to="/sign-in" replace />;
    case 'applications':
      window.location.replace('/applications/');
      return null;
    case 'association':
      return <Navigate to="/o/$associationId/cases" params={{ associationId: entry.associationId }} replace />;
    case 'picker':
      return <AssociationPicker memberships={items} isOperator={principal.is_operator} />;
  }
}
