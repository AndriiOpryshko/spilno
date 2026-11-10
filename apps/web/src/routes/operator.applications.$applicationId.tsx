import { createFileRoute } from '@tanstack/react-router';

import { ReviewPage } from '@/features/identity/operator/review-page';

function ReviewRoute() {
  const { applicationId } = Route.useParams();
  return <ReviewPage applicationId={applicationId} />;
}

export const Route = createFileRoute('/operator/applications/$applicationId')({
  component: ReviewRoute,
});
