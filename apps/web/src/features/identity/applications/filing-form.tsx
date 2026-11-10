import { Fragment, useState } from 'react';

import { useFileApplication, useRefileApplication } from '@/api/gen/endpoints/identity/identity';
import type { ApplicationFiling } from '@/api/gen/model';
import { Button } from '@/components/ui/button';
import { DateField } from '@/components/ui/date-field';
import { FileField } from '@/components/ui/file-field';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { InlineAlert } from '@/features/governance/inline-alert';
import { uk } from '@/i18n/uk';

type Kind = 'connection' | 'transfer';

interface Problem {
  code?: string;
  detail?: string;
  errors?: { pointer: string; detail: string }[];
}

interface FilingFormProps {
  kind: Kind;
  associationId?: string;
  refile?: { applicationId: string; filing: ApplicationFiling; onFiled: () => void };
}

const t = uk.applications.filing;

const textFields = [
  ['association_name', t.name],
  ['association_address', t.address],
  ['applicant_full_name', t.fullName],
] as const;

const refusalLinks: Record<string, { href: string; label: string }> = {
  'identity.association_already_connected': { href: '/applications/new?kind=transfer', label: t.linkTransfer },
  'identity.application_already_open': { href: '/applications', label: t.linkOpen },
  'identity.association_not_connected': { href: '/applications/new?kind=connection', label: t.linkConnection },
};

const refusalMessages: Record<string, string> = {
  'identity.term_ended': uk.applications.termEnded,
  'identity.application_already_open': uk.applications.alreadyOpen,
  'identity.association_already_connected': uk.applications.alreadyConnected,
};

const knownRefusalMessage = (problem?: Problem) => (problem?.code ? refusalMessages[problem.code] : undefined);

export function FilingForm({ kind, associationId, refile }: FilingFormProps) {
  const [values, setValues] = useState({
    edrpou: '',
    association_name: refile?.filing.association_name ?? '',
    association_address: refile?.filing.association_address ?? '',
    applicant_full_name: refile?.filing.applicant_full_name ?? '',
    term_start: refile?.filing.term_start ?? '',
    term_end: refile?.filing.term_end ?? '',
  });
  const [extract, setExtract] = useState<File | undefined>();
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [refusal, setRefusal] = useState<Problem | undefined>();
  const [filed, setFiled] = useState(false);
  const fileMutation = useFileApplication();
  const refileMutation = useRefileApplication();
  const mutation = refile ? refileMutation : fileMutation;

  const set = (name: keyof typeof values) => (e: React.ChangeEvent<HTMLInputElement>) =>
    setValues((v) => ({ ...v, [name]: e.target.value }));

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    setFieldErrors({});
    setRefusal(undefined);
    setFiled(false);
    const common = {
      association_name: values.association_name,
      association_address: values.association_address,
      applicant_full_name: values.applicant_full_name,
      position: 'head' as const,
      term_start: values.term_start,
      ...(values.term_end ? { term_end: values.term_end } : {}),
      extract: extract ?? new Blob([], { type: 'application/pdf' }),
    };
    const handlers = {
      onSuccess: () => {
        setFiled(true);
        refile?.onFiled();
      },
      onError: (err: unknown) => {
        const problem = ((err as { info?: Problem }).info ?? {}) as Problem;
        if (knownRefusalMessage(problem)) {
          setRefusal(problem);
        } else if (problem.errors?.length) {
          setFieldErrors({
            ...(extract ? {} : { extract: t.extractRequired }),
            ...Object.fromEntries(problem.errors.map((x) => [x.pointer.replace(/^\//, ''), x.detail])),
          });
        } else {
          setRefusal(problem);
        }
      },
    };
    if (refile) {
      refileMutation.mutate({ applicationId: refile.applicationId, data: common }, handlers);
    } else {
      fileMutation.mutate(
        {
          data: {
            kind,
            edrpou: values.edrpou,
            ...(associationId ? { association_id: associationId } : {}),
            ...common,
          },
        },
        handlers,
      );
    }
  };

  const field = (name: string, label: string, control: (p: object) => React.ReactNode) => {
    const error = fieldErrors[name];
    return (
      <div className="grid gap-1.5">
        <Label htmlFor={`filing-${name}`}>{label}</Label>
        {control({
          id: `filing-${name}`,
          'aria-invalid': error ? true : undefined,
          'aria-describedby': error ? `filing-${name}-error` : undefined,
        })}
        {error && (
          <p id={`filing-${name}-error`} className="text-[13px] text-destructive">
            {error}
          </p>
        )}
      </div>
    );
  };

  const link = refusal?.code ? refusalLinks[refusal.code] : undefined;
  const refusalMessage = knownRefusalMessage(refusal);
  const unknownRefusal = refusal && !refusal.detail && !refusalMessage;

  return (
    <form noValidate onSubmit={submit} className="mx-auto grid max-w-xl gap-4 px-4 py-8">
      <h1 className="text-2xl font-semibold">{kind === 'connection' ? t.connectionTitle : t.transferTitle}</h1>
      {filed && (
        <p role="status" className="font-semibold">
          {kind === 'connection' ? t.filedConnection : t.filedTransfer}
        </p>
      )}
      {refusal && (
        <InlineAlert
          title={refusalMessage ?? refusal.detail ?? t.failed}
          action={link && !unknownRefusal ? <a className="text-sm underline" href={link.href}>{link.label}</a> : undefined}
        />
      )}
      {!refile &&
        field('edrpou', t.edrpou, (p) => <Input {...p} inputMode="numeric" value={values.edrpou} onChange={set('edrpou')} />)}
      {textFields.map(([name, label]) => (
        <Fragment key={name}>
          {field(name, label, (p) => <Input {...p} value={values[name]} onChange={set(name)} />)}
        </Fragment>
      ))}
      <div className="grid gap-1.5">
        <span className="text-sm font-medium">{t.positionCaption}</span>
        <span>{t.positionHead}</span>
      </div>
      {field('term_start', t.termStart, (p) => <DateField {...p} value={values.term_start} onChange={set('term_start')} />)}
      {field('term_end', t.termEnd, (p) => <DateField {...p} value={values.term_end} onChange={set('term_end')} />)}
      {field('extract', t.extract, (p) => (
        <FileField {...p} accept=".pdf,.p7s,.asice" onChange={(e) => setExtract(e.target.files?.[0])} />
      ))}
      <Button type="submit" disabled={mutation.isPending}>
        {kind === 'connection' ? t.submitConnection : t.submitTransfer}
      </Button>
    </form>
  );
}
