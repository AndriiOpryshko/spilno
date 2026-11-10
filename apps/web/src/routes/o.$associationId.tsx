import { useQuery } from '@tanstack/react-query';
import { createFileRoute, Navigate, Outlet, useRouterState } from '@tanstack/react-router';
import { useEffect, useRef } from 'react';

import { getSession, listMemberships } from '@/api/gen/endpoints/identity/identity';
import { Sidebar } from '@/components/app-shell';
import { listCases } from '@/features/governance/api';
import { isNoAccessProblem } from '@/features/identity/shell/entry';

export const Route = createFileRoute('/o/$associationId')({
  component: AssociationLayout,
});

function AssociationLayout() {
  const { associationId } = Route.useParams();
  const session = useQuery({ queryKey: ['session'], queryFn: () => getSession(), retry: false });
  const memberships = useQuery({ queryKey: ['memberships'], queryFn: () => listMemberships(), enabled: session.isSuccess, retry: false });
  const cases = useQuery({ queryKey: ['cases', associationId], queryFn: listCases, enabled: session.isSuccess });

  // Every navigation inside the association re-reads the roster, so a person removed
  // by a decision made elsewhere is refused on the next action (AC-19).
  const pathname = useRouterState({ select: (state) => state.location.pathname });
  const lastPath = useRef(pathname);
  const refetchMemberships = memberships.refetch;
  useEffect(() => {
    if (lastPath.current === pathname) return;
    lastPath.current = pathname;
    void refetchMemberships();
  }, [pathname, refetchMemberships]);

  if (session.isError) return <Navigate to="/sign-in" replace />;
  if (isNoAccessProblem(memberships.error) || isNoAccessProblem(cases.error)) return <Navigate to="/no-access" replace />;
  if (!session.data || !memberships.data) return null;

  const membership = memberships.data.data.items.find((m) => m.association_id === associationId);
  // Not on this roster: the same page as for a missing association, nothing is read.
  if (!membership) return <Navigate to="/no-access" replace />;

  return (
    <div className="flex min-h-svh">
      <Sidebar associationId={associationId} associationName={membership.association_name} isOperator={session.data.data.is_operator} casesCount={cases.data?.length} />
      <main className="min-w-0 flex-1">
        <Outlet />
      </main>
    </div>
  );
}
