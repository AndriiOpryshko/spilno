import { Link } from '@tanstack/react-router';

import { useListApplicationsForReview } from '@/api/gen/endpoints/identity/identity';
import { Badge } from '@/components/ui/badge';
import { uk } from '@/i18n/uk';

export function OperatorQueue() {
  const { data, error } = useListApplicationsForReview();
  if (error && (error as { status?: number }).status === 403) {
    return <p className="mx-auto mt-8 max-w-2xl">{uk.operator.notOperator}</p>;
  }
  const items = data?.data.items ?? [];
  return (
    <section className="mx-auto mt-8 flex max-w-2xl flex-col gap-3">
      <h1 className="text-xl font-semibold">{uk.operator.queueTitle}</h1>
      <ul className="flex flex-col gap-2">
        {items.map((a) => (
          <li key={a.id} className="flex items-center justify-between gap-3 rounded-md border px-3 py-2">
            <Link to="/operator/applications/$applicationId" params={{ applicationId: a.id }} className="font-medium">
              {a.association_name}
            </Link>
            <Badge variant="secondary">{a.kind === 'transfer' ? uk.operator.kindTransfer : uk.operator.kindConnection}</Badge>
          </li>
        ))}
      </ul>
    </section>
  );
}
