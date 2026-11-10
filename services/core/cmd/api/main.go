// Command api runs the core service: every bounded context behind one HTTP server.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/oauth2/endpoints"

	"github.com/AndriiOpryshko/spilno/services/core/internal/finance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/config"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/google"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/observability"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/postgres"
	"github.com/AndriiOpryshko/spilno/services/core/internal/system"
)

// module is what every bounded context's module.go provides.
type module interface {
	Routes(r chi.Router) error
}

// rosterOf adapts the governance roster to the port the identity guard asks for.
type rosterOf struct{ gov *governance.Module }

func (g rosterOf) CurrentMember(ctx context.Context, association uuid.UUID, email string, date time.Time) (bool, error) {
	_, ok, err := g.gov.CurrentMember(ctx, association, email, date)
	return ok, err
}

// wireGoogle turns the Google connection and the storage connection operations on when the
// configuration is present; without it both stay off and the operations answer 501.
func wireGoogle(idm *identity.Module, gov *governance.Module, gc *identity.GoogleConfig) (bool, error) {
	if gc == nil {
		return false, nil
	}
	cfg := *gc
	if cfg.OAuth != nil && cfg.OAuth.Endpoint.TokenURL == "" {
		// A client without endpoints talks to Google's.
		o := *cfg.OAuth
		o.Endpoint = endpoints.Google
		cfg.OAuth = &o
	}
	if err := idm.EnableGoogle(cfg); err != nil {
		return false, err
	}
	idm.EnableStorageConnection(identity.StorageConnectionConfig{Consent: google.Consent{OAuth: cfg.OAuth}, Members: gov})
	return true, nil
}

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	logger, shutdownOTel, err := observability.Setup(ctx, cfg.ServiceName, cfg.OTLPEndpoint != "")
	if err != nil {
		return err
	}
	slog.SetDefault(logger)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownOTel(ctx); err != nil {
			slog.Error("otel shutdown", "err", err)
		}
	}()

	// goose runs as the owner; the API serves as a login role in spilno_app.
	migrationsPool, err := postgres.Open(ctx, cfg.MigrationsDatabaseURL)
	if err != nil {
		return err
	}
	err = postgres.Migrate(ctx, migrationsPool)
	migrationsPool.Close()
	if err != nil {
		return err
	}

	pool, err := postgres.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	r := httpapi.NewRouter()
	idm := identity.New(pool)
	_ = idm.RegisterMetrics()
	if cfg.DevSignIn {
		// Development and e2e only: signs in as the operator address, or a fixed test person.
		email := cfg.OperatorEmail
		if email == "" {
			email = "user@example.test"
		}
		idm.EnableSignIn(identity.SignInConfig{Provider: identity.DevSignInProvider(email), OperatorEmail: cfg.OperatorEmail})
	}
	gov := governance.New(pool)
	idm.EnableApprovals(identity.ApprovalConfig{SeatHead: func(ctx context.Context, association uuid.UUID, fullName, signInEmail string, termStart time.Time, decisionID uuid.UUID) (uuid.UUID, error) {
		return gov.SeatHead(ctx, association, governance.Person{FullName: fullName, SignInEmail: signInEmail}, governance.Term{Start: termStart}, decisionID)
	}, EndTerm: gov.EndTerm})
	idm.EnableMemberships(identity.MembershipsConfig{
		MembershipsOf: func(ctx context.Context, email string, date time.Time) ([]identity.RosterSeat, error) {
			ms, err := gov.MembershipsOf(ctx, email, date)
			out := make([]identity.RosterSeat, 0, len(ms))
			for _, m := range ms {
				out = append(out, identity.RosterSeat{AssociationID: m.AssociationID, Position: m.Position})
			}
			return out, err
		},
		CurrentMemberID: gov.CurrentMemberID,
	})
	gcfg, err := config.LoadGoogle()
	if err != nil {
		return err
	}
	var gc *identity.GoogleConfig
	if gcfg != nil {
		gc = &identity.GoogleConfig{KeyID: gcfg.TokenKeyID, Key: gcfg.TokenKey, OAuth: gcfg.OAuth(cfg.WebOrigin + "/storage/callback")}
	}
	wired, err := wireGoogle(idm, gov, gc)
	if err != nil {
		return err
	}
	if !wired {
		slog.Info("google connection is not configured")
	}
	if wired && cfg.DevStorage {
		slog.Warn("development storage consent is on; never use it in production")
		idm.EnableStorageConnection(identity.StorageConnectionConfig{
			Consent: google.DevConsent{ReturnURL: cfg.WebOrigin + "/storage/callback", Account: "osbb-storage@example.test"},
			Members: gov,
		})
	}
	for _, m := range []module{system.New(pool)} {
		if err := m.Routes(r); err != nil {
			return err
		}
	}
	// Association and operator routes sit behind the identity guard: session, roster, scoped transaction.
	var routeErr error
	r.Group(func(r chi.Router) {
		r.Use(idm.Guard(rosterOf{gov}))
		for _, m := range []module{finance.New(pool), gov, idm} {
			if routeErr = m.Routes(r); routeErr != nil {
				return
			}
		}
	})
	if routeErr != nil {
		return routeErr
	}

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           otelhttp.NewHandler(r, "http.server"),
		ReadHeaderTimeout: 5 * time.Second,
	}
	errc := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", cfg.HTTPAddr)
		errc <- srv.ListenAndServe()
	}()

	select {
	case err := <-errc:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		slog.Info("shutting down")
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
