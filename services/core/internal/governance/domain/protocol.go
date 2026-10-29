package domain

import (
	"strings"
	"time"
)

// ProtocolState is the lifecycle state of a protocol draft.
type ProtocolState string

// Lifecycle states of a protocol.
const (
	ProtocolDraft           ProtocolState = "draft"
	ProtocolReadyForVote    ProtocolState = "ready_for_vote"
	ProtocolReadyForSigning ProtocolState = "ready_for_signing"
	ProtocolSigned          ProtocolState = "signed"
	ProtocolWithdrawn       ProtocolState = "withdrawn"
)

// Effects tells the caller which side effects a successful action requires.
type Effects struct {
	ClosePollRound bool
	VoidPrintCopy  bool
}

// ProtocolTransition records one state change.
type ProtocolTransition struct {
	From, To ProtocolState
	Revision int
	ActorID  string
	At       time.Time
}

// Protocol is the lifecycle aggregate of one protocol draft.
type Protocol struct {
	state       ProtocolState
	revision    int
	number      int
	transitions []ProtocolTransition
}

// NewProtocol starts a protocol as a draft at revision 1 without a number.
func NewProtocol(_ string, _ time.Time) *Protocol {
	return &Protocol{state: ProtocolDraft, revision: 1}
}

// State returns the current state.
func (p *Protocol) State() ProtocolState { return p.state }

// Revision returns the content revision.
func (p *Protocol) Revision() int { return p.revision }

// Number returns the reserved protocol number, 0 when none.
func (p *Protocol) Number() int { return p.number }

// Transitions returns the recorded state changes, oldest first.
func (p *Protocol) Transitions() []ProtocolTransition {
	return append([]ProtocolTransition(nil), p.transitions...)
}

func (p *Protocol) guard(expectedRevision int) error {
	switch p.state {
	case ProtocolSigned:
		return &DomainError{Code: CodeProtocolLocked}
	case ProtocolWithdrawn:
		return &DomainError{Code: CodeTransitionDenied}
	}
	if expectedRevision != p.revision {
		return &DomainError{Code: CodeRevisionConflict}
	}
	return nil
}

func (p *Protocol) effectsOfLeaving() Effects {
	return Effects{
		ClosePollRound: p.state == ProtocolReadyForVote,
		VoidPrintCopy:  p.state == ProtocolReadyForSigning,
	}
}

func (p *Protocol) move(to ProtocolState, actorID string, at time.Time) {
	p.transitions = append(p.transitions, ProtocolTransition{
		From: p.state, To: to, Revision: p.revision, ActorID: actorID, At: at,
	})
	p.state = to
}

// reopen returns the protocol to draft with a new revision.
func (p *Protocol) reopen(actorID string, at time.Time) Effects {
	eff := p.effectsOfLeaving()
	p.revision++
	if p.state != ProtocolDraft {
		p.move(ProtocolDraft, actorID, at)
	}
	return eff
}

// Edit applies a content change; it bumps the revision and returns to draft.
func (p *Protocol) Edit(expectedRevision int, actorID string, at time.Time) (Effects, error) {
	if err := p.guard(expectedRevision); err != nil {
		return Effects{}, err
	}
	return p.reopen(actorID, at), nil
}

// ReturnToDraft sends a ready protocol back to draft.
func (p *Protocol) ReturnToDraft(expectedRevision int, actorID string, at time.Time) (Effects, error) {
	if err := p.guard(expectedRevision); err != nil {
		return Effects{}, err
	}
	if p.state == ProtocolDraft {
		return Effects{}, &DomainError{Code: CodeTransitionDenied}
	}
	return p.reopen(actorID, at), nil
}

// MarkReadyForVote opens the draft for voting.
func (p *Protocol) MarkReadyForVote(expectedRevision int, actorID string, at time.Time) (Effects, error) {
	if err := p.guard(expectedRevision); err != nil {
		return Effects{}, err
	}
	if p.state != ProtocolDraft {
		return Effects{}, &DomainError{Code: CodeTransitionDenied}
	}
	p.move(ProtocolReadyForVote, actorID, at)
	return Effects{}, nil
}

// MarkReadyForSigning prints the protocol; number is taken only on the first print.
func (p *Protocol) MarkReadyForSigning(expectedRevision int, actorID string, at time.Time, number int) (Effects, error) {
	if err := p.guard(expectedRevision); err != nil {
		return Effects{}, err
	}
	if p.state == ProtocolReadyForSigning {
		return Effects{}, &DomainError{Code: CodeTransitionDenied}
	}
	if p.number == 0 && number <= 0 {
		return Effects{}, &DomainError{Code: CodeDraftIncomplete}
	}
	eff := p.effectsOfLeaving()
	if p.number == 0 {
		p.number = number
	}
	p.move(ProtocolReadyForSigning, actorID, at)
	return eff, nil
}

// Sign locks a printed protocol.
func (p *Protocol) Sign(expectedRevision int, actorID string, at time.Time) (Effects, error) {
	if err := p.guard(expectedRevision); err != nil {
		return Effects{}, err
	}
	if p.state != ProtocolReadyForSigning {
		return Effects{}, &DomainError{Code: CodeNotReadyForSigning}
	}
	p.move(ProtocolSigned, actorID, at)
	return Effects{}, nil
}

// Withdraw abandons an unsigned protocol; a reason is required.
func (p *Protocol) Withdraw(expectedRevision int, actorID string, at time.Time, reason string) (Effects, error) {
	if err := p.guard(expectedRevision); err != nil {
		return Effects{}, err
	}
	if strings.TrimSpace(reason) == "" {
		return Effects{}, &DomainError{Code: CodeReasonRequired}
	}
	eff := Effects{ClosePollRound: p.state == ProtocolReadyForVote}
	p.move(ProtocolWithdrawn, actorID, at)
	return eff, nil
}

// EnsurePollOpen fails unless the protocol is open for voting.
func (p *Protocol) EnsurePollOpen() error {
	if p.state != ProtocolReadyForVote {
		return &DomainError{Code: CodePollNotOpen}
	}
	return nil
}
