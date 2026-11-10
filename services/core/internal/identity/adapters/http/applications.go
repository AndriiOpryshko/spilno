package http

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/domain"
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

	fields := map[string]string{}
	var extract *domain.ExtractFile
	for {
		part, err := req.Body.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if part.FormName() == "extract" {
			b, err := io.ReadAll(io.LimitReader(part, domain.MaxExtractBytes+1))
			if err != nil {
				return nil, err
			}
			extract = &domain.ExtractFile{Name: truncate(part.FileName(), 255), Content: b}
			continue
		}
		b, err := io.ReadAll(io.LimitReader(part, 4096))
		if err != nil {
			return nil, err
		}
		fields[part.FormName()] = string(b)
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

	f := domain.Filing{
		Kind: kind, Edrpou: fields["edrpou"], Name: fields["association_name"],
		Address: fields["association_address"], ApplicantName: fields["applicant_full_name"],
		Position: fields["position"], Extract: extract,
	}
	f.TermStart, _ = time.Parse(dateLayout, fields["term_start"])
	if s := fields["term_end"]; s != "" {
		if t, err := time.Parse(dateLayout, s); err == nil {
			f.TermEnd = &t
		}
	}
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
