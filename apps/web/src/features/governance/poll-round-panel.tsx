import { Card } from '@/components/ui/card';
import { uk } from '@/i18n/uk';

import { StatusGlyph } from './status-glyph';
import type { PollRound } from './types';

// Current раунд опитування (PollRoundPanel in Figma). Shows state and result only: no vote controls,
// and always labelled as not the meeting's vote (AC-46).
export function PollRoundPanel({ poll }: { poll: PollRound }) {
  const titleId = `poll-round-${poll.round}`;
  return (
    <Card aria-labelledby={titleId} role="region" className="gap-3 rounded-lg p-4 shadow-[0_1px_3px_rgba(22,22,40,0.06)]">
      <div className="flex items-center justify-between text-sm font-semibold">
        <h2 id={titleId}>{uk.protocol.poll.title(poll.round)}</h2>
        <span className="text-muted-foreground">
          {poll.voted} / {poll.boardSize}
        </span>
      </div>
      <p className="flex items-start gap-2 rounded-md bg-warning-subtle px-3 py-2 text-[13px] leading-[18px] font-medium">
        <StatusGlyph state="attention" className="mt-0.5" />
        {uk.protocol.poll.informal}
      </p>
      <p className="text-[13px] text-muted-foreground">{uk.protocol.poll.voted(poll.voted, poll.boardSize)}</p>
      <ol className="space-y-2 text-[13px]">
        {poll.items.map((item, i) => (
          <li key={item.id}>
            <p className="font-medium">
              {i + 1}. {item.title}
            </p>
            <p className="text-xs text-muted-foreground">{uk.protocol.poll.tally(item.for, item.against, item.abstain)}</p>
          </li>
        ))}
      </ol>
    </Card>
  );
}
