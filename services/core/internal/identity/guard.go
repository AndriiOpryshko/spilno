package identity

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
)

const (
	operatorPrefix = "/identity/operator/"
	sessionCookie  = "spilno_session"
)

// Membership is what the guard asks of the roster, on the transaction in ctx.
type Membership interface {
	CurrentMember(ctx context.Context, association uuid.UUID, email string, date time.Time) (bool, error)
}

var errServerFailure = errors.New("identity: handler answered 5xx")

// Guard is the session, roster and operator middleware. An /identity/operator path needs the
// оператор; any other path with {association_id} needs a current roster entry there and runs in
// db.InAssociation; every other path passes through.
func (m *Module) Guard(roster Membership) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			operator := strings.HasPrefix(r.URL.Path, operatorPrefix)
			rawID := chi.URLParam(r, "association_id")
			if !operator && rawID == "" {
				next.ServeHTTP(w, r)
				return
			}
			principal, err := m.principal(r)
			if errors.Is(err, app.ErrNoSession) {
				writeProblem(w, http.StatusUnauthorized, "Unauthorized", "Sign in to continue.", domain.CodeSessionRequired)
				return
			}
			if err != nil {
				writeProblem(w, http.StatusInternalServerError, "Internal Server Error", "", "")
				return
			}
			actor := Actor{AccountID: principal.AccountID.String(), Email: principal.Email, Operator: principal.IsOperator}
			if operator {
				if d := domain.Decide(actor, domain.OperatorPath, ""); !d.Allowed {
					writeProblem(w, http.StatusForbidden, "Forbidden", "Only the operator may do this.", d.Code)
					return
				}
				next.ServeHTTP(w, r.WithContext(withActor(r.Context(), actor)))
				return
			}
			m.serveAssociation(w, r, next, roster, actor, rawID)
		})
	}
}

func (m *Module) serveAssociation(w http.ResponseWriter, r *http.Request, next http.Handler, roster Membership, actor Actor, rawID string) {
	refuse := func() {
		writeProblem(w, http.StatusForbidden, "Forbidden", "You are not a current member of this association.", domain.CodeNotCurrentMember)
	}
	id, err := uuid.Parse(rawID)
	if err != nil {
		refuse()
		return
	}
	actor.AssociationID = rawID
	refused := false
	err = db.InAssociation(r.Context(), m.pool, id, actor.Email, func(ctx context.Context) error {
		ok, err := roster.CurrentMember(ctx, id, actor.Email, time.Now().UTC().Truncate(24*time.Hour))
		if err != nil {
			return err
		}
		actor.Member = ok
		if d := domain.Decide(actor, domain.AssociationPath, ""); !d.Allowed {
			refused = true
			return nil
		}
		rec := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r.WithContext(withActor(ctx, actor)))
		if rec.status >= http.StatusInternalServerError {
			return errServerFailure
		}
		return nil
	})
	switch {
	case refused:
		refuse()
	case errors.Is(err, errServerFailure):
	case err != nil:
		writeProblem(w, http.StatusInternalServerError, "Internal Server Error", "", "")
	}
}

func (m *Module) principal(r *http.Request) (domain.Principal, error) {
	if m.signIn == nil {
		return domain.Principal{}, app.ErrNoSession
	}
	c, err := r.Cookie(sessionCookie)
	if err != nil {
		return domain.Principal{}, app.ErrNoSession
	}
	return m.signIn.Principal(r.Context(), c.Value)
}

func writeProblem(w http.ResponseWriter, status int, title, detail, code string) {
	httpapi.WriteProblem(w, httpapi.Problem{Title: title, Status: status, Detail: detail, Code: code})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	if s.status == 0 {
		s.status = code
	}
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	return s.ResponseWriter.Write(b)
}

func (s *statusRecorder) Unwrap() http.ResponseWriter { return s.ResponseWriter }
