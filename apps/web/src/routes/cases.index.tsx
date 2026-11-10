import { createFileRoute, redirect } from '@tanstack/react-router';

// The old address has no association id: go through the entry, which picks one.
export const Route = createFileRoute('/cases/')({
  beforeLoad: () => {
    throw redirect({ to: '/' });
  },
});
