import { Link } from '@tanstack/react-router';

import { InlineAlert } from '@/features/governance/inline-alert';
import { uk } from '@/i18n/uk';

export function SuspendedNotice({ reason }: { reason: string }) {
  return <InlineAlert title={uk.association.disabled(reason)} />;
}

type ProblemError = { status?: number; info?: { code?: string; detail?: string } };

export function AssociationProblemNotice({ error, associationId }: { error: ProblemError; associationId: string }) {
  const code = error.info?.code;
  if (error.status === 403 && code === 'identity.association_disabled') return <SuspendedNotice reason={error.info?.detail ?? ''} />;
  if (error.status === 409 && code === 'governance.drive_not_connected') {
    return (
      <InlineAlert
        title={uk.storage.notConnected}
        action={
          <Link className="underline" to="/o/$associationId/association" params={{ associationId }}>
            {uk.storage.connect}
          </Link>
        }
      >
        {uk.storage.anyMemberCanConnect}
      </InlineAlert>
    );
  }
  return null;
}
