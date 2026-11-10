// Package governance is the governance bounded context: board meetings and their decisions.
//
// module.go is the only package other contexts and cmd/ may import from this
// context; everything below it is private to the context (enforced by depguard).
package governance

import (
	"context"
	"fmt"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	governancehttp "github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/http"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/postgres"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
)

type Module struct {
	handler *governancehttp.Handler
	roster  *app.RosterService
}

// Membership is a person's current seat on one association's board.
type Membership = ports.Membership

// Person is who is being seated on the board.
type Person struct{ FullName, SignInEmail string }

// Term is a board membership period; a zero End means open.
type Term = domain.Term

func New(pool *pgxpool.Pool) *Module {
	meetings := app.NewMeetingService(postgres.NewMeetings())
	roster := app.NewRosterService(postgres.NewRoster(), postgres.NewRoster())
	return &Module{handler: governancehttp.NewHandler(meetings, identity.New(pool).HasStorageConnection), roster: roster}
}

// Routes attaches the context's routes to r inside their own group, validated
// against the context's tag-pruned spec.
func (m *Module) Routes(r chi.Router) error {
	spec, err := gen.GetSpec()
	if err != nil {
		return fmt.Errorf("governance: load spec: %w", err)
	}
	r.Group(func(r chi.Router) {
		r.Use(httpapi.RequestValidator(spec))
		strict := gen.NewStrictHandlerWithOptions(m.handler, nil, gen.StrictHTTPServerOptions{
			RequestErrorHandlerFunc:  httpapi.StrictRequestError,
			ResponseErrorHandlerFunc: httpapi.StrictResponseError,
		})
		gen.HandlerWithOptions(strict, gen.ChiServerOptions{BaseRouter: r, ErrorHandlerFunc: httpapi.ParamBindError})
	})
	return nil
}

// CurrentMember reports the seat email holds on the association's board on date.
// Like the other roster methods it reads on the transaction in ctx (db.InAssociation).
func (m *Module) CurrentMember(ctx context.Context, association uuid.UUID, email string, date time.Time) (Membership, bool, error) {
	return m.roster.CurrentMember(ctx, association, email, date)
}

// MembershipsOf returns every board email sits on at date.
func (m *Module) MembershipsOf(ctx context.Context, email string, date time.Time) ([]Membership, error) {
	return m.roster.MembershipsOf(ctx, email, date)
}

// MemberSignInAddresses returns the sign-in addresses of the association's current members.
func (m *Module) MemberSignInAddresses(ctx context.Context, association uuid.UUID) ([]string, error) {
	return m.roster.MemberSignInAddresses(ctx, association)
}

// SeatHead makes person the head under the decision on the transaction in ctx and returns
// the roster entry's id: a new entry, or a new revision of the person's current one.
func (m *Module) SeatHead(ctx context.Context, association uuid.UUID, person Person, term Term, decisionID uuid.UUID) (uuid.UUID, error) {
	return m.roster.SeatHead(ctx, association, person.FullName, person.SignInEmail, term, decisionID)
}

// EndTerm ends the whole roster entry on date under the decision, on the transaction in ctx.
func (m *Module) EndTerm(ctx context.Context, association, memberID uuid.UUID, date time.Time, decisionID uuid.UUID) error {
	return m.roster.EndTerm(ctx, association, memberID, date, decisionID)
}
