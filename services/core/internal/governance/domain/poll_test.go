package domain

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

// Expected API (to be implemented in poll.go):
//
//	type VoteChoice string // "for","against","abstain"
//	const PollInformalLabel = "Попереднє внутрішнє опитування правління — не є голосуванням засідання і не замінює його"
//	type PollVote struct{ ItemID, MemberID string; Choice VoteChoice; At time.Time }
//	type PollTally struct {
//		Round                 int
//		ItemID                string
//		Label                 string     // always PollInformalLabel (AC-41)
//		Counted               bool       // false once the round is closed (AC-05d)
//		For, Against, Abstain int        // zero when not counted
//		Votes                 []PollVote // latest vote per member, kept even when not counted
//	}
//	func OpenPollRound(p *Protocol, number int, openedBy string, at time.Time) (*PollRound, error)
//		// poll_not_open unless p is ready_for_vote; freezes p.Revision()
//	func (r *PollRound) Number() int
//	func (r *PollRound) Revision() int                                 // frozen at open
//	func (r *PollRound) Closed() bool
//	func (r *PollRound) Close(actorID string, at time.Time)           // idempotent
//	func (r *PollRound) Cast(p *Protocol, actorID, memberID, itemID string, choice VoteChoice, at time.Time) error
//		// vote_on_behalf if memberID != actorID; poll_not_open unless p is ready_for_vote
//		// at the round's revision and the round is open
//	func (r *PollRound) Votes() []PollVote                            // every cast, oldest first
//	func (r *PollRound) Tally(itemID string) PollTally

const (
	codeVoteOnBehalf = "governance.vote_on_behalf"
	informalLabel    = "Попереднє внутрішнє опитування правління — не є голосуванням засідання і не замінює його"
	itemA            = "item-a"
	itemB            = "item-b"
)

// openRound puts p into ready_for_vote and opens round n on it.
func openRound(t *testing.T, p *Protocol, n int) *PollRound {
	t.Helper()
	if p.State() != stVote {
		if _, err := p.MarkReadyForVote(p.Revision(), actor, t0); err != nil {
			t.Fatalf("ready for vote: %v", err)
		}
	}
	r, err := OpenPollRound(p, n, actor, t0)
	if err != nil {
		t.Fatalf("open round %d: %v", n, err)
	}
	return r
}

func cast(t *testing.T, r *PollRound, p *Protocol, member, item string, c VoteChoice, at time.Time) {
	t.Helper()
	if err := r.Cast(p, member, member, item, c, at); err != nil {
		t.Fatalf("cast %s by %s: %v", c, member, err)
	}
}

func byMember(votes []PollVote) map[string]PollVote {
	out := make(map[string]PollVote, len(votes))
	for _, v := range votes {
		out[v.MemberID] = v
	}
	return out
}

// AC-30: starting the poll freezes the current wording as the round's revision.
func TestPollRound_FreezesTheRevisionAtOpen(t *testing.T) {
	p := protocolIn(t, stDraft)
	if _, err := p.Edit(p.Revision(), actor, t0); err != nil {
		t.Fatal(err)
	}
	r := openRound(t, p, 1)
	if r.Number() != 1 || r.Revision() != 2 || r.Closed() {
		t.Fatalf("round = (%d, rev %d, closed %v), want (1, rev 2, open)", r.Number(), r.Revision(), r.Closed())
	}
	if _, err := p.Edit(p.Revision(), actor, t1); err != nil {
		t.Fatal(err)
	}
	if r.Revision() != 2 {
		t.Fatalf("frozen revision moved to %d after an edit", r.Revision())
	}
}

// AC-32: a round opens only on a draft that is готовий до голосування.
func TestPollRound_OpensOnlyInReadyForVote(t *testing.T) {
	for _, s := range []ProtocolState{stDraft, stSigning, stSigned, stWithdr} {
		t.Run(string(s), func(t *testing.T) {
			_, err := OpenPollRound(protocolIn(t, s), 1, actor, t0)
			if got := domainCode(t, err); got != codePollClosed {
				t.Fatalf("code = %q, want %q", got, codePollClosed)
			}
		})
	}
}

