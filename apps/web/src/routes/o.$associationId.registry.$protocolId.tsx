import { createFileRoute } from '@tanstack/react-router';

import { SignedProtocolScreen } from '@/features/governance/registry/signed-protocol-screen';

export const Route = createFileRoute('/o/$associationId/registry/$protocolId')({
  component: SignedProtocolRoute,
});

function SignedProtocolRoute() {
  const { protocolId } = Route.useParams();
  return <SignedProtocolScreen protocolId={protocolId} />;
}
