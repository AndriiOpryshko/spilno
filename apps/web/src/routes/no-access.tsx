import { createFileRoute } from '@tanstack/react-router';

import { NoAccessPage } from '@/features/identity/shell/no-access-page';

export const Route = createFileRoute('/no-access')({
  component: NoAccessPage,
});
