package domain

import "time"

// VoteChoice is one member's answer on one agenda item in a poll round.
type VoteChoice string

// Poll answers a member may give.
const (
	VoteFor     VoteChoice = "for"
	VoteAgainst VoteChoice = "against"
	VoteAbstain VoteChoice = "abstain"
)

// PollInformalLabel marks every tally as an informal board poll, never a
// meeting vote (AC-41).
const PollInformalLabel = "Попереднє внутрішнє опитування правління — не є голосуванням засідання і не замінює його"

// PollVote is one cast of one member on one item.
type PollVote struct {
	ItemID, MemberID string
	Choice           VoteChoice
	At               time.Time
}

// PollTally is the plain per-item summary of a round: counts and the latest
// vote of each member, with no verdict on quorum or adoption.
type PollTally struct {
	Round                 int
	ItemID                string
	Label                 string
	Counted               bool
	For, Against, Abstain int
	Votes                 []PollVote
}

// PollRound is one попереднє голосування over a frozen protocol revision.
type PollRound struct {
	number   int
	revision int
	openedBy string
	openedAt time.Time
	closed   bool
	closedBy string
	closedAt time.Time
	votes    []PollVote
}

// OpenPollRound starts round number on p and freezes its current revision;
// it fails with poll_not_open unless p is ready for vote.
func OpenPollRound(p *Protocol, number int, openedBy string, at time.Time) (*PollRound, error) {
	if err := p.EnsurePollOpen(); err != nil {
		return nil, err
	}
	return &PollRound{number: number, revision: p.Revision(), openedBy: openedBy, openedAt: at}, nil
}

// Number returns the round number.
func (r *PollRound) Number() int { return r.number }

// Revision returns the protocol revision frozen when the round opened.
func (r *PollRound) Revision() int { return r.revision }

// OpenedBy returns the member who opened the round.
func (r *PollRound) OpenedBy() string { return r.openedBy }

// OpenedAt returns when the round opened.
func (r *PollRound) OpenedAt() time.Time { return r.openedAt }

// Closed reports whether the round is closed.
func (r *PollRound) Closed() bool { return r.closed }

// ClosedBy returns who closed the round, empty while it is open.
func (r *PollRound) ClosedBy() string { return r.closedBy }

// ClosedAt returns when the round closed, zero while it is open.
func (r *PollRound) ClosedAt() time.Time { return r.closedAt }

// Close ends the round; votes stay recorded but are no longer counted
// (AC-05d). Closing again keeps the first close.
func (r *PollRound) Close(actorID string, at time.Time) {
	if r.closed {
		return
	}
	r.closed, r.closedBy, r.closedAt = true, actorID, at
}

// Cast records memberID's choice on itemID. Only the member may cast or
// change their own vote (AC-05e), and only while the round is open and p is
// ready for vote at the round's revision (AC-32).
func (r *PollRound) Cast(p *Protocol, actorID, memberID, itemID string, choice VoteChoice, at time.Time) error {
	if memberID != actorID {
		return &DomainError{Code: CodeVoteOnBehalf}
	}
	if r.closed || p.Revision() != r.revision {
		return &DomainError{Code: CodePollNotOpen}
	}
	if err := p.EnsurePollOpen(); err != nil {
		return err
	}
	r.votes = append(r.votes, PollVote{ItemID: itemID, MemberID: memberID, Choice: choice, At: at})
	return nil
}

// Votes returns every cast in the round, oldest first.
func (r *PollRound) Votes() []PollVote {
	return append([]PollVote(nil), r.votes...)
}

// Tally summarises itemID: the latest vote of each member and, while the
// round is open, the counts per choice.
func (r *PollRound) Tally(itemID string) PollTally {
	t := PollTally{Round: r.number, ItemID: itemID, Label: PollInformalLabel, Counted: !r.closed}
	pos := map[string]int{}
	for _, v := range r.votes {
		if v.ItemID != itemID {
			continue
		}
		if i, ok := pos[v.MemberID]; ok {
			t.Votes[i] = v
			continue
		}
		pos[v.MemberID] = len(t.Votes)
		t.Votes = append(t.Votes, v)
	}
	if !t.Counted {
		return t
	}
	for _, v := range t.Votes {
		switch v.Choice {
		case VoteFor:
			t.For++
		case VoteAgainst:
			t.Against++
		case VoteAbstain:
			t.Abstain++
		}
	}
	return t
}
