import { createFileRoute, redirect } from '@tanstack/react-router';

export const Route = createFileRoute('/cases/$caseId/protocol')({
  beforeLoad: () => {
    throw redirect({ to: '/' });
  },
});
