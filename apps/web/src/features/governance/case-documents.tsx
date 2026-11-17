import { useQueryClient } from '@tanstack/react-query';
import { useRef, useState } from 'react';

import { getDownloadFileUrl, getGetCaseQueryKey, useAddCaseDocument } from '@/api/gen/endpoints/governance/governance';
import type { CaseDocument } from '@/api/gen/model';
import { Button } from '@/components/ui/button';
import { DateField } from '@/components/ui/date-field';
import { FileField } from '@/components/ui/file-field';
import { Label } from '@/components/ui/label';
import { AssociationProblemNotice } from '@/features/identity/association/suspended-notice';
import { formatDate, uk } from '@/i18n/uk';

import { InlineAlert } from './inline-alert';

const maxBytes = 20 * 1024 * 1024;
const maxNameLength = 255;
const allowedTypes = ['application/pdf', 'image/jpeg', 'image/png'];

type ProblemError = { status?: number; info?: { code?: string; detail?: string } };

// The version number of a document: how many documents lead to it through previous_version_id.
function versionOf(doc: CaseDocument, all: CaseDocument[]) {
  let n = 1;
  let previous = doc.previous_version_id;
  while (previous && n <= all.length) {
    n += 1;
    previous = all.find((d) => d.id === previous)?.previous_version_id ?? null;
  }
  return n;
}

// The documents of a справа with their versions, and the upload of a new one (AC-09, AC-28).
export function CaseDocuments({ associationId, caseId, documents }: { associationId: string; caseId: string; documents: CaseDocument[] }) {
  const queryClient = useQueryClient();
  const formRef = useRef<HTMLFormElement>(null);
  const [file, setFile] = useState<File | undefined>();
  const [refusal, setRefusal] = useState<'choose' | 'invalid' | 'empty' | 'name' | ProblemError | undefined>();
  const [previous, setPrevious] = useState<CaseDocument | undefined>();
  const [documentDate, setDocumentDate] = useState('');
  const upload = useAddCaseDocument();

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    setRefusal(undefined);
    if (!file) return setRefusal('choose');
    // The same rule the server applies (governance.file_invalid); nothing is sent for a file it would refuse.
    // The name is counted in characters (code points), as the server does.
    if (file.size === 0) return setRefusal('empty');
    const nameLength = [...file.name].length;
    if (nameLength === 0 || nameLength > maxNameLength) return setRefusal('name');
    if (file.size > maxBytes || !allowedTypes.includes(file.type)) return setRefusal('invalid');
    upload.mutate(
      { associationId, caseId, data: { file, ...(previous && { previous_version_id: previous.id }), ...(documentDate && { document_date: documentDate }) } },
      {
        onSuccess: () => {
          formRef.current?.reset();
          setFile(undefined);
          setPrevious(undefined);
          setDocumentDate('');
          void queryClient.invalidateQueries({ queryKey: getGetCaseQueryKey(associationId, caseId) });
        },
        onError: (err) => setRefusal((err as ProblemError) ?? {}),
      },
    );
  };

  return (
    <section className="space-y-4" aria-labelledby="case-documents-title">
      <h2 id="case-documents-title" className="text-lg font-semibold">
        {uk.cases.documentsTitle}
      </h2>
      {documents.length === 0 ? (
        <p className="text-sm text-muted-foreground">{uk.cases.noDocuments}</p>
      ) : (
        <ul className="overflow-hidden rounded-lg border bg-card">
          {documents.map((d) => (
            <li key={d.id} className="space-y-1 border-b px-4 py-3 last:border-b-0">
              <a className="text-sm font-medium underline" href={getDownloadFileUrl(associationId, d.file.id)}>
                {d.file.original_name}
              </a>
              <p className="text-xs text-muted-foreground">
                {uk.cases.version(versionOf(d, documents))}
                {d.document_date && ` · ${uk.cases.documentDate(formatDate(d.document_date))}`}
              </p>
              <Button type="button" variant="outline" size="sm" aria-label={uk.cases.newVersionFor(d.file.original_name)} onClick={() => setPrevious(d)}>
                {uk.cases.newVersion}
              </Button>
            </li>
          ))}
        </ul>
      )}

      <form ref={formRef} onSubmit={submit} className="space-y-3">
        <div className="space-y-2">
          <Label htmlFor="case-document-file">{uk.cases.fileLabel}</Label>
          <FileField
            id="case-document-file"
            accept={allowedTypes.join(',')}
            onChange={(e) => {
              setFile(e.target.files?.[0]);
              setRefusal(undefined);
            }}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="case-document-date">{uk.cases.documentDateLabel}</Label>
          <DateField id="case-document-date" value={documentDate} onChange={(e) => setDocumentDate(e.target.value)} />
        </div>
        {previous && (
          <p className="flex flex-wrap items-center gap-3 text-sm">
            {uk.cases.replacing(previous.file.original_name)}
            <Button type="button" variant="ghost" size="sm" onClick={() => setPrevious(undefined)}>
              {uk.cases.cancelNewVersion}
            </Button>
          </p>
        )}
        {refusal && <UploadRefusal refusal={refusal} associationId={associationId} />}
        <Button type="submit" disabled={upload.isPending}>
          {upload.isPending ? uk.cases.adding : uk.cases.addDocument}
        </Button>
      </form>
    </section>
  );
}

function UploadRefusal({ refusal, associationId }: { refusal: 'choose' | 'invalid' | 'empty' | 'name' | ProblemError; associationId: string }) {
  if (refusal === 'choose') return <InlineAlert title={uk.cases.chooseFile} />;
  if (refusal === 'empty') return <InlineAlert title={uk.cases.fileEmpty} />;
  if (refusal === 'name') return <InlineAlert title={uk.cases.fileNameInvalid} />;
  if (refusal === 'invalid' || refusal.info?.code === 'governance.file_invalid') return <InlineAlert title={uk.cases.fileInvalid} />;
  const { status, info } = refusal;
  const known =
    (status === 403 && info?.code === 'identity.association_disabled') || (status === 409 && info?.code === 'governance.drive_not_connected');
  return known ? <AssociationProblemNotice error={refusal} associationId={associationId} /> : <InlineAlert title={uk.cases.uploadFailed} />;
}
