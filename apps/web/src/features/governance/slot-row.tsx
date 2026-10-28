import { useMutation, useQueryClient } from '@tanstack/react-query';
import { cn } from 'cn';
import { useId, useState } from 'react';

import { Button } from '@/components/ui/button';
import { Kbd } from '@/components/ui/kbd';
import { Textarea } from '@/components/ui/textarea';
import { splitUah, uk } from '@/i18n/uk';

import { updateSlot } from './api';
import { InlineHint } from './inline-hint';
import { LimitNote } from './limit-note';
import { OriginTag } from './origin-tag';
import { type GlyphState, StatusGlyph } from './status-glyph';
import type { ProtocolSlot, ProtocolWarning } from './types';
import { validateSlot } from './validation';

interface SlotRowProps {
  caseId: string;
  slot: ProtocolSlot;
  warnings: ProtocolWarning[];
  onSaved: () => void;
  // Outside чернетка the text is frozen: no field, no hints, no Save (BR-09, BR-10, BR-13).
  readOnly?: boolean;
}

// One protocol slot (SlotRow in Figma): label column + value column on desktop, stacked on mobile.
export function SlotRow({ caseId, slot, warnings, onSaved, readOnly = false }: SlotRowProps) {
  const queryClient = useQueryClient();
  const fieldId = useId();
  const [draftValue, setValue] = useState(slot.value);
  const value = readOnly ? slot.value : draftValue;
  const editable = slot.editable && !readOnly;
  const dirty = value !== slot.value;
  const hint = editable && (dirty || slot.status !== 'verified') ? validateSlot(slot, value) : null;

  const save = useMutation({
    mutationFn: () => updateSlot(caseId, slot.id, value),
    onSuccess: (draft) => {
      queryClient.setQueryData(['protocol', caseId], draft);
      onSaved();
    },
  });

  const glyph: GlyphState = !value.trim() ? 'empty' : dirty || slot.status === 'needs_review' ? 'attention' : 'done';
  const origin = dirty ? 'editing' : slot.origin;
  const source = slot.sourceRef
    ? `${slot.sourceRef.documentTitle}${slot.editable ? '' : ` · ${uk.protocol.readOnly}`}`
    : slot.status === 'not_found'
      ? uk.protocol.notFoundInDocs
      : uk.protocol.manual;

  const onKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey) && dirty) {
      e.preventDefault();
      save.mutate();
    }
    if (e.key === 'Escape' && dirty) setValue(slot.value);
  };

  return (
    <div className="grid gap-3 border-b px-4 py-4 last:border-b-0 md:grid-cols-[220px_1fr] md:gap-6 md:px-6">
      <div className="space-y-1.5">
        <div className="flex items-center gap-2">
          <StatusGlyph state={glyph} />
          <label htmlFor={fieldId} className="flex-1 text-sm font-semibold">
            {slot.label}
          </label>
          <OriginTag origin={origin} className="md:hidden" />
        </div>
        <p className="text-xs leading-4 text-muted-foreground">{source}</p>
        <OriginTag origin={origin} className="max-md:hidden" />
      </div>

      <div className="min-w-0 space-y-3">
        {slot.amount ? (
          <AmountValue uah={slot.amount.uah} inWords={slot.amount.inWords} />
        ) : editable ? (
          <Textarea
            id={fieldId}
            value={value}
            rows={1}
            placeholder={uk.protocol.placeholder}
            onChange={(e) => setValue(e.target.value)}
            onKeyDown={onKeyDown}
            className={cn(
              'min-h-10 resize-none bg-card text-[15px] leading-[22px] md:text-[15px]',
              !value.trim() && 'border-dashed border-warning',
              'focus-visible:ring-ring/20',
            )}
          />
        ) : (
          <p id={fieldId} className={cn('text-[15px] leading-[22px]', !slot.value.trim() && 'text-muted-foreground')}>
            {slot.value.trim() ? slot.value : uk.protocol.notSpecified}
          </p>
        )}

        {hint && <InlineHint hint={hint} onPick={setValue} />}

        {dirty && (
          <div className="flex flex-wrap items-center gap-2">
            <Button size="sm" onClick={() => save.mutate()} disabled={save.isPending}>
              {save.isPending ? uk.protocol.saving : uk.protocol.save}
            </Button>
            <Button size="sm" variant="ghost" onClick={() => setValue(slot.value)} disabled={save.isPending}>
              {uk.protocol.cancel}
            </Button>
            <span className="hidden items-center gap-1.5 text-xs text-muted-foreground md:flex">
              <Kbd>⌘ Enter</Kbd> {uk.protocol.keySave} <Kbd>Esc</Kbd> {uk.protocol.keyCancel}
            </span>
          </div>
        )}
        {save.isError && <p className="text-[13px] text-destructive">{save.error.message}</p>}

        {warnings.map((w) => (
          <LimitNote key={w.id} warning={w} />
        ))}
      </div>
    </div>
  );
}

function AmountValue({ uah, inWords }: { uah: number; inWords: string }) {
  const { int, frac } = splitUah(uah);
  return (
    <div>
      <p className="flex items-baseline gap-1">
        <span className="text-[28px] font-bold tracking-tight md:text-[32px]">{int}</span>
        <span className="text-base font-medium text-muted-foreground">{frac}</span>
      </p>
      <p className="text-xs text-muted-foreground">{inWords}</p>
    </div>
  );
}
