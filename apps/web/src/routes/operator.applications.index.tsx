import { createFileRoute } from '@tanstack/react-router';

import { OperatorQueue } from '@/features/identity/operator/queue';

export const Route = createFileRoute('/operator/applications/')({
  component: OperatorQueue,
});
