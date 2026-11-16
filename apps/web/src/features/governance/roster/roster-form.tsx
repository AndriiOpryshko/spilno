import { useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';

import { getListBoardMembersQueryKey, useCreateBoardMember, useUpdateBoardMember } from '@/api/gen/endpoints/governance/governance';
import type { BoardMember } from '@/api/gen/model';
import { Button } from '@/components/ui/button';
import { DateField } from '@/components/ui/date-field';
import { Dialog, DialogContent, DialogDescription, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { InlineAlert } from '@/features/governance/inline-alert';
import { uk } from '@/i18n/uk';

interface Problem {
  detail?: string;
  errors?: { pointer: string; detail: string }[];
}

const t = uk.association.roster;

// Add (no member) or change (a member of the roster) in one dialog; mounted only while it is open.
export function RosterForm({ associationId, member, onClose }: { associationId: string; member?: BoardMember; onClose: () => void }) {
  const queryClient = useQueryClient();
  const [values, setValues] = useState({
    full_name: member?.full_name ?? '',
    sign_in_email: member?.sign_in_email ?? '',
    term_start: member?.term_start ?? '',
    term_end: member?.term_end ?? '',
  });
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [refusal, setRefusal] = useState<string | undefined>();
  const create = useCreateBoardMember();
  const update = useUpdateBoardMember();
  const pending = create.isPending || update.isPending;

  const set = (name: keyof typeof values) => (e: React.ChangeEvent<HTMLInputElement>) => setValues((v) => ({ ...v, [name]: e.target.value }));

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    setFieldErrors({});
    setRefusal(undefined);
    const data = {
      full_name: values.full_name,
      sign_in_email: values.sign_in_email || null,
      term_start: values.term_start,
      term_end: values.term_end || null,
    };
    const handlers = {
      onSuccess: () => {
        void queryClient.invalidateQueries({ queryKey: getListBoardMembersQueryKey(associationId) });
        onClose();
      },
      onError: (err: unknown) => {
        const problem = ((err as { info?: Problem }).info ?? {}) as Problem;
        if (problem.errors?.length) setFieldErrors(Object.fromEntries(problem.errors.map((x) => [x.pointer.replace(/^\//, ''), x.detail])));
        else setRefusal(problem.detail ?? t.failed);
      },
    };
    if (member) update.mutate({ associationId, boardMemberId: member.id, data }, handlers);
    else create.mutate({ associationId, data }, handlers);
  };

  const field = (name: keyof typeof values, label: string, control: (p: object) => React.ReactNode) => {
    const error = fieldErrors[name];
    return (
      <div className="grid gap-1.5">
        <Label htmlFor={`roster-${name}`}>{label}</Label>
        {control({ id: `roster-${name}`, 'aria-invalid': error ? true : undefined, 'aria-describedby': error ? `roster-${name}-error` : undefined })}
        {error && (
          <p id={`roster-${name}-error`} className="text-[13px] text-destructive">
            {error}
          </p>
        )}
      </div>
    );
  };

  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent>
        <DialogTitle>{member ? t.editTitle : t.addTitle}</DialogTitle>
        <DialogDescription>{t.formHint}</DialogDescription>
        <form noValidate onSubmit={submit} className="grid gap-4">
          {refusal && <InlineAlert title={refusal} />}
          {field('full_name', t.fullName, (p) => <Input {...p} value={values.full_name} onChange={set('full_name')} />)}
          {field('sign_in_email', t.email, (p) => <Input {...p} type="email" value={values.sign_in_email} onChange={set('sign_in_email')} />)}
          {field('term_start', t.termStart, (p) => <DateField {...p} value={values.term_start} onChange={set('term_start')} />)}
          {field('term_end', t.termEnd, (p) => <DateField {...p} value={values.term_end} onChange={set('term_end')} />)}
          <Button type="submit" disabled={pending}>
            {t.save}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  );
}
