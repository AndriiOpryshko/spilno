import { createFileRoute } from '@tanstack/react-router';

import { OperatorAssociations } from '@/features/identity/operator/associations';

export const Route = createFileRoute('/operator/associations')({
  component: OperatorAssociations,
});
