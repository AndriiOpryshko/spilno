import { createFileRoute } from '@tanstack/react-router';

import { RosterScreen } from '@/features/governance/roster/roster-screen';

export const Route = createFileRoute('/o/$associationId/roster')({
  component: RosterPage,
});

function RosterPage() {
  const { associationId } = Route.useParams();
  return <RosterScreen associationId={associationId} />;
}
