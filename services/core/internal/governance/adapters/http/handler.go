// Package http adapts the governance context's use cases to the generated strict server.
package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
)

type Handler struct {
	meetings *app.MeetingService
}

var _ gen.StrictServerInterface = (*Handler)(nil)

func NewHandler(meetings *app.MeetingService) *Handler {
	return &Handler{meetings: meetings}
}

func (h *Handler) CreateMeeting(ctx context.Context, req gen.CreateMeetingRequestObject) (gen.CreateMeetingResponseObject, error) {
	// The request validator rejects a missing body before this runs.
	b := req.Body
	fields := domain.MeetingFields{
		Title:          b.Title,
		Kind:           domain.MeetingKind("REGULAR"),
		HeldAt:         b.HeldAt,
		EndsAt:         b.EndsAt,
		Location:       b.Location,
		QuorumRequired: true,
	}
	if b.Kind != nil {
		fields.Kind = domain.MeetingKind(*b.Kind)
	}
	if b.QuorumRequired != nil {
		fields.QuorumRequired = *b.QuorumRequired
	}
	x, err := h.meetings.Create(ctx, fields)
	if v, ok := errors.AsType[*domain.ValidationError](err); ok {
		return gen.CreateMeeting422ApplicationProblemPlusJSONResponse{
			UnprocessableEntityApplicationProblemPlusJSONResponse: validationProblem(v),
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return gen.CreateMeeting201JSONResponse(toAPI(x)), nil
}

func (h *Handler) GetMeeting(ctx context.Context, req gen.GetMeetingRequestObject) (gen.GetMeetingResponseObject, error) {
	x, err := h.meetings.Get(ctx, req.Id)
	if errors.Is(err, domain.ErrMeetingNotFound) {
		return gen.GetMeeting404ApplicationProblemPlusJSONResponse{
			NotFoundApplicationProblemPlusJSONResponse: gen.NotFoundApplicationProblemPlusJSONResponse{
				Type:   "about:blank",
				Title:  "Not Found",
				Status: http.StatusNotFound,
				Detail: ptr("no meeting with id " + req.Id.String()),
			},
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return gen.GetMeeting200JSONResponse(toAPI(x)), nil
}

func (h *Handler) ListMeetings(ctx context.Context, req gen.ListMeetingsRequestObject) (gen.ListMeetingsResponseObject, error) {
	var limit int
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}
	var cursor string
	if req.Params.Cursor != nil {
		cursor = *req.Params.Cursor
	}
	page, err := h.meetings.List(ctx, limit, cursor)
	if errors.Is(err, app.ErrInvalidCursor) {
		return gen.ListMeetings422ApplicationProblemPlusJSONResponse{
			UnprocessableEntityApplicationProblemPlusJSONResponse: gen.UnprocessableEntityApplicationProblemPlusJSONResponse{
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
	out := gen.ListMeetings200JSONResponse{Items: make([]gen.Meeting, len(page.Items))}
	for i, x := range page.Items {
		out.Items[i] = toAPI(x)
	}
	if page.NextCursor != "" {
		out.NextCursor = &page.NextCursor
	}
	return out, nil
}

// bodyPointers maps domain fields to their JSON pointer in NewMeeting, the
// same pointer shape the request validator reports.
var bodyPointers = map[string]string{
	domain.FieldTitle:    "/title",
	domain.FieldKind:     "/kind",
	domain.FieldEndsAt:   "/ends_at",
	domain.FieldLocation: "/location",
}

func validationProblem(v *domain.ValidationError) gen.UnprocessableEntityApplicationProblemPlusJSONResponse {
	errs := make([]gen.FieldError, len(v.Fields))
	for i, f := range v.Fields {
		errs[i] = gen.FieldError{Pointer: bodyPointers[f.Field], Detail: f.Detail}
	}
	return gen.UnprocessableEntityApplicationProblemPlusJSONResponse{
		Type:   "about:blank",
		Title:  "Validation failed",
		Status: http.StatusUnprocessableEntity,
		Errors: &errs,
	}
}

func toAPI(x domain.Meeting) gen.Meeting {
	out := gen.Meeting{
		Id:             x.ID,
		Title:          x.Title,
		Kind:           gen.MeetingKind(x.Kind),
		HeldAt:         x.HeldAt.UTC(),
		Location:       x.Location,
		QuorumRequired: x.QuorumRequired,
		CreatedAt:      x.CreatedAt.UTC(),
		UpdatedAt:      x.UpdatedAt.UTC(),
	}
	if x.EndsAt != nil {
		t := x.EndsAt.UTC()
		out.EndsAt = &t
	}
	return out
}

func ptr[T any](v T) *T { return &v }
