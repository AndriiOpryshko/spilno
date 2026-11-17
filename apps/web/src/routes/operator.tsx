import { useQuery } from '@tanstack/react-query';
import { createFileRoute, Link, Navigate, Outlet } from '@tanstack/react-router';

import { getSession, useListApplicationsForReview } from '@/api/gen/endpoints/identity/identity';
import { isNoAccessProblem } from '@/features/identity/shell/entry';
import { RequestFallback } from '@/features/identity/shell/request-fallback';
import { uk } from '@/i18n/uk';

export const Route = createFileRoute('/operator')({
  component: OperatorLayout,
});

const linkClass = 'rounded-md px-3 py-2 hover:bg-accent';
const activeProps = { className: 'bg-accent font-medium' };

// The one guard for the оператор's section: it only routes. The server still refuses with
// `identity.not_operator`, and that answer sends even a session marked оператор to /no-access.
function OperatorLayout() {
  const session = useQuery({ queryKey: ['session'], queryFn: () => getSession(), retry: false });
  const isOperator = session.data?.data.is_operator === true;
  // The same read as the queue page, so the page shares the answer.
  const queue = useListApplicationsForReview(undefined, { query: { enabled: isOperator, retry: false } });

  if (session.isError) return <RequestFallback error={session.error} onRetry={() => void session.refetch()} />;
  if (!session.data) return null;
  if (!isOperator || isNoAccessProblem(queue.error)) return <Navigate to="/no-access" replace />;
  if (queue.isPending) return null;

  return (
    <>
      <nav className="mx-auto mt-8 flex max-w-2xl gap-1 text-sm">
        <Link to="/operator/applications" className={linkClass} activeProps={activeProps}>
          {uk.operator.navApplications}
        </Link>
        <Link to="/operator/associations" className={linkClass} activeProps={activeProps}>
          {uk.operator.navAssociations}
        </Link>
      </nav>
      <Outlet />
    </>
  );
}
