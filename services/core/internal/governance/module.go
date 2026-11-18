// Package governance is the governance bounded context: board meetings and their decisions.
//
// module.go is the only package other contexts and cmd/ may import from this
// context; everything below it is private to the context (enforced by depguard).
package governance

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/filesfake"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/gdrive"
	governancehttp "github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/http"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/postgres"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
)

type Module struct {
	handler *governancehttp.Handler
	roster  *app.RosterService
	files   *app.FileService
	cases   *app.CaseService
}

// Membership is a person's current seat on one association's board.
type Membership = ports.Membership

// Person is who is being seated on the board.
type Person struct{ FullName, SignInEmail string }

// Term is a board membership period; a zero End means open.
type Term = domain.Term

// New builds the module. hasStorage reports whether an association has connected its storage;
// cmd/api passes it from the identity module. access reports whether the request's association is
// read-only and why (ok false: no actor, the download is refused); cmd/api passes it from the identity guard's actor.
func New(_ *pgxpool.Pool, hasStorage func(context.Context, uuid.UUID) (bool, error), access func(context.Context) (readOnly bool, reason string, ok bool)) *Module {
	meetings := app.NewMeetingService(postgres.NewMeetings())
	roster := app.NewRosterService(postgres.NewRoster(), postgres.NewRoster())
	files := app.NewFileService(postgres.NewFiles())
	cases := app.NewCaseService(postgres.NewCases(), postgres.NewRoster())
	return &Module{handler: governancehttp.NewHandler(meetings, roster, app.NewRegistryService(postgres.NewRegistry()), files, cases, hasStorage, access), roster: roster, files: files, cases: cases}
}

// Files is the association-owned file storage the downloads read from.
type Files = ports.Files

// FilesResolver yields the storage of one association; nil Files means it has none.
type FilesResolver = ports.FilesResolver

// NewDevFiles is a resolver of in-memory storages, one per association, for development
// (SPILNO_DEV_STORAGE=true); they forget on restart.
func NewDevFiles() FilesResolver {
	var mu sync.Mutex
	stores := map[uuid.UUID]Files{}
	return func(_ context.Context, association uuid.UUID) (Files, error) {
		mu.Lock()
		defer mu.Unlock()
		if stores[association] == nil {
			stores[association] = filesfake.New()
		}
		return stores[association], nil
	}
}

// NewDriveFiles is the storage of one association on Google Drive: client is authorized as the
// association's Google account, every upload goes into folderID. The addresses have no trailing slash.
func NewDriveFiles(client *http.Client, folderID, apiURL, uploadURL string) Files {
	return gdrive.New(gdrive.Config{HTTP: client, APIURL: apiURL, UploadURL: uploadURL, FolderID: folderID})
}

// GoogleConnections is what the Drive resolver reads of an association's Google connection.
type GoogleConnections interface {
	GoogleTokenSource(ctx context.Context, association uuid.UUID) (oauth2.TokenSource, error)
	GoogleFolderID(ctx context.Context, association uuid.UUID) (string, error)
}

// NewDriveResolver resolves an association's Drive storage from its Google connection: nil Files when
// it has no connection (the error notConnected) or no folder yet. cmd/api and the module tests share it.
func NewDriveResolver(conns GoogleConnections, notConnected error, apiURL, uploadURL string) FilesResolver {
	return func(ctx context.Context, association uuid.UUID) (Files, error) {
		ts, err := conns.GoogleTokenSource(ctx, association)
		if errors.Is(err, notConnected) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		folder, err := conns.GoogleFolderID(ctx, association)
		if err != nil || folder == "" {
			return nil, err
		}
		return NewDriveFiles(oauth2.NewClient(ctx, ts), folder, apiURL, uploadURL), nil
	}
}

// WithFiles wires the storage behind downloadFile and addCaseDocument; without it they answer 503
// governance.storage_unavailable. Call it before Routes.
func (m *Module) WithFiles(files FilesResolver) *Module {
	m.files.SetFiles(files)
	m.cases.SetFiles(files)
	return m
}

// Routes attaches the context's routes to r inside their own group, validated
// against the context's tag-pruned spec.
func (m *Module) Routes(r chi.Router) error {
	spec, err := gen.GetSpec()
	if err != nil {
		return fmt.Errorf("governance: load spec: %w", err)
	}
	r.Group(func(r chi.Router) {
		r.Use(httpapi.BodyLimit(21<<20, 1<<20), httpapi.RequestValidator(spec))
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

// CurrentMemberID returns the board member id of the person whose sign-in address is email on date.
func (m *Module) CurrentMemberID(ctx context.Context, association uuid.UUID, email string, date time.Time) (uuid.UUID, bool, error) {
	return m.roster.CurrentMemberID(ctx, association, email, date)
}

// MemberSignInAddresses returns the sign-in addresses of the association's current members.
func (m *Module) MemberSignInAddresses(ctx context.Context, association uuid.UUID) ([]string, error) {
	return m.roster.MemberSignInAddresses(ctx, association)
}

// HasFiles reports whether the association has at least one stored file. Like the roster
// methods it reads on the transaction in ctx (db.InAssociation).
func (m *Module) HasFiles(ctx context.Context, association uuid.UUID) (bool, error) {
	return m.files.HasFiles(ctx, association)
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
