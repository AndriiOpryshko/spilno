package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/ports"
)

func (h *Handler) SearchRegistry(ctx context.Context, req gen.SearchRegistryRequestObject) (gen.SearchRegistryResponseObject, error) {
	p := req.Params
	f := ports.RegistryFilter{
		Counterparty: p.Counterparty, Subject: p.Subject,
		Amount: p.AmountKopiyky, AmountFrom: p.AmountFromKopiyky, AmountTo: p.AmountToKopiyky,
		DateFrom: dayOf(p.DateFrom), DateTo: dayOf(p.DateTo),
	}
	var limit int
	if p.Limit != nil {
		limit = *p.Limit
	}
	var cursor string
	if p.Cursor != nil {
		cursor = *p.Cursor
	}
	page, err := h.registry.Search(ctx, req.AssociationId, f, limit, cursor)
	if errors.Is(err, app.ErrInvalidCursor) {
		return gen.SearchRegistry422ApplicationProblemPlusJSONResponse(gen.Problem{
			Type:   "about:blank",
			Title:  "Validation failed",
			Status: http.StatusUnprocessableEntity,
			Errors: &[]gen.FieldError{{Pointer: "/query/cursor", Detail: "is not a cursor returned by this endpoint"}},
		}), nil
	}
	if fields, ok := errors.AsType[*domain.CodedFieldsError](err); ok {
		return gen.SearchRegistry422ApplicationProblemPlusJSONResponse(rosterInvalidProblem(fields)), nil
	}
	if err != nil {
		return nil, err
	}
	out := gen.SearchRegistry200JSONResponse{
		Items:            make([]gen.RegistryDecision, len(page.Items)),
		WithdrawnNumbers: make([]int32, len(page.WithdrawnNumbers)),
	}
	for i, d := range page.Items {
		out.Items[i] = registryDecisionToAPI(d)
	}
	for i, n := range page.WithdrawnNumbers {
		out.WithdrawnNumbers[i] = int32(n)
	}
	if page.NextCursor != "" {
		out.NextCursor = &page.NextCursor
	}
	return out, nil
}

func (h *Handler) GetSignedProtocol(ctx context.Context, req gen.GetSignedProtocolRequestObject) (gen.GetSignedProtocolResponseObject, error) {
	snap, err := h.registry.SignedProtocol(ctx, req.AssociationId, req.ProtocolId)
	if errors.Is(err, app.ErrSignedProtocolNotFound) {
		return gen.GetSignedProtocol404ApplicationProblemPlusJSONResponse{
			NotFoundApplicationProblemPlusJSONResponse: gen.NotFoundApplicationProblemPlusJSONResponse{
				Type:   "about:blank",
				Title:  "Not Found",
				Status: http.StatusNotFound,
				Detail: ptr("no signed protocol with id " + req.ProtocolId.String()),
			},
		}, nil
	}
	if err != nil {
		return nil, err
	}
	// The payload names its keys as the contract does; the Drive ids it carries have no field in
	// gen.SignedProtocol, so decoding drops them (ADR-0005).
	var out gen.SignedProtocol
	if err := json.Unmarshal(snap.Payload, &out); err != nil {
		return nil, fmt.Errorf("decode snapshot of protocol %s: %w", snap.ProtocolID, err)
	}
	out.ProtocolId = snap.ProtocolID
	out.Number = int32(snap.Number)
	out.MeetingDate = openapi_types.Date{Time: snap.MeetingDate}
	return gen.GetSignedProtocol200JSONResponse(out), nil
}

func dayOf(d *openapi_types.Date) *time.Time {
	if d == nil {
		return nil
	}
	return &d.Time
}

func registryDecisionToAPI(d ports.RegistryDecision) gen.RegistryDecision {
	out := gen.RegistryDecision{
		ProtocolId: d.ProtocolID, AgendaItemId: d.AgendaItemID,
		Number: int32(d.Number), MeetingDate: openapi_types.Date{Time: d.Key.MeetingDate},
		Subject: d.Subject, Counterparty: d.Counterparty, AmountKopiyky: d.AmountKopiyky,
	}
	if d.Currency != nil {
		out.Currency = ptr(gen.Currency(*d.Currency))
	}
	return out
}
