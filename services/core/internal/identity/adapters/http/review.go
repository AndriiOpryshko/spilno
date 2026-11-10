package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
)

// EnableReviews turns the оператор's queue, review page, extract download and check recording on.
func (h *Handler) EnableReviews(r *app.Reviews) { h.reviews = r }

func applicationNotFound() gen.NotFoundApplicationProblemPlusJSONResponse {
	detail := "No such application."
	return gen.NotFoundApplicationProblemPlusJSONResponse{Type: "about:blank", Title: "Not Found", Status: http.StatusNotFound, Detail: &detail}
}

func applicationBody(a ports.ReviewApplication) gen.Application {
	out := gen.Application{
		Id: a.ID, Kind: gen.ApplicationKind(a.Kind), State: gen.ApplicationState(a.State), Edrpou: a.Edrpou,
		AssociationName: a.AssociationName, CreatedAt: a.CreatedAt, FiledAt: a.FiledAt, RejectionReason: a.RejectionReason,
	}
	out.AssociationId = reviewUUIDPtr(a.AssociationID)
	return out
}

func reviewUUIDPtr(id *uuid.UUID) *openapi_types.UUID {
	if id == nil {
		return nil
	}
	v := openapi_types.UUID(*id)
	return &v
}

func checkBody(c ports.ReviewCheck) gen.ExtractCheck {
	return gen.ExtractCheck{
		Id: c.ID, FilingId: c.FilingID, SignatureValid: c.SignatureValid, SignerFullName: c.SignerFullName,
		SignedOn: openapi_types.Date{Time: c.SignedOn}, HeadFullName: c.HeadFullName, ExtractEdrpou: c.ExtractEdrpou,
		ExtractDatedOn: openapi_types.Date{Time: c.ExtractDatedOn}, CheckedAt: c.CheckedAt,
	}
}

func (h *Handler) ListApplicationsForReview(ctx context.Context, req gen.ListApplicationsForReviewRequestObject) (gen.ListApplicationsForReviewResponseObject, error) {
	if h.reviews == nil {
		return nil, httpapi.ErrNotImplemented
	}
	state, limit := string(domain.StatePending), defaultPageSize
	if req.Params.State != nil {
		state = string(*req.Params.State)
	}
	if req.Params.Limit != nil {
		limit = int(*req.Params.Limit)
	}
	var after *ports.AssociationCursor
	if req.Params.Cursor != nil {
		c, ok := decodeAssociationCursor(*req.Params.Cursor)
		if !ok {
			detail := "The cursor is not valid."
			return gen.ListApplicationsForReview422ApplicationProblemPlusJSONResponse{ValidationFailedApplicationProblemPlusJSONResponse: gen.ValidationFailedApplicationProblemPlusJSONResponse{
				Type: "about:blank", Title: "Unprocessable Entity", Status: http.StatusUnprocessableEntity, Detail: &detail,
				Errors: &[]gen.FieldError{{Pointer: "/cursor", Detail: detail}},
			}}, nil
		}
		after = c
	}
	rows, err := h.reviews.List(ctx, state, after, limit+1)
	if err != nil {
		return nil, err
	}
	page := gen.ApplicationPage{Items: []gen.Application{}}
	if len(rows) > limit {
		rows = rows[:limit]
		next := encodeAssociationCursor(ports.AssociationCursor{CreatedAt: rows[limit-1].CreatedAt, ID: rows[limit-1].ID})
		page.NextCursor = &next
	}
	for _, r := range rows {
		page.Items = append(page.Items, applicationBody(r))
	}
	return gen.ListApplicationsForReview200JSONResponse(page), nil
}

