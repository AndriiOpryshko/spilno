package http

import (
	"context"
	"errors"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
)

// EnableRejections turns rejectApplication on.
func (h *Handler) EnableRejections(r *app.Rejections) { h.rejections = r }

func (h *Handler) RejectApplication(ctx context.Context, req gen.RejectApplicationRequestObject) (gen.RejectApplicationResponseObject, error) {
	if h.rejections == nil || h.signIn == nil || req.Body == nil {
		return nil, httpapi.ErrNotImplemented
	}
	p, err := h.signIn.Principal(ctx, tokenOf(ctx))
	if err != nil {
		return nil, err
	}
	d, err := h.rejections.Reject(ctx, ports.NewRejection{
		ApplicationID: req.ApplicationId, FilingID: req.Body.FilingId,
		OperatorAccountID: p.AccountID, OperatorEmail: p.Email, Reason: req.Body.Reason,
	})
	if errors.Is(err, ports.ErrApplicationNotFound) {
		return gen.RejectApplication404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: applicationNotFound()}, nil
	}
	if pr, ok := stateProblem(err); ok {
		if pr.Status == 422 {
			return gen.RejectApplication422ApplicationProblemPlusJSONResponse(pr), nil
		}
		return gen.RejectApplication409ApplicationProblemPlusJSONResponse(pr), nil
	}
	if err != nil {
		return nil, err
	}
	return gen.RejectApplication200JSONResponse{
		Id: d.ID, FilingId: d.FilingID, Outcome: gen.DecisionOutcome(d.Outcome), DecidedAt: d.DecidedAt, Reason: d.Reason,
	}, nil
}
