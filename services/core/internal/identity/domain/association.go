package domain

import "strings"

// AssociationState is the state of an association (PRD §10 States).
type AssociationState string

const (
	AssociationConnected AssociationState = "connected"
	AssociationDisabled  AssociationState = "disabled"
)

// Association holds the state machine of one association.
type Association struct {
	State          AssociationState
	DisabledReason string
}

// Disable moves a connected association to disabled with a reason.
func (a *Association) Disable(reason string) error {
	if a.State == AssociationDisabled {
		return &CodedError{Code: CodeAssociationStateUnchanged}
	}
	if strings.TrimSpace(reason) == "" {
		return &CodedError{Code: CodeReasonRequired}
	}
	a.State = AssociationDisabled
	a.DisabledReason = reason
	return nil
}

// Enable moves a disabled association back to connected.
func (a *Association) Enable() error {
	if a.State != AssociationDisabled {
		return &CodedError{Code: CodeAssociationStateUnchanged}
	}
	a.State = AssociationConnected
	a.DisabledReason = ""
	return nil
}