// AC-31, AC-41: tally per item with each member and time, labelled informal.
func TestPollRound_TallyListsMembersTimesAndTheInformalLabel(t *testing.T) {
	p := protocolIn(t, stDraft)
	r := openRound(t, p, 1)
	choices := []VoteChoice{"for", "for", "for", "against", "abstain"}
	for i, c := range choices {
		cast(t, r, p, "member-"+string(rune('1'+i)), itemA, c, t0.Add(time.Duration(i)*time.Minute))
	}
	cast(t, r, p, "member-1", itemB, "against", t1)

	got := r.Tally(itemA)
	if got.Round != 1 || got.ItemID != itemA || !got.Counted {
		t.Fatalf("tally header = (%d, %q, counted %v), want (1, %q, true)", got.Round, got.ItemID, got.Counted, itemA)
	}
	if got.For != 3 || got.Against != 1 || got.Abstain != 1 {
		t.Fatalf("tally = %d/%d/%d, want 3/1/1", got.For, got.Against, got.Abstain)
	}
	if got.Label != informalLabel || PollInformalLabel != informalLabel {
		t.Fatalf("label = %q, want %q", got.Label, informalLabel)
	}
	votes := byMember(got.Votes)
	if len(got.Votes) != 5 || len(votes) != 5 {
		t.Fatalf("votes = %v, want one per member of 5", got.Votes)
	}
	if v := votes["member-4"]; v.Choice != "against" || !v.At.Equal(t0.Add(3*time.Minute)) || v.ItemID != itemA {
		t.Fatalf("member-4 vote = %+v", v)
	}
	if b := r.Tally(itemB); b.Against != 1 || b.For != 0 || len(b.Votes) != 1 {
		t.Fatalf("item B tally = %+v, want only member-1 against", b)
	}
}

// AC-31, AC-41: the tally gives no verdict on quorum, majority or adoption.
func TestPollTally_CarriesNoVerdict(t *testing.T) {
	banned := []string{"quorum", "majority", "adopt", "pass", "verdict", "result", "accepted", "decided"}
	typ := reflect.TypeOf(PollTally{})
	for i := 0; i < typ.NumField(); i++ {
		name := strings.ToLower(typ.Field(i).Name)
		for _, b := range banned {
			if strings.Contains(name, b) {
				t.Errorf("PollTally.%s looks like a verdict", typ.Field(i).Name)
			}
		}
	}
	for i := 0; i < reflect.TypeOf(&PollTally{}).NumMethod(); i++ {
		t.Errorf("PollTally has method %s; the tally is plain data", reflect.TypeOf(&PollTally{}).Method(i).Name)
	}
}

// AC-05e: a member casts and changes only their own vote.
func TestPollRound_RefusesAVoteOnBehalfOfAnotherMember(t *testing.T) {
	p := protocolIn(t, stDraft)
	r := openRound(t, p, 1)
	cast(t, r, p, "member-2", itemA, "for", t0)

	err := r.Cast(p, "member-1", "member-2", itemA, "against", t1)
	if got := domainCode(t, err); got != codeVoteOnBehalf {
		t.Fatalf("code = %q, want %q", got, codeVoteOnBehalf)
	}
	if v := byMember(r.Tally(itemA).Votes)["member-2"]; v.Choice != "for" {
		t.Fatalf("member-2 vote changed to %q by another member", v.Choice)
	}
	if len(r.Votes()) != 1 {
		t.Fatalf("refused vote was recorded: %v", r.Votes())
	}
}

// AC-05e: changing one's vote counts the latest choice and keeps the history.
func TestPollRound_MemberChangesOwnVote(t *testing.T) {
	p := protocolIn(t, stDraft)
	r := openRound(t, p, 1)
	cast(t, r, p, "member-1", itemA, "for", t0)
	cast(t, r, p, "member-1", itemA, "against", t1)

	got := r.Tally(itemA)
	if got.For != 0 || got.Against != 1 || len(got.Votes) != 1 {
		t.Fatalf("tally = %+v, want one against", got)
	}
	if !got.Votes[0].At.Equal(t1) {
		t.Fatalf("latest vote time = %v, want %v", got.Votes[0].At, t1)
	}
	if all := r.Votes(); len(all) != 2 || all[0].Choice != "for" || all[1].Choice != "against" {
		t.Fatalf("history = %v, want for then against", all)
	}
}

