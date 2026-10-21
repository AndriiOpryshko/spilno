import { createFileRoute } from '@tanstack/react-router';

import { CasesScreen } from '@/features/governance/cases-screen';

export const Route = createFileRoute('/cases/')({
  component: CasesScreen,
});
