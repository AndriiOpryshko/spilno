import { useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';

import { getGetMyApplicationQueryKey, useGetMyApplication } from '@/api/gen/endpoints/identity/identity';
import { Button } from '@/components/ui/button';
import { uk } from '@/i18n/uk';

import { FilingForm } from './filing-form';
import { formatDate } from './format-date';
import { StatusLine } from './status-line';

const t = uk.applications.detail;

export function ApplicationDetail({ applicationId }: { applicationId: string }) {
  const queryClient = useQueryClient();
  const query = useGetMyApplication(applicationId);
  const [refiling, setRefiling] = useState(false);

  const detail = query.data?.data;
  if (!detail) return <p className="px-4 py-8">{t.loading}</p>;

  const { application, filings, decisions } = detail;
  const latest = [...filings].sort((a, b) => b.created_at.localeCompare(a.created_at))[0];

  if (refiling && latest) {
    return (
      <FilingForm
        kind={application.kind}
        refile={{
          applicationId,
          filing: latest,
          onFiled: () => {
            setRefiling(false);
            void queryClient.invalidateQueries({ queryKey: getGetMyApplicationQueryKey(applicationId) });
          },
        }}
      />
    );
  }

  const events = [
    ...filings.map((f) => ({ id: f.id, at: f.created_at, text: t.filed(formatDate(f.created_at)) })),
    ...decisions.map((d) => ({
      id: d.id,
      at: d.decided_at,
      text: `${t.decision(formatDate(d.decided_at), d.outcome === 'approved' ? t.outcomeApproved : t.outcomeRejected)}${d.reason ? ` — ${d.reason}` : ''}`,
    })),
  ].sort((a, b) => b.at.localeCompare(a.at));

  return (
    <div className="mx-auto grid max-w-xl gap-4 px-4 py-8">
      <h1 className="text-2xl font-semibold">{application.association_name}</h1>
      <StatusLine application={application} />
      {application.state === 'rejected' && (
        <Button type="button" onClick={() => setRefiling(true)}>
          {t.refile}
        </Button>
      )}
      <ul aria-label={t.history} className="grid gap-2">
        {events.map((e) => (
          <li key={e.id}>{e.text}</li>
        ))}
      </ul>
    </div>
  );
}
