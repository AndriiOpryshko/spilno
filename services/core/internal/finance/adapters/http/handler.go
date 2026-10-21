// Package http adapts the finance context's use cases to the generated strict server.
package http

import (
	"context"
	"errors"
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/domain"
)

type Handler struct {
	contracts *app.ContractService
}

var _ gen.StrictServerInterface = (*Handler)(nil)

func NewHandler(contracts *app.ContractService) *Handler {
	return &Handler{contracts: contracts}
}

func (h *Handler) CreateContract(ctx context.Context, req gen.CreateContractRequestObject) (gen.CreateContractResponseObject, error) {
	// The request validator rejects a missing body before this runs.
	b := req.Body
	terms := domain.ContractTerms{
		Counterparty: b.Counterparty,
		AmountCents:  b.AmountCents,
		Currency:     domain.Currency(b.Currency),
		SignedAt:     b.SignedAt.Time,
	}
	if b.EndsAt != nil {
		terms.EndsAt = &b.EndsAt.Time
	}
	c, err := h.contracts.Create(ctx, terms)
	if v, ok := errors.AsType[*domain.ValidationError](err); ok {
		return gen.CreateContract422ApplicationProblemPlusJSONResponse{
			UnprocessableEntityApplicationProblemPlusJSONResponse: validationProblem(v),
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return gen.CreateContract201JSONResponse(toAPI(c)), nil
}

func (h *Handler) GetContract(ctx context.Context, req gen.GetContractRequestObject) (gen.GetContractResponseObject, error) {
	c, err := h.contracts.Get(ctx, req.Id)
	if errors.Is(err, domain.ErrContractNotFound) {
		return gen.GetContract404ApplicationProblemPlusJSONResponse{
			NotFoundApplicationProblemPlusJSONResponse: gen.NotFoundApplicationProblemPlusJSONResponse{
				Type:   "about:blank",
				Title:  "Not Found",
				Status: http.StatusNotFound,
				Detail: ptr("no contract with id " + req.Id.String()),
			},
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return gen.GetContract200JSONResponse(toAPI(c)), nil
}

func (h *Handler) ListContracts(ctx context.Context, req gen.ListContractsRequestObject) (gen.ListContractsResponseObject, error) {
	var limit int
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}
	var cursor string
	if req.Params.Cursor != nil {
		cursor = *req.Params.Cursor
	}
	page, err := h.contracts.List(ctx, limit, cursor)
	if errors.Is(err, app.ErrInvalidCursor) {
		return gen.ListContracts422ApplicationProblemPlusJSONResponse{
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
	out := gen.ListContracts200JSONResponse{Items: make([]gen.Contract, len(page.Items))}
	for i, c := range page.Items {
		out.Items[i] = toAPI(c)
	}
	if page.NextCursor != "" {
		out.NextCursor = &page.NextCursor
	}
	return out, nil
}

// bodyPointers maps domain fields to their JSON pointer in NewContract, the
// same pointer shape the request validator reports.
var bodyPointers = map[string]string{
	domain.FieldCounterparty: "/counterparty",
	domain.FieldAmountCents:  "/amount_cents",
	domain.FieldCurrency:     "/currency",
	domain.FieldEndsAt:       "/ends_at",
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

func toAPI(c domain.Contract) gen.Contract {
	out := gen.Contract{
		Id:           c.ID,
		Counterparty: c.Counterparty,
		AmountCents:  c.AmountCents,
		Currency:     gen.Currency(c.Currency),
		SignedAt:     openapi_types.Date{Time: c.SignedAt},
		CreatedAt:    c.CreatedAt.UTC(),
		UpdatedAt:    c.UpdatedAt.UTC(),
	}
	if c.EndsAt != nil {
		out.EndsAt = &openapi_types.Date{Time: *c.EndsAt}
	}
	return out
}

func ptr[T any](v T) *T { return &v }
