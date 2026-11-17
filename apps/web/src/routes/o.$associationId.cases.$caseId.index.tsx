import { createFileRoute } from '@tanstack/react-router';

import { CaseScreen } from '@/features/governance/cases-screen';

export const Route = createFileRoute('/o/$associationId/cases/$caseId/')({
  component: CasePage,
});

function CasePage() {
  const { caseId } = Route.useParams();
  return <CaseScreen key={caseId} caseId={caseId} />;
}
