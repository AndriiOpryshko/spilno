import { useQuery } from '@tanstack/react-query';
import { createFileRoute } from '@tanstack/react-router';

import { getAssociationCard } from '@/api/gen/endpoints/identity/identity';
import { AssociationCard } from '@/features/identity/association/association-card';

export const Route = createFileRoute('/o/$associationId/association')({
  component: AssociationPage,
});

function AssociationPage() {
  const { associationId } = Route.useParams();
  const card = useQuery({ queryKey: ['association-card', associationId], queryFn: () => getAssociationCard(associationId) });
  if (!card.data) return null;
  return <AssociationCard card={card.data.data} />;
}
