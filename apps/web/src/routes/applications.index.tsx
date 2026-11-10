import { createFileRoute } from '@tanstack/react-router';

import { useListMyApplications } from '@/api/gen/endpoints/identity/identity';
import { StatusLine } from '@/features/identity/applications/status-line';
import { uk } from '@/i18n/uk';

function ApplicationsList() {
  const items = useListMyApplications().data?.data.items ?? [];
  return (
    <div className="mx-auto grid max-w-xl gap-4 px-4 py-8">
      <h1 className="text-2xl font-semibold">{uk.applications.title}</h1>
      {items.length === 0 && <p>{uk.applications.detail.empty}</p>}
      <ul className="grid gap-4">
        {items.map((a) => (
          <li key={a.id} className="grid gap-1">
            <a className="font-medium underline" href={`/applications/${a.id}`}>
              {a.association_name}
            </a>
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
