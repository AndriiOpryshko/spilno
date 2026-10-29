package http

// Stubs for operations merged into the contract ahead of their implementation;
// each later task replaces its stub with a real handler and deletes it here.

import (
	"context"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
)

func (h *Handler) ListBoardMembers(_ context.Context, _ gen.ListBoardMembersRequestObject) (gen.ListBoardMembersResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) CreateBoardMember(_ context.Context, _ gen.CreateBoardMemberRequestObject) (gen.CreateBoardMemberResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) UpdateBoardMember(_ context.Context, _ gen.UpdateBoardMemberRequestObject) (gen.UpdateBoardMemberResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) ListCases(_ context.Context, _ gen.ListCasesRequestObject) (gen.ListCasesResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) CreateCase(_ context.Context, _ gen.CreateCaseRequestObject) (gen.CreateCaseResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) GetCase(_ context.Context, _ gen.GetCaseRequestObject) (gen.GetCaseResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) AddCaseDocument(_ context.Context, _ gen.AddCaseDocumentRequestObject) (gen.AddCaseDocumentResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) DownloadFile(_ context.Context, _ gen.DownloadFileRequestObject) (gen.DownloadFileResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) GetJob(_ context.Context, _ gen.GetJobRequestObject) (gen.GetJobResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) RetryJob(_ context.Context, _ gen.RetryJobRequestObject) (gen.RetryJobResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) AddAgendaItemFromCase(_ context.Context, _ gen.AddAgendaItemFromCaseRequestObject) (gen.AddAgendaItemFromCaseResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) CreateAssistantTurn(_ context.Context, _ gen.CreateAssistantTurnRequestObject) (gen.CreateAssistantTurnResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) ListProtocols(_ context.Context, _ gen.ListProtocolsRequestObject) (gen.ListProtocolsResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) GetProtocol(_ context.Context, _ gen.GetProtocolRequestObject) (gen.GetProtocolResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) AddDecisionSlot(_ context.Context, _ gen.AddDecisionSlotRequestObject) (gen.AddDecisionSlotResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) SaveAgendaItemWording(_ context.Context, _ gen.SaveAgendaItemWordingRequestObject) (gen.SaveAgendaItemWordingResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) ListAssistantTurns(_ context.Context, _ gen.ListAssistantTurnsRequestObject) (gen.ListAssistantTurnsResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) CopyWithdrawnProtocol(_ context.Context, _ gen.CopyWithdrawnProtocolRequestObject) (gen.CopyWithdrawnProtocolResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) StartPollRound(_ context.Context, _ gen.StartPollRoundRequestObject) (gen.StartPollRoundResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) CastPollVote(_ context.Context, _ gen.CastPollVoteRequestObject) (gen.CastPollVoteResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) MarkProtocolReadyForSigning(_ context.Context, _ gen.MarkProtocolReadyForSigningRequestObject) (gen.MarkProtocolReadyForSigningResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) ReturnProtocolToDraft(_ context.Context, _ gen.ReturnProtocolToDraftRequestObject) (gen.ReturnProtocolToDraftResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) LockProtocolWithSignedCopy(_ context.Context, _ gen.LockProtocolWithSignedCopyRequestObject) (gen.LockProtocolWithSignedCopyResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) ConfirmDecisionSlot(_ context.Context, _ gen.ConfirmDecisionSlotRequestObject) (gen.ConfirmDecisionSlotResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) CorrectDecisionSlot(_ context.Context, _ gen.CorrectDecisionSlotRequestObject) (gen.CorrectDecisionSlotResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) WithdrawProtocol(_ context.Context, _ gen.WithdrawProtocolRequestObject) (gen.WithdrawProtocolResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) SearchRegistry(_ context.Context, _ gen.SearchRegistryRequestObject) (gen.SearchRegistryResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) GetSignedProtocol(_ context.Context, _ gen.GetSignedProtocolRequestObject) (gen.GetSignedProtocolResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) GetAssociationSettings(_ context.Context, _ gen.GetAssociationSettingsRequestObject) (gen.GetAssociationSettingsResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) SaveAssociationSettings(_ context.Context, _ gen.SaveAssociationSettingsRequestObject) (gen.SaveAssociationSettingsResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}
