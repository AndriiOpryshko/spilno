package domain

import "time"

// Wording is one version of an agenda item's decision text with its author
// and time (AC-29).
type Wording struct {
	Text     string
	AuthorID string
	At       time.Time
}

// NamedSlot is a decision slot under its key within an agenda item.
type NamedSlot struct {
	Key  string
	Slot *Slot
}

// AgendaItem is one справа on a protocol draft: a subject, the versioned
// wording of its decision and the слоти рішення the wording refers to.
type AgendaItem struct {
	caseID   string
	subject  string
	wordings []Wording
	slots    []NamedSlot
}

// NewAgendaItem returns an agenda item for caseID with subject and no wording.
func NewAgendaItem(caseID, subject string) *AgendaItem {
	return &AgendaItem{caseID: caseID, subject: subject}
}

// CaseID returns the справа the item was opened for.
func (a *AgendaItem) CaseID() string { return a.caseID }

// Subject returns the item's subject.
func (a *AgendaItem) Subject() string { return a.subject }

// Reword records text as the current wording by authorID at at; earlier
// wordings and every slot stay as they are (AC-29).
func (a *AgendaItem) Reword(text, authorID string, at time.Time) {
	a.wordings = append(a.wordings, Wording{Text: text, AuthorID: authorID, At: at})
}

// Wording returns the current wording, false when there is none yet.
func (a *AgendaItem) Wording() (Wording, bool) {
	if len(a.wordings) == 0 {
		return Wording{}, false
	}
	return a.wordings[len(a.wordings)-1], true
}

// Wordings returns every wording, oldest first.
func (a *AgendaItem) Wordings() []Wording {
	return append([]Wording(nil), a.wordings...)
}

// SetSlot puts s under key, replacing a slot already held there.
func (a *AgendaItem) SetSlot(key string, s *Slot) {
	for i := range a.slots {
		if a.slots[i].Key == key {
			a.slots[i].Slot = s
			return
		}
	}
	a.slots = append(a.slots, NamedSlot{Key: key, Slot: s})
}

// Slot returns the slot under key.
func (a *AgendaItem) Slot(key string) (*Slot, bool) {
	for _, ns := range a.slots {
		if ns.Key == key {
			return ns.Slot, true
		}
	}
	return nil, false
}

// Slots returns the item's slots in the order they were added.
func (a *AgendaItem) Slots() []NamedSlot {
	return append([]NamedSlot(nil), a.slots...)
}
