import { createFileRoute, Link } from '@tanstack/react-router';

import { useListMyApplications } from '@/api/gen/endpoints/identity/identity';
import { StatusLine } from '@/features/identity/applications/status-line';
import { PageSkeleton, RequestFallback } from '@/features/identity/shell/request-fallback';
import { uk } from '@/i18n/uk';

function ApplicationsList() {
  const query = useListMyApplications();
  if (query.isError) return <RequestFallback error={query.error} onRetry={() => void query.refetch()} />;
  if (!query.data) return <PageSkeleton />;
  const items = query.data.data.items;
  return (
    <div className="mx-auto grid max-w-xl gap-4 px-4 py-8">
      <h1 className="text-2xl font-semibold">{uk.applications.title}</h1>
      {items.length === 0 && <p>{uk.applications.detail.empty}</p>}
      <ul className="grid gap-4">
        {items.map((a) => (
          <li key={a.id} className="grid gap-1">
            <Link className="font-medium underline" to="/applications/$applicationId" params={{ applicationId: a.id }}>
              {a.association_name}
            </Link>
            <StatusLine application={a} />
          </li>
        ))}
      </ul>
    </div>
  );
}

export const Route = createFileRoute('/applications/')({
  component: ApplicationsList,
});
