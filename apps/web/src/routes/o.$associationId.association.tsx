import { useQuery } from '@tanstack/react-query';
import { createFileRoute } from '@tanstack/react-router';

import { getAssociationCard } from '@/api/gen/endpoints/identity/identity';
import { AssociationCard } from '@/features/identity/association/association-card';
import { AssociationProblemNotice } from '@/features/identity/association/suspended-notice';
import { PageSkeleton, RequestFallback } from '@/features/identity/shell/request-fallback';
import { StorageLine } from '@/features/identity/storage/storage-line';

export const Route = createFileRoute('/o/$associationId/association')({
  component: AssociationPage,
});

function AssociationPage() {
  const { associationId } = Route.useParams();
  const card = useQuery({ queryKey: ['association-card', associationId], queryFn: () => getAssociationCard(associationId) });
  if (card.error) {
    const error = card.error as { status?: number; info?: { code?: string; detail?: string } };
    const known = (error.status === 403 && error.info?.code === 'identity.association_disabled') || (error.status === 409 && error.info?.code === 'governance.drive_not_connected');
    if (known) return <AssociationProblemNotice error={error} associationId={associationId} />;
    return <RequestFallback error={error} onRetry={() => void card.refetch()} />;
  }
  if (!card.data) return <PageSkeleton />;
  return (
    <>
      <AssociationCard card={card.data.data} />
      <div className="mx-auto max-w-2xl px-4 pb-8">
        <StorageLine associationId={associationId} />
      </div>
    </>
  );
}
