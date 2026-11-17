package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/ports"
)

func (h *Handler) ListBoardMembers(ctx context.Context, req gen.ListBoardMembersRequestObject) (gen.ListBoardMembersResponseObject, error) {
	var on *time.Time
	if req.Params.On != nil {
		on = &req.Params.On.Time
	}
	var limit int
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}
	var cursor string
	if req.Params.Cursor != nil {
		cursor = *req.Params.Cursor
	}
	page, err := h.roster.ListEntries(ctx, req.AssociationId, on, limit, cursor)
	if errors.Is(err, app.ErrInvalidCursor) {
		return gen.ListBoardMembers422ApplicationProblemPlusJSONResponse{
			ValidationFailedApplicationProblemPlusJSONResponse: gen.ValidationFailedApplicationProblemPlusJSONResponse{
				Type:   "about:blank",
				Title:  "Validation failed",
				Status: http.StatusUnprocessableEntity,
				Errors: &[]gen.FieldError{{Pointer: "/query/cursor", Detail: "is not a cursor returned by this endpoint"}},
			},
		}, nil
	}
	if err != nil {
		return nil, err
	}
	out := gen.ListBoardMembers200JSONResponse{Items: make([]gen.BoardMember, len(page.Items))}
	for i, e := range page.Items {
		out.Items[i] = boardMemberToAPI(e)
	}
	if page.NextCursor != "" {
		out.NextCursor = &page.NextCursor
	}
	return out, nil
}

func (h *Handler) CreateBoardMember(ctx context.Context, req gen.CreateBoardMemberRequestObject) (gen.CreateBoardMemberResponseObject, error) {
	e, err := h.roster.AddMember(ctx, req.AssociationId, boardMemberInput(*req.Body))
	if fields, ok := errors.AsType[*domain.CodedFieldsError](err); ok {
		return gen.CreateBoardMember422ApplicationProblemPlusJSONResponse(rosterInvalidProblem(fields)), nil
	}
	if err != nil {
		return nil, err
	}
	return gen.CreateBoardMember201JSONResponse(boardMemberToAPI(e)), nil
}

func (h *Handler) UpdateBoardMember(ctx context.Context, req gen.UpdateBoardMemberRequestObject) (gen.UpdateBoardMemberResponseObject, error) {
	e, err := h.roster.ChangeMember(ctx, req.AssociationId, req.BoardMemberId, boardMemberInput(*req.Body))
	if errors.Is(err, app.ErrRosterEntryNotFound) {
		return gen.UpdateBoardMember404ApplicationProblemPlusJSONResponse{
			NotFoundApplicationProblemPlusJSONResponse: gen.NotFoundApplicationProblemPlusJSONResponse{
				Type:   "about:blank",
				Title:  "Not Found",
				Status: http.StatusNotFound,
				Detail: ptr("no roster entry with id " + req.BoardMemberId.String()),
			},
		}, nil
	}
	if fields, ok := errors.AsType[*domain.CodedFieldsError](err); ok {
		return gen.UpdateBoardMember422ApplicationProblemPlusJSONResponse(rosterInvalidProblem(fields)), nil
	}
	if err != nil {
		return nil, err
	}
	return gen.UpdateBoardMember200JSONResponse(boardMemberToAPI(e)), nil
}

// boardMemberInput maps the request body; position and decision are not in it, so they stay unset.
func boardMemberInput(b gen.BoardMemberInput) domain.BoardMemberEntry {
	e := domain.BoardMemberEntry{FullName: b.FullName, Term: domain.Term{Start: b.TermStart.Time}}
	if b.SignInEmail != nil {
		e.SignInEmail = string(*b.SignInEmail)
	}
	if b.TermEnd != nil {
		e.Term.End = b.TermEnd.Time
	}
	return e
}

func rosterInvalidProblem(e *domain.CodedFieldsError) gen.Problem {
	errs := make([]gen.FieldError, len(e.Fields))
	for i, f := range e.Fields {
		errs[i] = gen.FieldError{Pointer: f.Pointer, Detail: f.Detail}
	}
	return gen.Problem{
		Type:   "about:blank",
		Title:  "Validation failed",
		Status: http.StatusUnprocessableEntity,
		Code:   ptr(e.Code),
		Errors: &errs,
	}
}

func boardMemberToAPI(e ports.RosterEntry) gen.BoardMember {
	out := gen.BoardMember{
		Id:        e.ID,
		FullName:  e.Entry.FullName,
		Position:  gen.BoardPosition(e.Entry.Position),
		TermStart: openapi_types.Date{Time: e.Entry.Term.Start},
		RevisedAt: e.RevisedAt.UTC(),
	}
	if e.Entry.SignInEmail != "" {
		out.SignInEmail = ptr(openapi_types.Email(e.Entry.SignInEmail))
	}
	if !e.Entry.Term.End.IsZero() {
		out.TermEnd = &openapi_types.Date{Time: e.Entry.Term.End}
	}
	if e.Entry.DecisionID != "" {
		id := uuid.MustParse(e.Entry.DecisionID)
		out.DecisionId = &id
	}
	if e.DecisionEffect != "" {
		eff := gen.BoardMemberDecisionEffect(e.DecisionEffect)
		out.DecisionEffect = &eff
	}
	if e.RevisedByMemberID != uuid.Nil {
		out.RevisedByMemberId = &e.RevisedByMemberID
	}
	return out
}
