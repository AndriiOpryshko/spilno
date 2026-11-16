import { useQuery } from '@tanstack/react-query';
import { createFileRoute, Navigate, Outlet, useRouterState } from '@tanstack/react-router';
import { useEffect, useRef } from 'react';

import { getAssociationCard, getSession, listMemberships } from '@/api/gen/endpoints/identity/identity';
import { Sidebar } from '@/components/app-shell';
import { listCases } from '@/features/governance/api';
import { SuspendedNotice } from '@/features/identity/association/suspended-notice';
import { isNoAccessProblem } from '@/features/identity/shell/entry';
import { PageSkeleton, RequestFallback } from '@/features/identity/shell/request-fallback';

export const Route = createFileRoute('/o/$associationId')({
  component: AssociationLayout,
});

function AssociationLayout() {
  const { associationId } = Route.useParams();
  const session = useQuery({ queryKey: ['session'], queryFn: () => getSession(), retry: false });
  const memberships = useQuery({ queryKey: ['memberships'], queryFn: () => listMemberships(), enabled: session.isSuccess, retry: false });
  const cases = useQuery({ queryKey: ['cases', associationId], queryFn: listCases, enabled: session.isSuccess });

  // A вимкнене ОСББ refuses its scoped reads with the reason in the problem detail (ADR-0007);
  // the membership carries the state only, so the reason comes from the card read.
  const disabled = memberships.data?.data.items.find((m) => m.association_id === associationId)?.association_state === 'disabled';
  const card = useQuery({ queryKey: ['association-card', associationId], queryFn: () => getAssociationCard(associationId), enabled: disabled, retry: false });
  const reason = (card.error as { info?: { detail?: string } } | null)?.info?.detail ?? '';

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

  if (isNoAccessProblem(memberships.error) || isNoAccessProblem(cases.error)) return <Navigate to="/no-access" replace />;
  if (session.isError) return <RequestFallback error={session.error} onRetry={() => void session.refetch()} />;
  if (memberships.isError && !memberships.data) return <RequestFallback error={memberships.error} onRetry={() => void memberships.refetch()} />;
  if (!session.data || !memberships.data) return <PageSkeleton />;

  const membership = memberships.data.data.items.find((m) => m.association_id === associationId);
  // Not on this roster: the same page as for a missing association, nothing is read.
  if (!membership) return <Navigate to="/no-access" replace />;

  return (
    <div className="flex min-h-svh">
      <Sidebar associationId={associationId} associationName={membership.association_name} isOperator={session.data.data.is_operator} casesCount={cases.data?.length} />
      <main className="min-w-0 flex-1">
        {disabled && !card.isPending && !card.isFetching && (
          <div className="px-4 pt-4 lg:px-8">
            <SuspendedNotice reason={reason} />
          </div>
        )}
        <Outlet />
      </main>
    </div>
  );
}
