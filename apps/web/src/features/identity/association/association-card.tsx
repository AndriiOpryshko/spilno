import { Link } from '@tanstack/react-router';

import { getDownloadConfirmationExtractUrl } from '@/api/gen/endpoints/identity/identity';
import type { AssociationCard as Card } from '@/api/gen/model';
import { Badge } from '@/components/ui/badge';
import { formatDate, uk } from '@/i18n/uk';

export function AssociationCard({ card }: { card: Card }) {
  const t = uk.association.card;
  return (
    <article className="mx-auto max-w-2xl space-y-6 px-4 py-8">
      <header className="space-y-1">
        <h1 className="text-2xl font-semibold">{card.name}</h1>
        {card.address && <p className="text-sm text-muted-foreground">{card.address}</p>}
        {card.edrpou && <p className="text-sm text-muted-foreground">{card.edrpou}</p>}
      </header>
      <section className="space-y-1">
        <h2 className="text-sm font-medium text-muted-foreground">{t.head}</h2>
        <p>{card.head_full_name ?? t.noHead}</p>
        {card.connected_by_hand && <Badge variant="outline">{t.byHand}</Badge>}
        <div>
          <Link className="underline" to="/applications/new" search={{ kind: 'transfer', association_id: card.id }}>
            {t.fileTransfer}
          </Link>
        </div>
      </section>
      <section className="space-y-3">
        <h2 className="text-sm font-medium text-muted-foreground">{t.history}</h2>
        <ol className="space-y-3">
          {card.history.map((e, i) => (
            <li key={`${e.kind}-${e.occurred_at}-${i}`} className="space-y-1 rounded-md border p-3">
              <p className="font-medium">
                {t.kind[e.kind]} · {t.decidedOn(formatDate(e.occurred_at))}
              </p>
              {e.reason && <p>{e.reason}</p>}
              {e.kind === 'transfer_approved' &&
                e.filed_at &&
                [
                  { what: t.markAppointed, name: e.head_full_name },
                  { what: t.markEnded, name: e.ended_head_full_name },
                ].map(
                  ({ what, name }) =>
                    name && (
                      <p key={what} className="text-sm text-muted-foreground">
                        {t.markOf(what, name, formatDate(e.filed_at as string), formatDate(e.occurred_at))}
                      </p>
                    ),
                )}
              {e.head_full_name && <p>{e.head_full_name}</p>}
              {e.signer_full_name && <p>{t.signer(e.signer_full_name)}</p>}
              {e.extract_dated_on && <p>{t.extractDated(formatDate(e.extract_dated_on))}</p>}
              {e.extract_id && (
                <a className="underline" href={getDownloadConfirmationExtractUrl(card.id, e.extract_id)}>
                  {t.extractDownload}
                </a>
              )}
            </li>
          ))}
        </ol>
      </section>
    </article>
  );
}
