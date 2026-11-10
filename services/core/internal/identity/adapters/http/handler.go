// Package http adapts the identity context to the generated strict server.
package http

import (
	"context"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
)

// Handler stubs every operation until the identity tasks implement them.
type Handler struct{}

var _ gen.StrictServerInterface = (*Handler)(nil)

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) GetSession(_ context.Context, _ gen.GetSessionRequestObject) (gen.GetSessionResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) StartSignIn(_ context.Context, _ gen.StartSignInRequestObject) (gen.StartSignInResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) CompleteSignIn(_ context.Context, _ gen.CompleteSignInRequestObject) (gen.CompleteSignInResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) SignOut(_ context.Context, _ gen.SignOutRequestObject) (gen.SignOutResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) ListMyApplications(_ context.Context, _ gen.ListMyApplicationsRequestObject) (gen.ListMyApplicationsResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) FileApplication(_ context.Context, _ gen.FileApplicationRequestObject) (gen.FileApplicationResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) GetMyApplication(_ context.Context, _ gen.GetMyApplicationRequestObject) (gen.GetMyApplicationResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) RefileApplication(_ context.Context, _ gen.RefileApplicationRequestObject) (gen.RefileApplicationResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) DownloadMyExtract(_ context.Context, _ gen.DownloadMyExtractRequestObject) (gen.DownloadMyExtractResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) GetAssociationCard(_ context.Context, _ gen.GetAssociationCardRequestObject) (gen.GetAssociationCardResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) DownloadConfirmationExtract(_ context.Context, _ gen.DownloadConfirmationExtractRequestObject) (gen.DownloadConfirmationExtractResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) GetStorageConnection(_ context.Context, _ gen.GetStorageConnectionRequestObject) (gen.GetStorageConnectionResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) CompleteStorageConnection(_ context.Context, _ gen.CompleteStorageConnectionRequestObject) (gen.CompleteStorageConnectionResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) StartStorageConnection(_ context.Context, _ gen.StartStorageConnectionRequestObject) (gen.StartStorageConnectionResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) ListMemberships(_ context.Context, _ gen.ListMembershipsRequestObject) (gen.ListMembershipsResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) ListApplicationsForReview(_ context.Context, _ gen.ListApplicationsForReviewRequestObject) (gen.ListApplicationsForReviewResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) GetApplicationForReview(_ context.Context, _ gen.GetApplicationForReviewRequestObject) (gen.GetApplicationForReviewResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) ApproveApplication(_ context.Context, _ gen.ApproveApplicationRequestObject) (gen.ApproveApplicationResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) RecordExtractCheck(_ context.Context, _ gen.RecordExtractCheckRequestObject) (gen.RecordExtractCheckResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) DownloadExtractForReview(_ context.Context, _ gen.DownloadExtractForReviewRequestObject) (gen.DownloadExtractForReviewResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) RejectApplication(_ context.Context, _ gen.RejectApplicationRequestObject) (gen.RejectApplicationResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) ListAssociationsForOperator(_ context.Context, _ gen.ListAssociationsForOperatorRequestObject) (gen.ListAssociationsForOperatorResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) DisableAssociation(_ context.Context, _ gen.DisableAssociationRequestObject) (gen.DisableAssociationResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) EnableAssociation(_ context.Context, _ gen.EnableAssociationRequestObject) (gen.EnableAssociationResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}
