package identity

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/postgres"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/kyivday"
)

const (
	operatorPrefix = "/identity/operator/"
	sessionCookie  = "spilno_session"

	notMemberDetail = "You are not a current member of this association."
	disabledDetail  = "This association is disabled and read-only."
)

// Membership is what the guard asks of the roster, on the transaction in ctx.
type Membership interface {
	CurrentMember(ctx context.Context, association uuid.UUID, email string, date time.Time) (bool, error)
}

var errHandlerRefused = errors.New("identity: handler answered 4xx or 5xx")

// Guard is the session, roster and operator middleware. An /identity/operator path needs the
// оператор; any other path with {association_id} needs a current roster entry there and runs in
// db.InAssociation; every other path passes through.
func (m *Module) Guard(roster Membership) func(http.Handler) http.Handler {
	m.applications.UseRoster(roster)
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
	code, detail := domain.CodeNotCurrentMember, notMemberDetail
	refuse := func() { writeProblem(w, http.StatusForbidden, "Forbidden", detail, code) }
	id, err := uuid.Parse(rawID)
	if err != nil {
		refuse()
		return
	}
	actor.AssociationID = rawID
	refused := false
	outsideIO := outsideIOOperations[r.Method+" "+routePattern(r)]
	var buf *bufferedResponse
	err = db.InAssociation(r.Context(), m.pool, id, actor.Email, func(ctx context.Context) error {
		ok, err := roster.CurrentMember(ctx, id, actor.Email, kyivday.Today(m.now()))
		if err != nil {
			return err
		}
		actor.Member = ok
		if ok {
			disabled, reason, err := postgres.DisabledState(ctx, id)
			if err != nil {
				return err
			}
			actor.ReadOnly, actor.DisabledReason = disabled, reason
		}
		if d := domain.Decide(actor, domain.AssociationPath, operationID(r)); !d.Allowed {
			refused = true
			code, detail = d.Code, notMemberDetail
			if d.Code == domain.CodeAssociationDisabled {
				detail = d.Reason
				if detail == "" {
					detail = disabledDetail
				}
			}
			return nil
		}
		if outsideIO {
			return nil
		}
		buf = newBufferedResponse()
		next.ServeHTTP(buf, r.WithContext(withActor(ctx, actor)))
		if buf.status >= http.StatusBadRequest {
			return errHandlerRefused
		}
		return nil
	})
	switch {
	case refused:
		refuse()
	case err == nil && outsideIO:
		// The check transaction has committed and holds no connection. The handler opens its own
		// association transactions with db.InScope and writes straight to the client.
		ctx := db.WithScope(withActor(r.Context(), actor), m.pool, id, actor.Email)
		next.ServeHTTP(w, r.WithContext(ctx))
	case errors.Is(err, errHandlerRefused):
		buf.flushTo(w)
	case err != nil:
		writeProblem(w, http.StatusInternalServerError, "Internal Server Error", "", "")
	default:
		buf.flushTo(w)
	}
}

// readOnlyOperations are the operations a disabled association still serves, by method and route.
var readOnlyOperations = map[string]string{
	"GET /governance/associations/{association_id}/registry/decisions":               "searchRegistry",
	"GET /governance/associations/{association_id}/registry/protocols/{protocol_id}": "getSignedProtocol",
	"GET /governance/associations/{association_id}/files/{file_id}":                  "downloadFile",
}

// outsideIOOperations are the operations that talk to an outside service (file storage, the Google
// code exchange) or stream a large body. They run after the guard's check transaction, not inside it.
var outsideIOOperations = map[string]bool{
	"POST /governance/associations/{association_id}/cases/{case_id}/documents": true, // addCaseDocument
	"GET /governance/associations/{association_id}/files/{file_id}":            true, // downloadFile
	"POST /identity/associations/{association_id}/storage-connection":          true, // completeStorageConnection
}

func routePattern(r *http.Request) string {
	if rc := chi.RouteContext(r.Context()); rc != nil {
		return rc.RoutePattern()
	}
	return ""
}

// operationID names the operation the request routed to when the decision table cares about it.
func operationID(r *http.Request) string {
	rc := chi.RouteContext(r.Context())
	if rc == nil {
		return ""
	}
	return readOnlyOperations[r.Method+" "+rc.RoutePattern()]
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

// bufferedResponse holds a handler's response until the association transaction has committed.
type bufferedResponse struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func newBufferedResponse() *bufferedResponse { return &bufferedResponse{header: http.Header{}} }

func (b *bufferedResponse) Header() http.Header { return b.header }

func (b *bufferedResponse) WriteHeader(code int) {
	if b.status == 0 {
		b.status = code
	}
}

func (b *bufferedResponse) Write(p []byte) (int, error) {
	if b.status == 0 {
		b.status = http.StatusOK
	}
	return b.body.Write(p)
}

// Flush is a no-op: nothing leaves before the commit.
func (b *bufferedResponse) Flush() {}

func (b *bufferedResponse) flushTo(w http.ResponseWriter) {
	for k, v := range b.header {
		w.Header()[k] = v
	}
	if b.status == 0 {
		b.status = http.StatusOK
	}
	w.WriteHeader(b.status)
	_, _ = w.Write(b.body.Bytes())
}
