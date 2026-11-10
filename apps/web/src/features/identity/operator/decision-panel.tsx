import { useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';

import {
  getGetApplicationForReviewQueryKey,
  useApproveApplication,
  useGetSession,
  useRejectApplication,
} from '@/api/gen/endpoints/identity/identity';
import type { ApplicationReview } from '@/api/gen/model';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { formatDate, uk } from '@/i18n/uk';

const t = uk.operator;

const messageFor = (err: unknown) => {
  const code = (err as { info?: { code?: string } }).info?.code;
  return (code && t.refusal[code]) || t.decisionFailed;
};

function ErrorText({ children }: { children: string }) {
  return (
    <p role="alert" className="text-sm text-destructive">
      {children}
    </p>
  );
}

export function DecisionPanel({
  applicationId,
  filingId,
  review,
}: {
  applicationId: string;
  filingId: string;
  review: ApplicationReview;
}) {
  const client = useQueryClient();
  const session = useGetSession();
  const approve = useApproveApplication();
  const reject = useRejectApplication();
  const [document, setDocument] = useState('');
  const [reason, setReason] = useState('');
  const [open, setOpen] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [reasonError, setReasonError] = useState<string | null>(null);

  const { application, decisions } = review;
  const decision = decisions[0];
  if (application.state !== 'pending') {
    if (!decision) return null;
    const date = formatDate(decision.decided_at);
    return <p>{application.state === 'approved' ? t.approvedOn(date) : t.rejectedOn(date)}</p>;
  }
  const email = session.data?.data.email;
  if (email && email === review.applicant_email) return <p>{t.ownApplication}</p>;

  const refresh = () => client.invalidateQueries({ queryKey: getGetApplicationForReviewQueryKey(applicationId) });

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-col gap-1">
        <Label htmlFor="decision-document">{t.reliedOnDocument}</Label>
        <Textarea id="decision-document" maxLength={500} value={document} onChange={(e) => setDocument(e.target.value)} />
      </div>
      <div className="flex gap-2">
        <Button
          type="button"
          disabled={approve.isPending}
          onClick={() => {
            setError(null);
            approve.mutate(
              { applicationId, data: { filing_id: filingId, ...(document ? { relied_on_document: document } : {}) } },
              { onSuccess: refresh, onError: (err) => setError(messageFor(err)) },
            );
          }}
        >
          {t.approve}
        </Button>
        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger asChild>
            <Button type="button" variant="outline">
              {t.reject}
            </Button>
          </DialogTrigger>
          <DialogContent aria-describedby={undefined}>
            <DialogTitle>{t.rejectTitle}</DialogTitle>
            <div className="flex flex-col gap-1">
              <Label htmlFor="decision-reason">{t.reasonLabel}</Label>
              <Input id="decision-reason" value={reason} onChange={(e) => setReason(e.target.value)} />
            </div>
            {reasonError && <ErrorText>{reasonError}</ErrorText>}
            <Button
              type="button"
              disabled={reject.isPending}
              onClick={() => {
                setReasonError(null);
                if (!reason.trim()) {
                  setReasonError(t.rejectReasonRequired);
                  return;
                }
                reject.mutate(
                  { applicationId, data: { filing_id: filingId, reason } },
                  {
                    onSuccess: () => {
                      setOpen(false);
                      void refresh();
                    },
                    onError: (err) => setReasonError(messageFor(err)),
                  },
                );
              }}
            >
              {t.reject}
            </Button>
          </DialogContent>
        </Dialog>
      </div>
      {error && <ErrorText>{error}</ErrorText>}
    </div>
  );
}
