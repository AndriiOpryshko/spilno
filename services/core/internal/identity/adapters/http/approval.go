package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
)

// EnableApprovals turns approveApplication on.
func (h *Handler) EnableApprovals(a *app.Approvals) { h.approvals = a }

func (h *Handler) ApproveApplication(ctx context.Context, req gen.ApproveApplicationRequestObject) (gen.ApproveApplicationResponseObject, error) {
	if h.approvals == nil || h.signIn == nil {
		return nil, httpapi.ErrNotImplemented
	}
	p, err := h.signIn.Principal(ctx, tokenOf(ctx))
	if err != nil {
		return nil, err
	}
	in := app.ApproveInput{
		ApprovalRequest:   ports.ApprovalRequest{ApplicationID: req.ApplicationId, FilingID: req.Body.FilingId, OperatorEmail: p.Email},
		OperatorAccountID: p.AccountID.String(),
	}
	if req.Body.ReliedOnDocument != nil {
		in.Document = *req.Body.ReliedOnDocument
	}
	d, err := h.approvals.Approve(ctx, in)
	var ce *domain.CodedError
	switch {
	case errors.Is(err, ports.ErrApplicationNotFound):
		return gen.ApproveApplication404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: applicationNotFound()}, nil
	case errors.Is(err, ports.ErrAssociationConnected):
		return approvalConflict(domain.CodeAssociationAlreadyConnected), nil
	case errors.As(err, &ce):
		return approvalConflict(ce.Code), nil
	case err != nil:
		return nil, err
	}
	out := gen.ApplicationDecision{
		Id: d.ID, FilingId: d.FilingID, Outcome: gen.DecisionOutcome(d.Outcome), DecidedAt: d.DecidedAt,
		Reason: d.Reason, ReliedOnDocument: d.ReliedOnDocument,
	}
	out.AssociationId = reviewUUIDPtr(d.AssociationID)
	return gen.ApproveApplication200JSONResponse(out), nil
}

func approvalConflict(code string) gen.ApproveApplication409ApplicationProblemPlusJSONResponse {
	detail := "The application cannot be approved."
	return gen.ApproveApplication409ApplicationProblemPlusJSONResponse{
		Type: "about:blank", Title: http.StatusText(http.StatusConflict), Status: http.StatusConflict, Detail: &detail, Code: &code,
	}
}
