package domain

import (
	"strconv"
	"strings"
	"time"
)

// CodeSlotUnconfirmed blocks printing while a слот рішення is unconfirmed or
// needs review (AC-21, AC-28b, AC-09); it is listed in ErrorRegistry.
const CodeSlotUnconfirmed = "governance.slot_unconfirmed"

// Blocker is one reason a draft cannot be marked готовий до підпису.
type Blocker struct {
	Code    string
	Pointer string
	Detail  string
}

// AgendaItems returns the draft's agenda items in order.
func (p *Protocol) AgendaItems() []*AgendaItem {
	return append([]*AgendaItem(nil), p.agenda...)
}

// AddAgendaItem adds item to the draft (AC-03c). Like any content change it
// bumps the revision and returns a ready protocol to draft.
func (p *Protocol) AddAgendaItem(expectedRevision int, item *AgendaItem, actorID string, at time.Time) (Effects, error) {
	if err := p.guard(expectedRevision); err != nil {
		return Effects{}, err
	}
	p.agenda = append(p.agenda, item)
	return p.reopen(actorID, at), nil
}

// DocumentRevised sends every slot citing documentID at another version to
// review (AC-09); when any slot changed, the protocol returns to draft like any
// other change. It reports the effects and whether any slot changed.
func (p *Protocol) DocumentRevised(documentID, versionID, actorID string, at time.Time) (Effects, bool) {
	if p.state == ProtocolSigned || p.state == ProtocolWithdrawn {
		return Effects{}, false
	}
	changed := false
	for _, item := range p.agenda {
		for _, ns := range item.slots {
			if ns.Slot.DocumentRevised(documentID, versionID) {
				changed = true
			}
		}
	}
	if !changed {
		return Effects{}, false
	}
	return p.reopen(actorID, at), true
}

// Readiness lists every reason the draft cannot be marked готовий до підпису,
// all at once: no agenda item, an item without a subject (AC-04), and every
// slot that is not confirmed (AC-21, AC-28b, AC-09). Empty means ready.
func (p *Protocol) Readiness() []Blocker {
	var out []Blocker
	if len(p.agenda) == 0 {
		out = append(out, Blocker{Code: CodeDraftIncomplete, Pointer: "/agenda_items", Detail: "no agenda item"})
	}
	for i, item := range p.agenda {
		base := "/agenda_items/" + strconv.Itoa(i)
		if strings.TrimSpace(item.subject) == "" {
			out = append(out, Blocker{Code: CodeDraftIncomplete, Pointer: base + "/subject", Detail: "agenda item without a subject"})
		}
		for _, ns := range item.slots {
			if !ns.Slot.Confirmed() {
				out = append(out, Blocker{
					Code:    CodeSlotUnconfirmed,
					Pointer: base + "/slots/" + ns.Key,
					Detail:  "slot is " + string(ns.Slot.Status()),
				})
			}
		}
	}
	return out
}

// EnsureReadyForSigning returns every blocker of Readiness as one error, nil
// when the draft may be marked готовий до підпису. Its code is
// draft_incomplete when an agenda part is missing, else slot_unconfirmed.
func (p *Protocol) EnsureReadyForSigning() error {
	blockers := p.Readiness()
	if len(blockers) == 0 {
		return nil
	}
	code := CodeSlotUnconfirmed
	fields := make([]FieldIssue, len(blockers))
	for i, b := range blockers {
		if b.Code == CodeDraftIncomplete {
			code = CodeDraftIncomplete
		}
		fields[i] = FieldIssue{Pointer: b.Pointer, Detail: b.Detail}
	}
	return &CodedFieldsError{Code: code, Fields: fields}
}
