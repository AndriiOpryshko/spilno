import { useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';

import { getGetApplicationForReviewQueryKey, useRecordExtractCheck } from '@/api/gen/endpoints/identity/identity';
import { Button } from '@/components/ui/button';
import { DateField } from '@/components/ui/date-field';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { uk } from '@/i18n/uk';

const fieldLabels: Record<string, string> = {
  signer_full_name: uk.operator.signer,
  signed_on: uk.operator.signedOn,
  head_full_name: uk.operator.head,
  extract_edrpou: uk.operator.edrpou,
  extract_dated_on: uk.operator.extractDatedOn,
};

export function CheckForm({ applicationId, filingId }: { applicationId: string; filingId: string }) {
  const client = useQueryClient();
  const record = useRecordExtractCheck();
  const [signer, setSigner] = useState('');
  const [signedOn, setSignedOn] = useState('');
  const [head, setHead] = useState('');
  const [edrpou, setEdrpou] = useState('');
  const [datedOn, setDatedOn] = useState('');
  const [valid, setValid] = useState(true);
  const [failed, setFailed] = useState(false);
  const [broken, setBroken] = useState<string[]>([]);
  const invalid = (field: string) => (broken.includes(field) ? true : undefined);
  const describedBy = (field: string) => (broken.includes(field) ? `check-${field}-error` : undefined);
  const message = (field: string) =>
    broken.includes(field) && (
      <p id={`check-${field}-error`} className="text-[13px] text-destructive">
        {uk.operator.checkFieldMessages[field]}
      </p>
    );

  return (
    <form
      className="flex flex-col gap-3"
      onSubmit={(e) => {
        e.preventDefault();
        setFailed(false);
        setBroken([]);
        record.mutate(
          {
            applicationId,
            data: {
              filing_id: filingId,
              signature_valid: valid,
              signer_full_name: signer,
              signed_on: signedOn,
              head_full_name: head,
              extract_edrpou: edrpou,
              extract_dated_on: datedOn,
            },
          },
          {
            onSuccess: () => client.invalidateQueries({ queryKey: getGetApplicationForReviewQueryKey(applicationId) }),
            onError: (err) => {
              const errors = (err as { info?: { errors?: { pointer?: string }[] } }).info?.errors ?? [];
              const fields = errors.map((x) => (x.pointer ?? '').replace(/^\//, '')).filter((f) => f in fieldLabels);
              setBroken(fields);
              setFailed(true);
            },
          },
        );
      }}
    >
      <h2 className="font-medium">{uk.operator.checkTitle}</h2>
      <div className="flex flex-col gap-1">
        <Label htmlFor="check-signer">{uk.operator.signer}</Label>
        <Input id="check-signer" aria-invalid={invalid('signer_full_name')} aria-describedby={describedBy('signer_full_name')} value={signer} onChange={(e) => setSigner(e.target.value)} />
        {message('signer_full_name')}
      </div>
      <div className="flex flex-col gap-1">
        <Label htmlFor="check-signed-on">{uk.operator.signedOn}</Label>
        <DateField id="check-signed-on" aria-invalid={invalid('signed_on')} aria-describedby={describedBy('signed_on')} value={signedOn} onChange={(e) => setSignedOn(e.target.value)} />
        {message('signed_on')}
      </div>
      <div className="flex flex-col gap-1">
        <Label htmlFor="check-head">{uk.operator.head}</Label>
        <Input id="check-head" aria-invalid={invalid('head_full_name')} aria-describedby={describedBy('head_full_name')} value={head} onChange={(e) => setHead(e.target.value)} />
        {message('head_full_name')}
      </div>
      <div className="flex flex-col gap-1">
        <Label htmlFor="check-edrpou">{uk.operator.edrpou}</Label>
        <Input id="check-edrpou" aria-invalid={invalid('extract_edrpou')} aria-describedby={describedBy('extract_edrpou')} inputMode="numeric" maxLength={8} value={edrpou} onChange={(e) => setEdrpou(e.target.value)} />
        {message('extract_edrpou')}
      </div>
      <div className="flex flex-col gap-1">
        <Label htmlFor="check-dated-on">{uk.operator.extractDatedOn}</Label>
        <DateField id="check-dated-on" aria-invalid={invalid('extract_dated_on')} aria-describedby={describedBy('extract_dated_on')} value={datedOn} onChange={(e) => setDatedOn(e.target.value)} />
        {message('extract_dated_on')}
      </div>
      <div className="flex items-center gap-2">
        <input id="check-valid" type="checkbox" checked={valid} onChange={(e) => setValid(e.target.checked)} />
        <Label htmlFor="check-valid">{uk.operator.signatureValid}</Label>
      </div>
      <Button type="submit" disabled={record.isPending}>
        {uk.operator.recordCheck}
      </Button>
      {failed && (
        <p role="alert" className="text-sm text-destructive">
          {broken.length ? uk.operator.checkFieldsBroken(broken.map((f) => fieldLabels[f]).join(', ')) : uk.operator.checkFailed}
        </p>
      )}
    </form>
  );
}
