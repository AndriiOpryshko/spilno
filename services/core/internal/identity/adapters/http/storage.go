package http

import (
	"context"
	"errors"
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
)

// EnableStorageConnections turns the three storage connection operations on.
func (h *Handler) EnableStorageConnections(s *app.StorageConnections) { h.storage = s }

func (h *Handler) storageStatus(ctx context.Context, c ports.GoogleConnection) (gen.StorageConnection, error) {
	by, err := h.storage.ConnectedBy(ctx, c)
	if err != nil {
		return gen.StorageConnection{}, err
	}
	email := openapi_types.Email(c.AccountEmail)
	at := c.CreatedAt
	return gen.StorageConnection{Connected: true, GoogleAccountEmail: &email, ConnectedByMemberId: by, ConnectedAt: &at}, nil
}

func (h *Handler) GetStorageConnection(ctx context.Context, req gen.GetStorageConnectionRequestObject) (gen.GetStorageConnectionResponseObject, error) {
	if h.storage == nil {
		return nil, httpapi.ErrNotImplemented
	}
	c, err := h.storage.Get(ctx, req.AssociationId)
	if errors.Is(err, ports.ErrGoogleConnectionNotFound) {
		return gen.GetStorageConnection200JSONResponse(gen.StorageConnection{}), nil
	}
	if err != nil {
		return nil, err
	}
	st, err := h.storageStatus(ctx, c)
	if err != nil {
		return nil, err
	}
	return gen.GetStorageConnection200JSONResponse(st), nil
}

func (h *Handler) StartStorageConnection(ctx context.Context, req gen.StartStorageConnectionRequestObject) (gen.StartStorageConnectionResponseObject, error) {
	if h.storage == nil || h.signIn == nil {
		return nil, httpapi.ErrNotImplemented
	}
	p, err := h.signIn.Principal(ctx, tokenOf(ctx))
	if errors.Is(err, app.ErrNoSession) {
		return gen.StartStorageConnection401ApplicationProblemPlusJSONResponse{SessionRequiredApplicationProblemPlusJSONResponse: sessionRequired()}, nil
	}
	if err != nil {
		return nil, err
	}
	link, err := h.storage.Start(p.AccountID, req.AssociationId)
	if err != nil {
		return nil, err
	}
	return gen.StartStorageConnection302Response{Headers: gen.StartStorageConnection302ResponseHeaders{Location: &link}}, nil
}

func (h *Handler) CompleteStorageConnection(ctx context.Context, req gen.CompleteStorageConnectionRequestObject) (gen.CompleteStorageConnectionResponseObject, error) {
	if h.storage == nil || h.signIn == nil {
		return nil, httpapi.ErrNotImplemented
	}
	p, err := h.signIn.Principal(ctx, tokenOf(ctx))
	if errors.Is(err, app.ErrNoSession) {
		return gen.CompleteStorageConnection401ApplicationProblemPlusJSONResponse{SessionRequiredApplicationProblemPlusJSONResponse: sessionRequired()}, nil
	}
	if err != nil {
		return nil, err
	}
	c, err := h.storage.Complete(ctx, p.AccountID, req.AssociationId, req.Body.Code, req.Body.State)
	var ce *domain.CodedError
	switch {
	case errors.Is(err, app.ErrBadStorageState):
		return storageInvalid("/state", "unknown, expired or already used"), nil
	case errors.Is(err, ports.ErrStorageCodeRefused):
		return storageInvalid("/code", "refused by the provider"), nil
	case errors.As(err, &ce):
		detail := "This Google account is a member's sign-in account; connect an association-owned account."
		code := ce.Code
		return gen.CompleteStorageConnection409ApplicationProblemPlusJSONResponse{
			Type: "about:blank", Title: http.StatusText(http.StatusConflict), Status: http.StatusConflict, Detail: &detail, Code: &code,
		}, nil
	case err != nil:
		return nil, err
	}
	st, err := h.storageStatus(ctx, c)
	if err != nil {
		return nil, err
	}
	return gen.CompleteStorageConnection201JSONResponse(st), nil
}

func storageInvalid(pointer, detail string) gen.CompleteStorageConnectionResponseObject {
	errs := []gen.FieldError{{Pointer: pointer, Detail: detail}}
	return gen.CompleteStorageConnection422ApplicationProblemPlusJSONResponse{
		ValidationFailedApplicationProblemPlusJSONResponse: gen.ValidationFailedApplicationProblemPlusJSONResponse{
			Type: "about:blank", Title: "Validation failed", Status: http.StatusUnprocessableEntity, Errors: &errs,
		},
	}
}
