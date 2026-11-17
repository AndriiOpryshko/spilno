import { createFileRoute } from '@tanstack/react-router';

import { RegistryScreen } from '@/features/governance/registry/registry-screen';

export const Route = createFileRoute('/o/$associationId/registry/')({
  component: RegistryScreen,
});
