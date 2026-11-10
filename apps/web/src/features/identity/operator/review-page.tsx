import { useGetApplicationForReview } from '@/api/gen/endpoints/identity/identity';
import { Badge } from '@/components/ui/badge';
import { uk } from '@/i18n/uk';

import { CheckForm } from './check-form';
import { DecisionPanel } from './decision-panel';
import { ExtractSource, NotChecked, SignatureSource } from './source-label';

function Row({ label, children, source }: { label: string; children: React.ReactNode; source: React.ReactNode }) {
  return (
    <div className="flex flex-col">
      <dt className="text-sm text-muted-foreground">{label}</dt>
      <dd className="flex flex-col">
        <span>{children}</span>
        {source}
      </dd>
    </div>
  );
}

export function ReviewPage({ applicationId }: { applicationId: string }) {
  const { data, error } = useGetApplicationForReview(applicationId);
  if (error && (error as { status?: number }).status === 403) {
    return <p className="mx-auto mt-8 max-w-2xl">{uk.operator.notOperator}</p>;
  }
  const review = data?.data;
  if (!review) return null;

  const { application, filings, checks, warnings } = review;
  const filing = filings[filings.length - 1];
  const check = filing ? checks.find((c) => c.filing_id === filing.id) : undefined;
  const transfer = application.kind === 'transfer';
  const markerText = transfer ? uk.operator.competingTransfer : uk.operator.competingConnection;
  const showMarker = review.competing_application_ids.length > 0;

  return (
    <section className="mx-auto mt-8 flex max-w-2xl flex-col gap-4">
      <h1 className="text-xl font-semibold">{application.association_name}</h1>
      {showMarker && <Badge variant="secondary">{markerText}</Badge>}
      {review.storage_account_email && (
        <p className="text-sm">{`${uk.operator.storageAccount}: ${review.storage_account_email}`}</p>
      )}
      {warnings.map((w) => (
        <p key={w.kind} role="status" className="rounded-md border px-3 py-2 text-sm">
          {w.kind === 'extract_age' ? `${uk.operator.extractAgeBefore} ${w.detail} — ${uk.operator.extractAgeAfter}` : w.detail}
        </p>
      ))}
      {filing && (
        <>
          <div className="flex flex-col gap-2">
            <Badge variant="outline" className="self-start">
              {uk.operator.enteredByApplicant}
            </Badge>
            <dl className="flex flex-col gap-2">
              <Row label={uk.applications.filing.fullName} source={null}>
                <span>{filing.applicant_full_name}</span>
              </Row>
              <Row label={uk.applications.filing.name} source={null}>
                {filing.association_name}
              </Row>
            </dl>
          </div>
          <dl className="flex flex-col gap-2">
            <Row label={uk.operator.signer} source={check ? <SignatureSource checkedAt={check.checked_at} /> : <NotChecked />}>
              {check?.signer_full_name}
            </Row>
            <Row label={uk.operator.head} source={check ? <ExtractSource checkedAt={check.checked_at} /> : <NotChecked />}>
              {check?.head_full_name}
            </Row>
            <Row label={uk.operator.edrpou} source={check ? <ExtractSource checkedAt={check.checked_at} /> : <NotChecked />}>
              {check?.extract_edrpou}
            </Row>
          </dl>
          <CheckForm applicationId={applicationId} filingId={filing.id} />
        </>
      )}
      {filing && <DecisionPanel applicationId={applicationId} filingId={filing.id} review={review} />}
    </section>
  );
}
