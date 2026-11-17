import { Link } from '@tanstack/react-router';

import { getDownloadExtractForReviewUrl, useGetApplicationForReview } from '@/api/gen/endpoints/identity/identity';
import type { ExtractCheck, ReviewWarning } from '@/api/gen/model';
import { Badge } from '@/components/ui/badge';
import { formatDate, uk } from '@/i18n/uk';

import { CheckForm } from './check-form';
import { DecisionPanel } from './decision-panel';
import { EnteredByApplicant, ExtractSource, NotChecked, SignatureSource } from './source-label';

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

function ExtractLink({ applicationId, filingId }: { applicationId: string; filingId: string }) {
  return (
    <a href={getDownloadExtractForReviewUrl(applicationId, filingId)} className="self-start text-sm underline">
      {uk.operator.downloadExtract}
    </a>
  );
}

function CompetingItem({ id }: { id: string }) {
  const { data } = useGetApplicationForReview(id);
  const review = data?.data;
  if (!review) return null;
  const [filing] = [...review.filings].sort((a, b) => b.created_at.localeCompare(a.created_at));
  return (
    <li className="flex flex-col rounded-md border px-3 py-2">
      <Link to="/operator/applications/$applicationId" params={{ applicationId: id }} className="underline">
        {review.application.association_name}
      </Link>
      {filing && <span>{filing.applicant_full_name}</span>}
      <span className="text-xs text-muted-foreground">{formatDate(filing?.created_at ?? review.application.created_at)}</span>
    </li>
  );
}

// The warning is worded here from its kind and the values the page already holds; the server's detail is English.
function warningText(warning: ReviewWarning, typedName: string, check: ExtractCheck) {
  const { kind } = warning;
  if (kind === 'name_mismatch') {
    return uk.operator.nameMismatch(typedName, check.signer_full_name, check.head_full_name, warning.differing);
  }
  if (kind === 'extract_age') {
    return `${uk.operator.extractAgeBefore} ${formatDate(check.extract_dated_on)} — ${uk.operator.extractAgeAfter}`;
  }
  return undefined;
}

export function ReviewPage({ applicationId }: { applicationId: string }) {
  const { data } = useGetApplicationForReview(applicationId);
  const review = data?.data;
  if (!review) return null;

  const { application, filings, checks, warnings } = review;
  const [filing, ...earlier] = [...filings].sort((a, b) => b.created_at.localeCompare(a.created_at));
  const check = filing
    ? checks.filter((c) => c.filing_id === filing.id).sort((a, b) => b.checked_at.localeCompare(a.checked_at))[0]
    : undefined;
  const transfer = application.kind === 'transfer';
  const markerText = transfer ? uk.operator.competingTransfer : uk.operator.competingConnection;
  const extractSource = check ? <ExtractSource checkedAt={check.checked_at} /> : <NotChecked />;
  const signatureSource = check ? <SignatureSource checkedAt={check.checked_at} /> : <NotChecked />;
  const showMarker = review.competing_application_ids.length > 0;

  return (
    <section className="mx-auto mt-8 flex max-w-2xl flex-col gap-4">
      <h1 className="text-xl font-semibold">{application.association_name}</h1>
      {showMarker && (
        <div className="flex flex-col gap-2">
          <Badge variant="secondary" className="self-start">
            {markerText}
          </Badge>
          <ul className="flex flex-col gap-2">
            {review.competing_application_ids.map((id) => (
              <CompetingItem key={id} id={id} />
            ))}
          </ul>
        </div>
      )}
      {review.storage_account_email && (
        <p className="text-sm">{`${uk.operator.storageAccount}: ${review.storage_account_email}`}</p>
      )}
      {warnings.map((w) => {
        const text = check && warningText(w, filing?.applicant_full_name ?? '', check);
        return (
          text && (
            <p key={w.kind} role="status" className="rounded-md border px-3 py-2 text-sm">
              {text}
            </p>
          )
        );
      })}
      {filing && (
        <>
          <div className="flex flex-col gap-2">
            <dl className="flex flex-col gap-2">
              <Row label={uk.applications.filing.fullName} source={<EnteredByApplicant />}>
                <span>{filing.applicant_full_name}</span>
              </Row>
              <Row label={uk.applications.filing.name} source={<EnteredByApplicant />}>
                {filing.association_name}
              </Row>
              <Row label={uk.applications.filing.edrpou} source={<EnteredByApplicant />}>
                {application.edrpou}
              </Row>
              <Row label={uk.applications.filing.address} source={<EnteredByApplicant />}>
                {filing.association_address}
              </Row>
              <Row label={uk.applications.filing.positionCaption} source={<EnteredByApplicant />}>
                {uk.applications.filing.positionHead}
              </Row>
              <Row label={uk.operator.term} source={<EnteredByApplicant />}>
                {[filing.term_start, filing.term_end].filter(Boolean).map((d) => formatDate(d as string)).join(' — ')}
              </Row>
            </dl>
            <ExtractLink applicationId={applicationId} filingId={filing.id} />
          </div>
          <dl className="flex flex-col gap-2">
            <Row label={uk.operator.signer} source={signatureSource}>
              {check?.signer_full_name}
            </Row>
            <Row label={uk.operator.signedOn} source={signatureSource}>
              {check && formatDate(check.signed_on)}
            </Row>
            <Row label={uk.operator.head} source={extractSource}>
              {check?.head_full_name}
            </Row>
            <Row label={uk.operator.edrpou} source={extractSource}>
              {check?.extract_edrpou}
            </Row>
            <Row label={uk.operator.extractDatedOn} source={extractSource}>
              {check && formatDate(check.extract_dated_on)}
            </Row>
          </dl>
          <CheckForm applicationId={applicationId} filingId={filing.id} />
        </>
      )}
      {filing && <DecisionPanel applicationId={applicationId} filingId={filing.id} review={review} />}
      {earlier.length > 0 && (
        <div className="flex flex-col gap-2">
          <h2 className="text-base font-semibold">{uk.operator.earlierFilings}</h2>
          <ul className="flex flex-col gap-2">
            {earlier.map((f) => (
              <li key={f.id} className="flex flex-col rounded-md border px-3 py-2">
                <span>{f.applicant_full_name}</span>
                <ExtractLink applicationId={applicationId} filingId={f.id} />
              </li>
            ))}
          </ul>
        </div>
      )}
    </section>
  );
}
