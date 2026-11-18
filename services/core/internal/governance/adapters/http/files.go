package http

import (
	"context"
	"errors"
	"mime"
	"net/http"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
)

const (
	disabledDetail  = "Доступ призупинено"
	notMemberDetail = "You are not a current member of this association."
)

func (h *Handler) DownloadFile(ctx context.Context, req gen.DownloadFileRequestObject) (gen.DownloadFileResponseObject, error) {
	var acc Access
	if h.access != nil {
		acc = h.access(ctx)
	}
	if !acc.Actor {
		return gen.DownloadFile403ApplicationProblemPlusJSONResponse{
			Type: "about:blank", Title: "Forbidden", Status: http.StatusForbidden,
			Code: ptr(acc.NotMemberCode), Detail: ptr(notMemberDetail),
		}, nil
	}
	d, err := h.files.Download(ctx, req.AssociationId, req.FileId, acc.ReadOnly)
	switch {
	case errors.Is(err, app.ErrFileNotFound):
		return gen.DownloadFile404ApplicationProblemPlusJSONResponse{
			NotFoundApplicationProblemPlusJSONResponse: gen.NotFoundApplicationProblemPlusJSONResponse{
				Type: "about:blank", Title: "Not Found", Status: http.StatusNotFound,
				Detail: ptr("no file with id " + req.FileId.String()),
			},
		}, nil
	case errors.Is(err, app.ErrFileNotSigned):
		detail := acc.Reason
		if detail == "" {
			detail = disabledDetail
		}
		return gen.DownloadFile403ApplicationProblemPlusJSONResponse{
			Type: "about:blank", Title: "Forbidden", Status: http.StatusForbidden,
			Code: ptr(acc.DisabledCode), Detail: ptr(detail),
		}, nil
	}
	if de, ok := errors.AsType[*domain.DomainError](err); ok && de.Code == domain.CodeStorageUnavailable {
		return gen.DownloadFile503ApplicationProblemPlusJSONResponse{
			StorageUnavailableApplicationProblemPlusJSONResponse: gen.StorageUnavailableApplicationProblemPlusJSONResponse{
				Type: "about:blank", Title: "Service Unavailable", Status: http.StatusServiceUnavailable,
				Code: ptr(domain.CodeStorageUnavailable), Detail: ptr("Сховище документів недоступне, спробуйте пізніше."),
			},
		}, nil
	}
	if err != nil {
		return nil, err
	}
	headers := gen.DownloadFile200ResponseHeaders{
		ContentDisposition:  ptr(mime.FormatMediaType("attachment", map[string]string{"filename": d.Name})),
		XContentTypeOptions: ptr("nosniff"),
	}
	switch d.MediaType {
	case "image/jpeg":
		return gen.DownloadFile200ImagejpegResponse{Body: d.Body, Headers: headers, ContentLength: d.Size}, nil
	case "image/png":
		return gen.DownloadFile200ImagepngResponse{Body: d.Body, Headers: headers, ContentLength: d.Size}, nil
	}
	return gen.DownloadFile200ApplicationpdfResponse{Body: d.Body, Headers: headers, ContentLength: d.Size}, nil
}
