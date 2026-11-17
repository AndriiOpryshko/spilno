package http

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/ports"
)

func (h *Handler) ListCases(ctx context.Context, req gen.ListCasesRequestObject) (gen.ListCasesResponseObject, error) {
	var limit int
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}
	var cursor string
	if req.Params.Cursor != nil {
		cursor = *req.Params.Cursor
	}
	page, err := h.cases.List(ctx, req.AssociationId, limit, cursor)
	if errors.Is(err, app.ErrInvalidCursor) {
		return gen.ListCases422ApplicationProblemPlusJSONResponse{
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
	out := gen.ListCases200JSONResponse{Items: make([]gen.Case, len(page.Items))}
	for i, c := range page.Items {
		out.Items[i] = caseToAPI(c)
	}
	if page.NextCursor != "" {
		out.NextCursor = &page.NextCursor
	}
	return out, nil
}

func (h *Handler) CreateCase(ctx context.Context, req gen.CreateCaseRequestObject) (gen.CreateCaseResponseObject, error) {
	c, err := h.cases.Create(ctx, req.AssociationId, req.Body.Title)
	if v, ok := errors.AsType[*domain.ValidationError](err); ok {
		errs := make([]gen.FieldError, len(v.Fields))
		for i, f := range v.Fields {
			errs[i] = gen.FieldError{Pointer: bodyPointers[f.Field], Detail: f.Detail}
		}
		return gen.CreateCase422ApplicationProblemPlusJSONResponse{
			ValidationFailedApplicationProblemPlusJSONResponse: gen.ValidationFailedApplicationProblemPlusJSONResponse{
				Type: "about:blank", Title: "Validation failed", Status: http.StatusUnprocessableEntity, Errors: &errs,
			},
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return gen.CreateCase201JSONResponse(caseToAPI(c)), nil
}

func (h *Handler) GetCase(ctx context.Context, req gen.GetCaseRequestObject) (gen.GetCaseResponseObject, error) {
	c, err := h.cases.Get(ctx, req.AssociationId, req.CaseId)
	if errors.Is(err, domain.ErrCaseNotFound) {
		return gen.GetCase404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: caseNotFound(req.CaseId)}, nil
	}
	if err != nil {
		return nil, err
	}
	out := gen.GetCase200JSONResponse{
		Id: c.ID, Title: c.Title, CreatedByMemberId: c.CreatedByMemberID, CreatedAt: c.CreatedAt.UTC(),
		Documents: make([]gen.CaseDocument, len(c.Documents)),
	}
	for i, d := range c.Documents {
		out.Documents[i] = caseDocumentToAPI(d)
	}
	return out, nil
}

func (h *Handler) AddCaseDocument(ctx context.Context, req gen.AddCaseDocumentRequestObject) (gen.AddCaseDocumentResponseObject, error) {
	// The service opens its own transactions: the upload runs between them, outside any.
	connected, err := h.storage(ctx, req.AssociationId)
	if err != nil {
		return nil, err
	}
	if !connected {
		return gen.AddCaseDocument409ApplicationProblemPlusJSONResponse(driveNotConnectedProblem()), nil
	}
	up, issue, err := readUpload(req.Body)
	if err != nil {
		return nil, err
	}
	if issue != nil {
		return gen.AddCaseDocument422ApplicationProblemPlusJSONResponse(uploadProblem(nil, *issue)), nil
	}
	doc, err := h.cases.AddDocument(ctx, req.AssociationId, req.CaseId, up)
	switch {
	case errors.Is(err, domain.ErrCaseNotFound):
		return gen.AddCaseDocument404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: caseNotFound(req.CaseId)}, nil
	case errors.Is(err, domain.ErrPreviousVersionInvalid):
		return gen.AddCaseDocument422ApplicationProblemPlusJSONResponse(uploadProblem(nil, domain.FieldIssue{
			Pointer: "/previous_version_id", Detail: "is not a document of this case",
		})), nil
	}
	if fields, ok := errors.AsType[*domain.CodedFieldsError](err); ok {
		return gen.AddCaseDocument422ApplicationProblemPlusJSONResponse(uploadProblem(&fields.Code, fields.Fields...)), nil
	}
	if de, ok := errors.AsType[*domain.DomainError](err); ok && de.Code == domain.CodeStorageUnavailable {
		return gen.AddCaseDocument503ApplicationProblemPlusJSONResponse{
			StorageUnavailableApplicationProblemPlusJSONResponse: gen.StorageUnavailableApplicationProblemPlusJSONResponse{
				Type: "about:blank", Title: "Service Unavailable", Status: http.StatusServiceUnavailable,
				Code: ptr(domain.CodeStorageUnavailable), Detail: ptr("Сховище документів недоступне, спробуйте пізніше."),
			},
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return gen.AddCaseDocument201JSONResponse(caseDocumentToAPI(doc)), nil
}

// readUpload reads the multipart body: the file held in memory up to one byte past the limit, which the
// service then refuses, and the optional previous_version_id and document_date. A malformed field comes
// back as the issue to answer 422 with.
func readUpload(body *multipart.Reader) (app.Upload, *domain.FieldIssue, error) {
	var up app.Upload
	var haveFile bool
	for {
		part, err := body.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return app.Upload{}, nil, err
		}
		switch part.FormName() {
		case "file":
			data, err := io.ReadAll(io.LimitReader(part, domain.MaxDocumentSize+1))
			if err != nil {
				return app.Upload{}, nil, err
			}
			up.Name, up.Data, haveFile = part.FileName(), data, true
		case "previous_version_id":
			raw, err := io.ReadAll(io.LimitReader(part, 128))
			if err != nil {
				return app.Upload{}, nil, err
			}
			id, err := uuid.ParseBytes(raw)
			if err != nil {
				return app.Upload{}, &domain.FieldIssue{Pointer: "/previous_version_id", Detail: "must be a uuid"}, nil
			}
			up.PreviousVersionID = &id
		case "document_date":
			raw, err := io.ReadAll(io.LimitReader(part, 64))
			if err != nil {
				return app.Upload{}, nil, err
			}
			day, err := time.Parse("2006-01-02", string(raw))
			if err != nil {
				return app.Upload{}, &domain.FieldIssue{Pointer: "/document_date", Detail: "must be a date as YYYY-MM-DD"}, nil
			}
			up.DocumentDate = &day
		}
	}
	if !haveFile {
		return app.Upload{}, &domain.FieldIssue{Pointer: "/file", Detail: "is required"}, nil
	}
	return up, nil, nil
}

func uploadProblem(code *string, issues ...domain.FieldIssue) gen.Problem {
	errs := make([]gen.FieldError, len(issues))
	for i, f := range issues {
		errs[i] = gen.FieldError{Pointer: f.Pointer, Detail: f.Detail}
	}
	return gen.Problem{Type: "about:blank", Title: "Validation failed", Status: http.StatusUnprocessableEntity, Code: code, Errors: &errs}
}

func caseNotFound(id uuid.UUID) gen.NotFoundApplicationProblemPlusJSONResponse {
	return gen.NotFoundApplicationProblemPlusJSONResponse{
		Type: "about:blank", Title: "Not Found", Status: http.StatusNotFound,
		Detail: ptr("no case with id " + id.String()),
	}
}

func caseToAPI(c ports.CaseRecord) gen.Case {
	return gen.Case{Id: c.ID, Title: c.Title, CreatedByMemberId: c.CreatedByMemberID, CreatedAt: c.CreatedAt.UTC()}
}

func caseDocumentToAPI(d ports.CaseDocumentRecord) gen.CaseDocument {
	out := gen.CaseDocument{
		Id: d.ID, CaseId: d.CaseID, CreatedByMemberId: d.CreatedByMemberID, CreatedAt: d.CreatedAt.UTC(),
		PreviousVersionId: d.PreviousVersionID,
		File: gen.FileRef{
			Id: d.File.ID, OriginalName: d.File.Name, MediaType: gen.FileRefMediaType(d.File.MediaType),
			Sha256: d.File.SHA256, SizeBytes: d.File.Size, UploadedByMemberId: d.File.UploadedByMemberID,
			CreatedAt: d.File.CreatedAt.UTC(),
		},
	}
	if d.DocumentDate != nil {
		out.DocumentDate = &openapi_types.Date{Time: *d.DocumentDate}
	}
	return out
}
