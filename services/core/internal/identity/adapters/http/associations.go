package http

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
)

// EnableAssociationStates turns the three оператор operations on the association state on.
func (h *Handler) EnableAssociationStates(s *app.AssociationStates) { h.states = s }

const defaultPageSize = 20

func encodeAssociationCursor(c ports.AssociationCursor) string {
	return base64.RawURLEncoding.EncodeToString([]byte(c.CreatedAt.UTC().Format(time.RFC3339Nano) + "|" + c.ID.String()))
}

func decodeAssociationCursor(s string) (*ports.AssociationCursor, bool) {
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, false
	}
	at, id, ok := strings.Cut(string(raw), "|")
	if !ok {
		return nil, false
	}
	t, err := time.Parse(time.RFC3339Nano, at)
	if err != nil {
		return nil, false
	}
	u, err := uuid.Parse(id)
	if err != nil {
		return nil, false
	}
	return &ports.AssociationCursor{CreatedAt: t, ID: u}, true
}

func (h *Handler) ListAssociationsForOperator(ctx context.Context, req gen.ListAssociationsForOperatorRequestObject) (gen.ListAssociationsForOperatorResponseObject, error) {
	if h.states == nil || h.signIn == nil {
		return nil, httpapi.ErrNotImplemented
	}
	limit := defaultPageSize
	if req.Params.Limit != nil {
		limit = int(*req.Params.Limit)
	}
	var after *ports.AssociationCursor
	if req.Params.Cursor != nil {
		c, ok := decodeAssociationCursor(*req.Params.Cursor)
		if !ok {
			detail := "The cursor is not valid."
			return gen.ListAssociationsForOperator422ApplicationProblemPlusJSONResponse{ValidationFailedApplicationProblemPlusJSONResponse: gen.ValidationFailedApplicationProblemPlusJSONResponse{
				Type: "about:blank", Title: "Unprocessable Entity", Status: http.StatusUnprocessableEntity, Detail: &detail,
				Errors: &[]gen.FieldError{{Pointer: "/cursor", Detail: detail}},
			}}, nil
		}
		after = c
	}
	rows, err := h.states.List(ctx, after, limit+1)
	if err != nil {
		return nil, err
	}
	page := gen.OperatorAssociationPage{Items: []gen.OperatorAssociation{}}
	if len(rows) > limit {
		rows = rows[:limit]
		next := encodeAssociationCursor(ports.AssociationCursor{CreatedAt: rows[limit-1].CreatedAt, ID: rows[limit-1].ID})
		page.NextCursor = &next
	}
	for _, r := range rows {
		item := gen.OperatorAssociation{
			Id: r.ID, Name: r.Name, State: gen.AssociationState(r.State),
			ConnectedByHand: r.ConnectedByHand, CreatedAt: r.CreatedAt,
		}
		if r.Edrpou != "" {
			e := r.Edrpou
			item.Edrpou = &e
		}
		page.Items = append(page.Items, item)
	}
	return gen.ListAssociationsForOperator200JSONResponse(page), nil
}

func stateChangeBody(c ports.StateChange) gen.AssociationStateChange {
	b := gen.AssociationStateChange{AssociationId: c.AssociationID, State: gen.AssociationState(c.State), ChangedAt: c.ChangedAt}
	if c.Reason != "" {
		r := c.Reason
		b.Reason = &r
	}
	return b
}

func stateProblem(err error) (gen.Problem, bool) {
	var ce *domain.CodedError
	if !errors.As(err, &ce) {
		return gen.Problem{}, false
	}
	status := http.StatusConflict
	if ce.Code == domain.CodeReasonRequired {
		status = http.StatusUnprocessableEntity
	}
	detail := "The association state change was refused."
	return gen.Problem{Type: "about:blank", Title: http.StatusText(status), Status: status, Detail: &detail, Code: &ce.Code}, true
}

func associationNotFound() gen.NotFoundApplicationProblemPlusJSONResponse {
	detail := "No such association."
	return gen.NotFoundApplicationProblemPlusJSONResponse{Type: "about:blank", Title: "Not Found", Status: http.StatusNotFound, Detail: &detail}
}

func (h *Handler) operatorEmail(ctx context.Context) (string, error) {
	p, err := h.signIn.Principal(ctx, tokenOf(ctx))
	return p.Email, err
}

func (h *Handler) DisableAssociation(ctx context.Context, req gen.DisableAssociationRequestObject) (gen.DisableAssociationResponseObject, error) {
	if h.states == nil || h.signIn == nil {
		return nil, httpapi.ErrNotImplemented
	}
	email, err := h.operatorEmail(ctx)
	if err != nil {
		return nil, err
	}
	reason := ""
	if req.Body != nil {
		reason = req.Body.Reason
	}
	c, err := h.states.Disable(ctx, openapi_types.UUID(req.AssociationId), email, reason)
	if errors.Is(err, ports.ErrAssociationNotFound) {
		return gen.DisableAssociation404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: associationNotFound()}, nil
	}
	if p, ok := stateProblem(err); ok {
		if p.Status == http.StatusUnprocessableEntity {
			return gen.DisableAssociation422ApplicationProblemPlusJSONResponse(p), nil
		}
		return gen.DisableAssociation409ApplicationProblemPlusJSONResponse(p), nil
	}
	if err != nil {
		return nil, err
	}
	return gen.DisableAssociation200JSONResponse(stateChangeBody(c)), nil
}

func (h *Handler) EnableAssociation(ctx context.Context, req gen.EnableAssociationRequestObject) (gen.EnableAssociationResponseObject, error) {
	if h.states == nil || h.signIn == nil {
		return nil, httpapi.ErrNotImplemented
	}
	email, err := h.operatorEmail(ctx)
	if err != nil {
		return nil, err
	}
	c, err := h.states.Enable(ctx, openapi_types.UUID(req.AssociationId), email)
	if errors.Is(err, ports.ErrAssociationNotFound) {
		return gen.EnableAssociation404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: associationNotFound()}, nil
	}
	if p, ok := stateProblem(err); ok {
		return gen.EnableAssociation409ApplicationProblemPlusJSONResponse(p), nil
	}
	if err != nil {
		return nil, err
	}
	return gen.EnableAssociation200JSONResponse(stateChangeBody(c)), nil
}
