import { createFileRoute } from '@tanstack/react-router';

import { ApplicationDetail } from '@/features/identity/applications/application-detail';

function DetailRoute() {
  const { applicationId } = Route.useParams();
  return <ApplicationDetail applicationId={applicationId} />;
}

export const Route = createFileRoute('/applications/$applicationId')({
  component: DetailRoute,
});
