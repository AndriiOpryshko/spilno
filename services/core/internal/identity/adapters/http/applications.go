package http

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
)

const dateLayout = "2006-01-02"

func (h *Handler) FileApplication(ctx context.Context, req gen.FileApplicationRequestObject) (gen.FileApplicationResponseObject, error) {
	if h.signIn == nil || h.applications == nil {
		return nil, httpapi.ErrNotImplemented
	}
	p, err := h.signIn.Principal(ctx, tokenOf(ctx))
	if errors.Is(err, app.ErrNoSession) {
		return gen.FileApplication401ApplicationProblemPlusJSONResponse{SessionRequiredApplicationProblemPlusJSONResponse: sessionRequired()}, nil
	}
	if err != nil {
		return nil, err
	}

	fields, extract, err := readFilingParts(req.Body)
	if err != nil {
		return nil, err
	}
	kind := domain.FilingKind(fields["kind"])
	if kind != domain.FilingConnection && kind != domain.FilingTransfer {
		return nil, httpapi.ErrNotImplemented
	}
	var assocID *uuid.UUID
	if kind == domain.FilingTransfer && fields["association_id"] != "" {
		id, err := uuid.Parse(fields["association_id"])
		if err != nil {
			id = uuid.Nil
		}
		assocID = &id
	}

	f := filingFromFields(fields, extract)
	f.Kind, f.Edrpou = kind, fields["edrpou"]
	out, err := h.applications.File(ctx, app.FileInput{AccountID: p.AccountID, Email: p.Email, AssociationID: assocID, Filing: f})
	var fe *domain.CodedFieldsError
	var ce *domain.CodedError
	switch {
	case errors.As(err, &fe):
		return fileProblem(fe.Code, "The application is not valid.", fieldErrors(fe.Fields)), nil
	case errors.As(err, &ce):
		return fileProblem(ce.Code, "The application cannot be filed.", nil), nil
	case err != nil:
		return nil, err
	}
	return gen.FileApplication201JSONResponse{
		Id: out.ID, Kind: gen.ApplicationKind(f.Kind), State: gen.ApplicationState(domain.StatePending),
		Edrpou: f.Edrpou, AssociationName: f.Name, CreatedAt: out.CreatedAt, FiledAt: out.FiledAt,
	}, nil
}

// readFilingParts reads the text fields and the extract file of a multipart filing.
func readFilingParts(body *multipart.Reader) (map[string]string, *domain.ExtractFile, error) {
	fields := map[string]string{}
	var extract *domain.ExtractFile
	for {
		part, err := body.NextPart()
		if errors.Is(err, io.EOF) {
			return fields, extract, nil
		}
		if err != nil {
			return nil, nil, err
		}
		if part.FormName() == "extract" {
			b, err := io.ReadAll(io.LimitReader(part, domain.MaxExtractBytes+1))
			if err != nil {
				return nil, nil, err
			}
			extract = &domain.ExtractFile{Name: truncate(part.FileName(), 255), Content: b}
			continue
		}
		b, err := io.ReadAll(io.LimitReader(part, 4096))
		if err != nil {
			return nil, nil, err
		}
		fields[part.FormName()] = string(b)
	}
}

// filingFromFields builds the kind-independent part of a filing from multipart text fields.
func filingFromFields(fields map[string]string, extract *domain.ExtractFile) domain.Filing {
	f := domain.Filing{
		Name: fields["association_name"], Address: fields["association_address"],
		ApplicantName: fields["applicant_full_name"], Position: fields["position"], Extract: extract,
	}
	f.TermStart, _ = time.Parse(dateLayout, fields["term_start"])
	if s := fields["term_end"]; s != "" {
		if t, err := time.Parse(dateLayout, s); err == nil {
			f.TermEnd = &t
		}
	}
	return f
}

func truncate(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n])
	}
	return s
}

func fieldErrors(in []domain.FieldIssue) *[]gen.FieldError {
	out := make([]gen.FieldError, len(in))
	for i, f := range in {
		out[i] = gen.FieldError{Pointer: f.Pointer, Detail: f.Detail}
	}
	return &out
}

func fileProblem(code, detail string, errs *[]gen.FieldError) gen.FileApplicationResponseObject {
	status := http.StatusConflict
	for _, d := range domain.ErrorRegistry {
		if d.Code == code {
			status = d.Status
		}
	}
	p := gen.Problem{Type: "about:blank", Title: http.StatusText(status), Status: status, Detail: &detail, Code: &code, Errors: errs}
	if status == http.StatusUnprocessableEntity {
		return gen.FileApplication422ApplicationProblemPlusJSONResponse(p)
	}
	return gen.FileApplication409ApplicationProblemPlusJSONResponse(p)
}

