import type { Application } from '@/api/gen/model';
import { uk } from '@/i18n/uk';

import { formatDate } from './format-date';

const t = uk.applications.status;

export function StatusLine({ application }: { application: Application }) {
  if (application.state === 'approved') {
    return (
      <p>
        <span className="font-semibold">{t.approved}</span>{' '}
        {application.association_id && (
          <a className="underline" href={`/o/${application.association_id}/association`}>
            {t.openAssociation(application.association_name)}
          </a>
        )}
      </p>
    );
  }
  if (application.state === 'rejected') {
    return (
      <p>
        <span className="font-semibold">{t.rejected}</span>
        {application.rejection_reason && <> {t.reason(application.rejection_reason)}</>}
      </p>
    );
  }
  const email = import.meta.env.VITE_OPERATOR_CONTACT_EMAIL;
  return (
    <div>
      <p className="font-semibold">{t.pending(formatDate(application.filed_at))}</p>
      {email && <p>{t.contact(email)}</p>}
    </div>
  );
}
