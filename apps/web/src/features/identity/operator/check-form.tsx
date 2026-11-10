import { useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';

import { getGetApplicationForReviewQueryKey, useRecordExtractCheck } from '@/api/gen/endpoints/identity/identity';
import { Button } from '@/components/ui/button';
import { DateField } from '@/components/ui/date-field';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { uk } from '@/i18n/uk';

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

  return (
    <form
      className="flex flex-col gap-3"
      onSubmit={(e) => {
        e.preventDefault();
        setFailed(false);
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
            onError: () => setFailed(true),
          },
        );
      }}
    >
      <h2 className="font-medium">{uk.operator.checkTitle}</h2>
      <div className="flex flex-col gap-1">
        <Label htmlFor="check-signer">{uk.operator.signer}</Label>
        <Input id="check-signer" value={signer} onChange={(e) => setSigner(e.target.value)} />
      </div>
      <div className="flex flex-col gap-1">
        <Label htmlFor="check-signed-on">{uk.operator.signedOn}</Label>
        <DateField id="check-signed-on" value={signedOn} onChange={(e) => setSignedOn(e.target.value)} />
      </div>
      <div className="flex flex-col gap-1">
        <Label htmlFor="check-head">{uk.operator.head}</Label>
        <Input id="check-head" value={head} onChange={(e) => setHead(e.target.value)} />
      </div>
      <div className="flex flex-col gap-1">
        <Label htmlFor="check-edrpou">{uk.operator.edrpou}</Label>
        <Input id="check-edrpou" inputMode="numeric" maxLength={8} value={edrpou} onChange={(e) => setEdrpou(e.target.value)} />
      </div>
      <div className="flex flex-col gap-1">
        <Label htmlFor="check-dated-on">{uk.operator.extractDatedOn}</Label>
        <DateField id="check-dated-on" value={datedOn} onChange={(e) => setDatedOn(e.target.value)} />
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
          {uk.operator.checkFailed}
        </p>
      )}
    </form>
  );
}