func myApplicationBody(v ports.ApplicationView) gen.Application {
	return gen.Application{
		Id: v.ID, Kind: gen.ApplicationKind(v.Kind), State: gen.ApplicationState(v.State), Edrpou: v.Edrpou,
		AssociationId: v.AssociationID, AssociationName: v.AssociationName, CreatedAt: v.CreatedAt,
		FiledAt: v.FiledAt, RejectionReason: v.RejectionReason,
	}
}

func encodeApplicationCursor(c ports.ApplicationCursor) string {
	return base64.RawURLEncoding.EncodeToString([]byte(c.CreatedAt.UTC().Format(time.RFC3339Nano) + "|" + c.ID.String()))
}

func decodeApplicationCursor(s string) (*ports.ApplicationCursor, bool) {
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
	return &ports.ApplicationCursor{CreatedAt: t, ID: u}, true
}

func (h *Handler) ListMyApplications(ctx context.Context, req gen.ListMyApplicationsRequestObject) (gen.ListMyApplicationsResponseObject, error) {
	if h.signIn == nil || h.applications == nil {
		return nil, httpapi.ErrNotImplemented
	}
	p, err := h.signIn.Principal(ctx, tokenOf(ctx))
	if errors.Is(err, app.ErrNoSession) {
		return gen.ListMyApplications401ApplicationProblemPlusJSONResponse{SessionRequiredApplicationProblemPlusJSONResponse: sessionRequired()}, nil
	}
	if err != nil {
		return nil, err
	}
	limit := defaultPageSize
	if req.Params.Limit != nil {
		limit = int(*req.Params.Limit)
	}
	var after *ports.ApplicationCursor
	if req.Params.Cursor != nil {
		c, ok := decodeApplicationCursor(*req.Params.Cursor)
		if !ok {
			detail := "The cursor is not valid."
			return gen.ListMyApplications422ApplicationProblemPlusJSONResponse{ValidationFailedApplicationProblemPlusJSONResponse: gen.ValidationFailedApplicationProblemPlusJSONResponse{
				Type: "about:blank", Title: "Unprocessable Entity", Status: http.StatusUnprocessableEntity, Detail: &detail,
				Errors: &[]gen.FieldError{{Pointer: "/cursor", Detail: detail}},
			}}, nil
		}
		after = c
	}
	rows, err := h.applications.ListMine(ctx, p.AccountID, after, limit+1)
	if err != nil {
		return nil, err
	}
	page := gen.ApplicationPage{Items: []gen.Application{}}
	if len(rows) > limit {
		rows = rows[:limit]
		next := encodeApplicationCursor(ports.ApplicationCursor{CreatedAt: rows[limit-1].CreatedAt, ID: rows[limit-1].ID})
		page.NextCursor = &next
	}
	for _, r := range rows {
		page.Items = append(page.Items, myApplicationBody(r))
	}
	return gen.ListMyApplications200JSONResponse(page), nil
}

func (h *Handler) GetMyApplication(ctx context.Context, req gen.GetMyApplicationRequestObject) (gen.GetMyApplicationResponseObject, error) {
	if h.signIn == nil || h.applications == nil {
		return nil, httpapi.ErrNotImplemented
	}
	p, err := h.signIn.Principal(ctx, tokenOf(ctx))
	if errors.Is(err, app.ErrNoSession) {
		return gen.GetMyApplication401ApplicationProblemPlusJSONResponse{SessionRequiredApplicationProblemPlusJSONResponse: sessionRequired()}, nil
	}
	if err != nil {
		return nil, err
	}
	d, err := h.applications.GetMine(ctx, p.AccountID, req.ApplicationId)
	if errors.Is(err, ports.ErrApplicationNotFound) {
		return gen.GetMyApplication404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: applicationNotFound()}, nil
	}
	if err != nil {
		return nil, err
	}
	out := gen.ApplicationDetail{Application: myApplicationBody(d.Application), Filings: []gen.ApplicationFiling{}, Decisions: []gen.ApplicationDecision{}}
	for _, f := range d.Filings {
		item := gen.ApplicationFiling{
			Id: f.ID, ApplicantFullName: f.ApplicantFullName, AssociationAddress: f.AssociationAddress,
			AssociationName: f.AssociationName, CreatedAt: f.CreatedAt, Position: gen.BoardPosition(f.Position),
			TermStart: openapi_types.Date{Time: f.TermStart},
			Extract: gen.ExtractRef{
				Id: f.ExtractID, MediaType: f.MediaType, OriginalName: f.OriginalName,
				Sha256: hex.EncodeToString(f.Sha256), SizeBytes: f.SizeBytes,
			},
		}
		if f.TermEnd != nil {
			item.TermEnd = &openapi_types.Date{Time: *f.TermEnd}
		}
		out.Filings = append(out.Filings, item)
	}
	for _, x := range d.Decisions {
		out.Decisions = append(out.Decisions, gen.ApplicationDecision{
			Id: x.ID, FilingId: x.FilingID, AssociationId: x.AssociationID, Outcome: gen.DecisionOutcome(x.Outcome),
			Reason: x.Reason, ReliedOnDocument: x.ReliedOnDocument, DecidedAt: x.CreatedAt,
		})
	}
	return gen.GetMyApplication200JSONResponse(out), nil
}

