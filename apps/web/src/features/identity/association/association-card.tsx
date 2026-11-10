import type { AssociationCard as Card } from '@/api/gen/model';
import { Badge } from '@/components/ui/badge';
import { formatDate, uk } from '@/i18n/uk';

import { extractHref } from './extract-href';

export function AssociationCard({ card }: { card: Card }) {
  const t = uk.association.card;
  const newRequest = `/applications/new?kind=transfer&association_id=${card.id}`;
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
        {card.connected_by_hand && (
          <div className="space-y-2">
            <Badge variant="outline">{t.byHand}</Badge>
            <div>
              <a className="underline" href={newRequest}>
                {t.fileTransfer}
              </a>
            </div>
          </div>
        )}
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
                [t.markAppointed, t.markEnded].map((what) => (
                  <p key={what} className="text-sm text-muted-foreground">
                    {t.mark(what, formatDate(e.filed_at as string), formatDate(e.occurred_at))}
                  </p>
                ))}
              {e.head_full_name && <p>{e.head_full_name}</p>}
              {e.signer_full_name && <p>{t.signer(e.signer_full_name)}</p>}
              {e.extract_dated_on && <p>{t.extractDated(formatDate(e.extract_dated_on))}</p>}
              {e.extract_id && (
                <a className="underline" href={extractHref(card.id, e.extract_id)}>
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
