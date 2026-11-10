package http

import (
	"bytes"
	"context"
	"errors"
	"mime"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
)

// EnableCards turns getAssociationCard and downloadConfirmationExtract on.
func (h *Handler) EnableCards(c *app.Cards) { h.cards = c }

// EnableMemberships turns listMemberships on.
func (h *Handler) EnableMemberships(m *app.Memberships) { h.memberships = m }

func (h *Handler) ListMemberships(ctx context.Context, _ gen.ListMembershipsRequestObject) (gen.ListMembershipsResponseObject, error) {
	if h.memberships == nil || h.signIn == nil {
		return nil, httpapi.ErrNotImplemented
	}
	p, err := h.signIn.Principal(ctx, tokenOf(ctx))
	if errors.Is(err, app.ErrNoSession) {
		return gen.ListMemberships401ApplicationProblemPlusJSONResponse{SessionRequiredApplicationProblemPlusJSONResponse: sessionRequired()}, nil
	}
	if err != nil {
		return nil, err
	}
	ms, err := h.memberships.List(ctx, p.Email)
	if err != nil {
		return nil, err
	}
	page := gen.MembershipPage{Items: make([]gen.Membership, 0, len(ms))}
	for _, m := range ms {
		page.Items = append(page.Items, gen.Membership{
			AssociationId: m.AssociationID, AssociationName: m.AssociationName, AssociationState: gen.AssociationState(m.State),
			BoardMemberId: m.BoardMemberID, Position: gen.BoardPosition(m.Position),
		})
	}
	return gen.ListMemberships200JSONResponse(page), nil
}

func (h *Handler) GetAssociationCard(ctx context.Context, req gen.GetAssociationCardRequestObject) (gen.GetAssociationCardResponseObject, error) {
	if h.cards == nil {
		return nil, httpapi.ErrNotImplemented
	}
	c, err := h.cards.Card(ctx, req.AssociationId)
	if err != nil {
		return nil, err
	}
	out := gen.AssociationCard{
		Id: c.ID, Name: c.Name, State: gen.AssociationState(c.State), Edrpou: c.Edrpou, Address: c.Address,
		HeadFullName: c.HeadFullName, ConnectedByHand: c.ConnectedByHand, History: make([]gen.AssociationEvent, 0, len(c.History)),
	}
	for _, e := range c.History {
		ev := gen.AssociationEvent{
			Kind: gen.AssociationEventKind(e.Kind), OccurredAt: e.OccurredAt, DecisionId: e.DecisionID, FiledAt: e.FiledAt,
			Reason: e.Reason, HeadFullName: e.HeadFullName, SignerFullName: e.SignerFullName, ExtractId: e.ExtractID,
		}
		if e.ExtractDatedOn != nil {
			ev.ExtractDatedOn = &openapi_types.Date{Time: *e.ExtractDatedOn}
		}
		out.History = append(out.History, ev)
	}
	return gen.GetAssociationCard200JSONResponse(out), nil
}

func (h *Handler) DownloadConfirmationExtract(ctx context.Context, req gen.DownloadConfirmationExtractRequestObject) (gen.DownloadConfirmationExtractResponseObject, error) {
	if h.cards == nil {
		return nil, httpapi.ErrNotImplemented
	}
	e, err := h.cards.Extract(ctx, req.AssociationId, req.ExtractId)
	if errors.Is(err, ports.ErrCardNotFound) {
		return gen.DownloadConfirmationExtract404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: applicationNotFound()}, nil
	}
	if err != nil {
		return nil, err
	}
	disposition := mime.FormatMediaType("attachment", map[string]string{"filename": e.Name})
	nosniff := "nosniff"
	return gen.DownloadConfirmationExtract200ApplicationoctetStreamResponse{ExtractFileApplicationoctetStreamResponse: gen.ExtractFileApplicationoctetStreamResponse{
		Body: bytes.NewReader(e.Content), ContentLength: int64(len(e.Content)),
		Headers: gen.ExtractFileResponseHeaders{ContentDisposition: &disposition, XContentTypeOptions: &nosniff},
	}}, nil
}
