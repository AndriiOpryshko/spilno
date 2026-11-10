import { Link } from '@tanstack/react-router';

import { uk } from '@/i18n/uk';

export function SuspendedNotice({ reason }: { reason: string }) {
  return <p role="alert">{uk.association.disabled(reason)}</p>;
}

type ProblemError = { status?: number; info?: { code?: string; detail?: string } };

export function AssociationProblemNotice({ error, associationId }: { error: ProblemError; associationId: string }) {
  const code = error.info?.code;
  if (error.status === 403 && code === 'identity.association_disabled') return <SuspendedNotice reason={error.info?.detail ?? ''} />;
  if (error.status === 409 && code === 'governance.drive_not_connected') {
    return (
      <div role="alert" className="space-y-1">
        <p>{uk.storage.notConnected}</p>
        <p className="text-sm text-muted-foreground">{uk.storage.anyMemberCanConnect}</p>
        <Link className="underline" to="/o/$associationId/association" params={{ associationId }}>
          {uk.storage.connect}
        </Link>
      </div>
    );
  }
  return null;
}