func (h *Handler) RefileApplication(ctx context.Context, req gen.RefileApplicationRequestObject) (gen.RefileApplicationResponseObject, error) {
	if h.signIn == nil || h.applications == nil {
		return nil, httpapi.ErrNotImplemented
	}
	p, err := h.signIn.Principal(ctx, tokenOf(ctx))
	if errors.Is(err, app.ErrNoSession) {
		return gen.RefileApplication401ApplicationProblemPlusJSONResponse{SessionRequiredApplicationProblemPlusJSONResponse: sessionRequired()}, nil
	}
	if err != nil {
		return nil, err
	}
	fields, extract, err := readFilingParts(req.Body)
	if err != nil {
		return nil, err
	}
	f := filingFromFields(fields, extract)
	out, err := h.applications.Refile(ctx, app.RefileInput{AccountID: p.AccountID, ApplicationID: req.ApplicationId, Filing: f})
	var fe *domain.CodedFieldsError
	var ce *domain.CodedError
	switch {
	case errors.Is(err, ports.ErrApplicationNotFound):
		return gen.RefileApplication404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: applicationNotFound()}, nil
	case errors.As(err, &fe):
		detail := "The application is not valid."
		return gen.RefileApplication422ApplicationProblemPlusJSONResponse(refileProblem(fe.Code, detail, fieldErrors(fe.Fields))), nil
	case errors.As(err, &ce):
		return refileRefusal(ce.Code), nil
	case err != nil:
		return nil, err
	}
	return gen.RefileApplication201JSONResponse(myApplicationBody(out)), nil
}

func refileProblem(code, detail string, errs *[]gen.FieldError) gen.Problem {
	status := http.StatusConflict
	for _, d := range domain.ErrorRegistry {
		if d.Code == code {
			status = d.Status
		}
	}
	return gen.Problem{Type: "about:blank", Title: http.StatusText(status), Status: status, Detail: &detail, Code: &code, Errors: errs}
}

func refileRefusal(code string) gen.RefileApplicationResponseObject {
	p := refileProblem(code, "The application cannot be filed again.", nil)
	if p.Status == http.StatusUnprocessableEntity {
		return gen.RefileApplication422ApplicationProblemPlusJSONResponse(p)
	}
	return gen.RefileApplication409ApplicationProblemPlusJSONResponse(p)
}

func (h *Handler) DownloadMyExtract(ctx context.Context, req gen.DownloadMyExtractRequestObject) (gen.DownloadMyExtractResponseObject, error) {
	if h.signIn == nil || h.applications == nil {
		return nil, httpapi.ErrNotImplemented
	}
	p, err := h.signIn.Principal(ctx, tokenOf(ctx))
	if errors.Is(err, app.ErrNoSession) {
		return gen.DownloadMyExtract401ApplicationProblemPlusJSONResponse{SessionRequiredApplicationProblemPlusJSONResponse: sessionRequired()}, nil
	}
	if err != nil {
		return nil, err
	}
	e, err := h.applications.Extract(ctx, p.AccountID, req.ApplicationId, req.FilingId)
	if errors.Is(err, ports.ErrApplicationNotFound) {
		return gen.DownloadMyExtract404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: applicationNotFound()}, nil
	}
	if err != nil {
		return nil, err
	}
	disposition := mime.FormatMediaType("attachment", map[string]string{"filename": e.Name})
	nosniff := "nosniff"
	return gen.DownloadMyExtract200ApplicationoctetStreamResponse{ExtractFileApplicationoctetStreamResponse: gen.ExtractFileApplicationoctetStreamResponse{
		Body: bytes.NewReader(e.Content), ContentLength: int64(len(e.Content)),
		Headers: gen.ExtractFileResponseHeaders{ContentDisposition: &disposition, XContentTypeOptions: &nosniff},
	}}, nil
}
