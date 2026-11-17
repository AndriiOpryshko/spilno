import { Link } from '@tanstack/react-router';

import type { AgendaItem, DecisionSlot, FileRef, MeetingFacts, SignedProtocol, TextFragment } from '@/api/gen/model';
import { getDownloadFileUrl } from '@/api/gen/endpoints/governance/governance';
import { formatDate, formatUah, uk } from '@/i18n/uk';

import { ProtocolStateBadge } from '../protocol-state-badge';

const t = uk.registry.signed;

// The value a slot froze with: the confirmed one, else the newest.
function slotText(slot: DecisionSlot | undefined) {
  const value = slot?.values.find((v) => v.id === slot.confirmation?.slot_value_id) ?? slot?.values[0];
  if (!value) return t.noValue;
  if (value.amount_kopiyky !== null) return formatUah(value.amount_kopiyky / 100);
  if (value.date_value !== null) return value.date_end_value ? `${formatDate(value.date_value)} – ${formatDate(value.date_end_value)}` : formatDate(value.date_value);
  return value.text_value ?? t.noValue;
}

function fragmentText(fragment: TextFragment, item: AgendaItem, protocol: SignedProtocol) {
  if ('text' in fragment) return fragment.text;
  if ('slot' in fragment) return slotText(item.slots.find((s) => s.id === fragment.slot));
  if (fragment.meeting === 'date') return formatDate(protocol.meeting.meeting_date);
  if (fragment.meeting === 'chair') return t.chair;
  return t.present(protocol.meeting_facts.present_member_ids.length);
}

// The signed protocol, read only: the frozen text, the facts of the meeting, the files to download and the
// poll history. A locked protocol has no way to change here; only a new decision refers to it (AC-17).
export function SignedProtocolView({ associationId, protocol }: { associationId: string; protocol: SignedProtocol }) {
  const facts = protocol.meeting_facts;
  return (
    <div className="max-w-3xl space-y-8">
      <header className="space-y-2">
        <p className="text-xs font-semibold tracking-wide text-muted-foreground uppercase">{t.eyebrow}</p>
        <ProtocolStateBadge state="signed" />
        <h1 className="text-[22px] leading-7 font-semibold lg:text-[28px] lg:leading-9">{t.title(protocol.number)}</h1>
        <p className="text-sm text-muted-foreground">{t.meetingDate(formatDate(protocol.meeting_date))}</p>
        <p className="text-sm text-muted-foreground">{uk.protocol.stateIntro.signed}</p>
        <p className="text-sm">
          <Link className="underline underline-offset-2" to="/o/$associationId/cases" params={{ associationId }}>
            {t.newDecision}
          </Link>
        </p>
      </header>

      <section aria-labelledby="signed-agenda" className="space-y-4">
        <h2 id="signed-agenda" className="text-lg font-semibold">
          {t.agenda}
        </h2>
        {protocol.agenda_items.map((item) => (
          <AgendaItemCard key={item.id} item={item} protocol={protocol} />
        ))}
      </section>

      {facts.has_handwritten_corrections && (
        <section aria-labelledby="signed-corrections" className="space-y-2 rounded-lg border bg-card p-4">
          <h2 id="signed-corrections" className="text-lg font-semibold">
            {t.corrections}
          </h2>
          {facts.corrections_note && <p className="text-sm">{facts.corrections_note}</p>}
          <p className="text-[13px] text-muted-foreground">{t.correctionsHint}</p>
        </section>
      )}

      <FilesSection associationId={associationId} protocol={protocol} />
      <PollHistory protocol={protocol} />
    </div>
  );
}

function AgendaItemCard({ item, protocol }: { item: AgendaItem; protocol: SignedProtocol }) {
  const votes = protocol.meeting_facts.item_votes.find((v) => v.agenda_item_id === item.id);
  return (
    <article className="space-y-2 rounded-lg border bg-card p-4">
      <h3 className="text-sm font-semibold">
        {t.agendaItem(item.position)}
        {item.subject && ` · ${item.subject}`}
      </h3>
      <p className="text-sm leading-6">{item.fragments.map((fragment) => fragmentText(fragment, item, protocol)).join('')}</p>
      {votes && (
        <p className="text-[13px] text-muted-foreground">
          {t.meetingVotes}: {uk.protocol.poll.tally(votes.votes_for, votes.votes_against, votes.votes_abstained)}
        </p>
      )}
    </article>
  );
}

function FileLink({ associationId, label, file }: { associationId: string; label: string; file: FileRef }) {
  return (
    <li className="flex flex-wrap items-baseline gap-x-3 gap-y-1">
      <span className="w-40 text-[13px] text-muted-foreground">{label}</span>
      <a href={getDownloadFileUrl(associationId, file.id)} className="text-sm font-medium underline underline-offset-2">
        {t.download(file.original_name)}
      </a>
    </li>
  );
}

function FilesSection({ associationId, protocol }: { associationId: string; protocol: SignedProtocol }) {
  const facts: MeetingFacts = protocol.meeting_facts;
  return (
    <section aria-labelledby="signed-files" className="space-y-3">
      <h2 id="signed-files" className="text-lg font-semibold">
        {t.files}
      </h2>
      <ul className="space-y-2">
        <FileLink associationId={associationId} label={t.scan} file={facts.signed_scan} />
        <FileLink associationId={associationId} label={t.printCopy} file={protocol.print_copy.file} />
        {protocol.documents.map((doc) => (
          <FileLink key={doc.id} associationId={associationId} label={t.cited} file={doc.file} />
        ))}
      </ul>
    </section>
  );
}

function PollHistory({ protocol }: { protocol: SignedProtocol }) {
  if (protocol.poll_history.length === 0) return null;
  const subjectOf = (id: string) => protocol.agenda_items.find((i) => i.id === id)?.subject ?? '';
  return (
    <section aria-labelledby="signed-polls" className="space-y-3">
      <h2 id="signed-polls" className="text-lg font-semibold">
        {t.pollHistory}
      </h2>
      {protocol.poll_history.map((round) => (
        <article key={round.id} className="space-y-1 rounded-lg border bg-card p-4">
          <h3 className="text-sm font-semibold">
            {t.pollRound(round.round_number)} · {round.closed_at ? t.pollClosed(formatDate(round.closed_at)) : t.pollOpen}
          </h3>
          <p className="text-[13px] text-muted-foreground">{round.label}</p>
          <ul className="space-y-0.5 text-[13px]">
            {round.items.map((i) => (
              <li key={i.agenda_item_id}>
                {subjectOf(i.agenda_item_id)} · {uk.protocol.poll.tally(i.tally.for, i.tally.against, i.tally.abstain)}
              </li>
            ))}
          </ul>
        </article>
      ))}
    </section>
  );
}
