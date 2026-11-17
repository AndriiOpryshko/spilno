import { getDownloadConfirmationExtractUrl } from '@/api/gen/endpoints/identity/identity';
import type { AssociationEvent } from '@/api/gen/model';
import { formatDate, uk } from '@/i18n/uk';

type Props = { kind: 'appointed' | 'ended'; decisionId: string; history: AssociationEvent[]; associationId: string };

export function RosterDecisionMark({ kind, decisionId, history, associationId }: Props) {
  const event = history.find((e) => e.decision_id === decisionId);
  if (!event || !event.filed_at) return null;
  const t = uk.association.card;
  const what = kind === 'appointed' ? t.markAppointed : t.markEnded;
  return (
    <p className="text-sm text-muted-foreground">
      {t.mark(what, formatDate(event.filed_at), formatDate(event.occurred_at))}
      {event.extract_id && (
        <>
          {' '}
          <a className="underline" href={getDownloadConfirmationExtractUrl(associationId, event.extract_id)}>
            {t.extract}
          </a>
        </>
      )}
    </p>
  );
}
