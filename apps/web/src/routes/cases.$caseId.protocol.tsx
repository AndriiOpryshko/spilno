import { createFileRoute } from '@tanstack/react-router';

import { ProtocolScreen } from '@/features/governance/protocol-screen';

export const Route = createFileRoute('/cases/$caseId/protocol')({
  component: ProtocolPage,
});

function ProtocolPage() {
  const { caseId } = Route.useParams();
  return <ProtocolScreen key={caseId} caseId={caseId} />;
}
