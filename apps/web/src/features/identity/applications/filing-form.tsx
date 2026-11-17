import { Link } from '@tanstack/react-router';
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

const maxExtractBytes = 5 * 1024 * 1024;
const extractExtensions = ['.pdf', '.p7s', '.asice'];

const extractRefused = (file: File) =>
  file.size > maxExtractBytes ||
  !(file.type === 'application/pdf' || extractExtensions.some((ext) => file.name.toLowerCase().endsWith(ext)));

const textFields = [
  ['association_name', t.name],
  ['association_address', t.address],
  ['applicant_full_name', t.fullName],
] as const;

// The way out of a refusal: the заявка of the other kind, or the open one.
function RefusalLink({ code }: { code: string }) {
  switch (code) {
    case 'identity.association_already_connected':
      return <Link className="text-sm underline" to="/applications/new" search={{ kind: 'transfer' }}>{t.linkTransfer}</Link>;
    case 'identity.application_already_open':
      return <Link className="text-sm underline" to="/applications">{t.linkOpen}</Link>;
    case 'identity.association_not_connected':
      return <Link className="text-sm underline" to="/applications/new" search={{ kind: 'connection' }}>{t.linkConnection}</Link>;
    default:
      return null;
  }
}

const refusalMessages: Record<string, string> = {
  'identity.term_ended': uk.applications.termEnded,
  'identity.application_already_open': uk.applications.alreadyOpen,
  'identity.association_already_connected': uk.applications.alreadyConnected,
  'identity.association_not_connected': uk.association.notConnected,
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
  const [summary, setSummary] = useState<string | undefined>();
  const [refusal, setRefusal] = useState<Problem | undefined>();
  const [filed, setFiled] = useState(false);
  const fileMutation = useFileApplication();
  const refileMutation = useRefileApplication();
  const mutation = refile ? refileMutation : fileMutation;

  const set = (name: keyof typeof values) => (e: React.ChangeEvent<HTMLInputElement>) =>
    setValues((v) => ({ ...v, [name]: e.target.value }));

  const refuseExtract = () => {
    setFieldErrors({ extract: t.fieldErrors.extract ?? t.fieldInvalid });
    setSummary(uk.applications.invalid(t.fieldNames.extract ?? ''));
  };

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    setFieldErrors({});
    setRefusal(undefined);
    setSummary(undefined);
    setFiled(false);
    if (extract && extractRefused(extract)) {
      refuseExtract();
      return;
    }
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
        } else if ((err as { status?: number }).status === 413) {
          refuseExtract();
        } else if (problem.errors?.length) {
          const reported = problem.errors.map((x) => x.pointer.replace(/^\//, '')).filter((name) => name in t.fieldNames);
          // The server leaves /extract out when the multipart file is empty next to other errors.
          const broken = [...new Set(extract ? reported : [...reported, 'extract'])];
          if (broken.length) {
            setFieldErrors(Object.fromEntries(broken.map((name) => [name, t.fieldErrors[name] ?? t.fieldInvalid])));
            setSummary(uk.applications.invalid(broken.map((name) => t.fieldNames[name]).join(', ')));
          } else {
            setRefusal(problem);
          }
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

  const refusalMessage = knownRefusalMessage(refusal);

  return (
    <form noValidate onSubmit={submit} className="mx-auto grid max-w-xl gap-4 px-4 py-8">
      <h1 className="text-2xl font-semibold">{kind === 'connection' ? t.connectionTitle : t.transferTitle}</h1>
      {filed && (
        <p role="status" className="font-semibold">
          {kind === 'connection' ? t.filedConnection : t.filedTransfer}
        </p>
      )}
      {summary && <InlineAlert title={summary} />}
      {refusal && (
        <InlineAlert
          title={refusalMessage ?? t.failed}
          action={refusal.code ? <RefusalLink code={refusal.code} /> : undefined}
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
