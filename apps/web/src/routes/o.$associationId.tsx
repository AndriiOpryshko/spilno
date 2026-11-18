import { createFileRoute, Navigate, Outlet, useRouterState } from '@tanstack/react-router';
import { useEffect, useRef } from 'react';

import { useListCases } from '@/api/gen/endpoints/governance/governance';
import { useGetAssociationCard, useGetSession } from '@/api/gen/endpoints/identity/identity';
import { AssociationShell } from '@/components/app-shell';
import { SuspendedNotice } from '@/features/identity/association/suspended-notice';
import { isNoAccessProblem } from '@/features/identity/shell/entry';
import { PageSkeleton, RequestFallback } from '@/features/identity/shell/request-fallback';
import { useAllMemberships } from '@/features/identity/shell/use-all-memberships';

export const Route = createFileRoute('/o/$associationId')({
  component: AssociationLayout,
});

function AssociationLayout() {
  const { associationId } = Route.useParams();
  const session = useGetSession({ query: { retry: false } });
  const memberships = useAllMemberships({ enabled: session.isSuccess });
  const cases = useListCases(associationId, undefined, { query: { enabled: session.isSuccess, retry: false } });

  // A вимкнене ОСББ refuses its scoped reads with the reason in the problem detail (ADR-0007);
  // the membership carries the state only, so the reason comes from the card read.
  const disabled = memberships.data?.data.items.find((m) => m.association_id === associationId)?.association_state === 'disabled';
  const card = useGetAssociationCard(associationId, { query: { enabled: disabled, retry: false } });
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

  // The справи count is the first page's length: shown only when that is the whole list.
  const casesCount = cases.data && cases.data.data.next_cursor === null ? cases.data.data.items.length : undefined;
  const notice = disabled && !card.isPending && !card.isFetching ? <SuspendedNotice reason={reason} /> : undefined;

  return (
    <AssociationShell associationId={associationId} associationName={membership.association_name} isOperator={session.data.data.is_operator} casesCount={casesCount} notice={notice}>
      <Outlet />
    </AssociationShell>
  );
}