func (h *Handler) GetApplicationForReview(ctx context.Context, req gen.GetApplicationForReviewRequestObject) (gen.GetApplicationForReviewResponseObject, error) {
	if h.reviews == nil {
		return nil, httpapi.ErrNotImplemented
	}
	v, err := h.reviews.Get(ctx, req.ApplicationId)
	if errors.Is(err, ports.ErrApplicationNotFound) {
		return gen.GetApplicationForReview404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: applicationNotFound()}, nil
	}
	if err != nil {
		return nil, err
	}
	out := gen.ApplicationReview{
		Application: applicationBody(v.Application), ApplicantEmail: openapi_types.Email(v.ApplicantEmail),
		Checks: []gen.ExtractCheck{}, Decisions: []gen.ApplicationDecision{}, Filings: []gen.ApplicationFiling{},
		Warnings: []gen.ReviewWarning{}, CompetingApplicationIds: []openapi_types.UUID{},
	}
	for _, f := range v.Filings {
		af := gen.ApplicationFiling{
			Id: f.ID, ApplicantFullName: f.ApplicantFullName, AssociationAddress: f.AssociationAddress,
			AssociationName: f.AssociationName, CreatedAt: f.CreatedAt, Position: gen.BoardPosition(f.Position),
			TermStart: openapi_types.Date{Time: f.TermStart},
			Extract: gen.ExtractRef{
				Id: f.ExtractID, MediaType: f.ExtractMediaType, OriginalName: f.ExtractName,
				Sha256: fmt.Sprintf("%x", f.ExtractSha256), SizeBytes: f.ExtractSizeBytes,
			},
		}
		if f.TermEnd != nil {
			af.TermEnd = &openapi_types.Date{Time: *f.TermEnd}
		}
		out.Filings = append(out.Filings, af)
	}
	for _, c := range v.Checks {
		out.Checks = append(out.Checks, checkBody(c))
	}
	for _, d := range v.Decisions {
		ad := gen.ApplicationDecision{
			Id: d.ID, FilingId: d.FilingID, Outcome: gen.DecisionOutcome(d.Outcome), DecidedAt: d.DecidedAt,
			Reason: d.Reason, ReliedOnDocument: d.ReliedOnDocument,
		}
		ad.AssociationId = reviewUUIDPtr(d.AssociationID)
		out.Decisions = append(out.Decisions, ad)
	}
	if v.StorageAccountEmail != nil {
		e := openapi_types.Email(*v.StorageAccountEmail)
		out.StorageAccountEmail = &e
	}
	out.CompetingApplicationIds = append(out.CompetingApplicationIds, v.Competing...)
	for _, w := range v.Warnings {
		out.Warnings = append(out.Warnings, gen.ReviewWarning{Kind: gen.ReviewWarningKind(w.Kind), Detail: w.Detail})
	}
	return gen.GetApplicationForReview200JSONResponse(out), nil
}

func (h *Handler) DownloadExtractForReview(ctx context.Context, req gen.DownloadExtractForReviewRequestObject) (gen.DownloadExtractForReviewResponseObject, error) {
	if h.reviews == nil {
		return nil, httpapi.ErrNotImplemented
	}
	e, err := h.reviews.Extract(ctx, req.ApplicationId, req.FilingId)
	if errors.Is(err, ports.ErrApplicationNotFound) {
		return gen.DownloadExtractForReview404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: applicationNotFound()}, nil
	}
	if err != nil {
		return nil, err
	}
	disposition := fmt.Sprintf("attachment; filename=%q", e.OriginalName)
	nosniff := "nosniff"
	return gen.DownloadExtractForReview200ApplicationoctetStreamResponse{ExtractFileApplicationoctetStreamResponse: gen.ExtractFileApplicationoctetStreamResponse{
		Body: bytes.NewReader(e.Content), ContentLength: int64(len(e.Content)),
		Headers: gen.ExtractFileResponseHeaders{ContentDisposition: &disposition, XContentTypeOptions: &nosniff},
	}}, nil
}

func (h *Handler) RecordExtractCheck(ctx context.Context, req gen.RecordExtractCheckRequestObject) (gen.RecordExtractCheckResponseObject, error) {
	if h.reviews == nil || h.signIn == nil {
		return nil, httpapi.ErrNotImplemented
	}
	email, err := h.operatorEmail(ctx)
	if err != nil {
		return nil, err
	}
	b := req.Body
	c, err := h.reviews.RecordCheck(ctx, ports.NewCheck{
		ApplicationID: req.ApplicationId, FilingID: b.FilingId, SignatureValid: b.SignatureValid,
		SignerFullName: b.SignerFullName, SignedOn: b.SignedOn.Time, HeadFullName: b.HeadFullName,
		ExtractEdrpou: b.ExtractEdrpou, ExtractDatedOn: b.ExtractDatedOn.Time, CheckedByEmail: email,
	})
	if errors.Is(err, ports.ErrApplicationNotFound) {
		return gen.RecordExtractCheck404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: applicationNotFound()}, nil
	}
	if err != nil {
		return nil, err
	}
	return gen.RecordExtractCheck201JSONResponse(checkBody(c)), nil
}