// AC-32: voting is refused unless the draft is готовий до голосування.
func TestPollRound_RefusesVotesOutsideReadyForVote(t *testing.T) {
	leave := map[string]func(p *Protocol) error{
		"draft": func(p *Protocol) error { _, err := p.ReturnToDraft(p.Revision(), actor, t1); return err },
		"ready_for_signing": func(p *Protocol) error {
			_, err := p.MarkReadyForSigning(p.Revision(), actor, t1, 14)
			return err
		},
		"signed": func(p *Protocol) error {
			if _, err := p.MarkReadyForSigning(p.Revision(), actor, t1, 14); err != nil {
				return err
			}
			_, err := p.Sign(p.Revision(), actor, t1)
			return err
		},
		"withdrawn": func(p *Protocol) error { _, err := p.Withdraw(p.Revision(), actor, t1, "причина"); return err },
	}
	for name, to := range leave {
		t.Run(name, func(t *testing.T) {
			p := protocolIn(t, stDraft)
			r := openRound(t, p, 1)
			if err := to(p); err != nil {
				t.Fatal(err)
			}
			err := r.Cast(p, "member-1", "member-1", itemA, "for", t1)
			if got := domainCode(t, err); got != codePollClosed {
				t.Fatalf("code = %q, want %q", got, codePollClosed)
			}
			if len(r.Votes()) != 0 {
				t.Fatalf("refused vote was recorded: %v", r.Votes())
			}
		})
	}
}

// AC-05d: a closed round keeps its votes uncounted; the next round starts empty.
func TestPollRound_ClosedRoundKeepsVotesUncounted(t *testing.T) {
	p := protocolIn(t, stDraft)
	r1 := openRound(t, p, 1)
	eff, err := p.ReturnToDraft(p.Revision(), actor, t0)
	if err != nil || !eff.ClosePollRound {
		t.Fatalf("return to draft: eff %+v, err %v", eff, err)
	}
	r1.Close(actor, t0)

	r2 := openRound(t, p, 2)
	cast(t, r2, p, "member-1", itemA, "for", t0)
	cast(t, r2, p, "member-2", itemA, "against", t0)

	eff, err = p.Edit(p.Revision(), actor, t1)
	if err != nil || !eff.ClosePollRound {
		t.Fatalf("edit: eff %+v, err %v", eff, err)
	}
	r2.Close(actor, t1)
	r2.Close("member-9", t1.Add(time.Hour))

	if !r2.Closed() {
		t.Fatal("round 2 is not closed")
	}
	if len(r2.Votes()) != 2 {
		t.Fatalf("round 2 history = %v, want both votes kept", r2.Votes())
	}
	got := r2.Tally(itemA)
	if got.Counted || got.For != 0 || got.Against != 0 || got.Abstain != 0 {
		t.Fatalf("closed tally = %+v, want uncounted with zero counts", got)
	}
	if len(got.Votes) != 2 || got.Label != informalLabel {
		t.Fatalf("closed tally = %+v, want both votes listed with the label", got)
	}

	r3 := openRound(t, p, 3)
	if r3.Revision() != p.Revision() || r3.Revision() == r2.Revision() {
		t.Fatalf("round 3 revision = %d, want the new %d", r3.Revision(), p.Revision())
	}
	if len(r3.Votes()) != 0 || len(r3.Tally(itemA).Votes) != 0 {
		t.Fatalf("round 3 starts with votes: %v", r3.Votes())
	}
	if err := r2.Cast(p, "member-3", "member-3", itemA, "for", t1); domainCode(t, err) != codePollClosed {
		t.Fatalf("vote into closed round 2: %v, want %q", err, codePollClosed)
	}
}
