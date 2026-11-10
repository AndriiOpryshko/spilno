// Package http adapts the identity context to the generated strict server.
package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
)

// Handler serves the identity operations; those not built yet answer 501.
type Handler struct {
	signIn       *app.SignIn // nil until EnableSignIn
	applications *app.Applications
	states       *app.AssociationStates  // nil until EnableAssociationStates
	reviews      *app.Reviews            // nil until EnableReviews
	storage      *app.StorageConnections // nil until EnableStorageConnections
}

var _ gen.StrictServerInterface = (*Handler)(nil)

func NewHandler() *Handler { return &Handler{} }

// EnableSignIn turns the four session operations on.
func (h *Handler) EnableSignIn(s *app.SignIn) { h.signIn = s }

// EnableApplications turns fileApplication on.
func (h *Handler) EnableApplications(a *app.Applications) { h.applications = a }

const sessionCookie = "spilno_session"

type tokenKey struct{}

// CaptureSession puts the session cookie's token in the request context, since the
// strict handlers do not see the request.
func CaptureSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c, err := r.Cookie(sessionCookie); err == nil {
			r = r.WithContext(context.WithValue(r.Context(), tokenKey{}, c.Value))
		}
		next.ServeHTTP(w, r)
	})
}

func tokenOf(ctx context.Context) string {
	t, _ := ctx.Value(tokenKey{}).(string)
	return t
}

func sessionRequired() gen.SessionRequiredApplicationProblemPlusJSONResponse {
	code := domain.CodeSessionRequired
	detail := "Sign in to continue."
	return gen.SessionRequiredApplicationProblemPlusJSONResponse{
		Type: "about:blank", Title: "Unauthorized", Status: http.StatusUnauthorized, Detail: &detail, Code: &code,
	}
}

func cookieHeader(token string, maxAge int) *string {
	c := (&http.Cookie{Name: sessionCookie, Value: token, Path: "/", MaxAge: maxAge, HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode}).String()
	return &c
}

func (h *Handler) GetSession(ctx context.Context, _ gen.GetSessionRequestObject) (gen.GetSessionResponseObject, error) {
	if h.signIn == nil {
		return nil, httpapi.ErrNotImplemented
	}
	p, err := h.signIn.Principal(ctx, tokenOf(ctx))
	if errors.Is(err, app.ErrNoSession) {
		return gen.GetSession401ApplicationProblemPlusJSONResponse{SessionRequiredApplicationProblemPlusJSONResponse: sessionRequired()}, nil
	}
	if err != nil {
		return nil, err
	}
	return gen.GetSession200JSONResponse{AccountId: p.AccountID, Email: openapi_types.Email(p.Email), IsOperator: p.IsOperator}, nil
}

func (h *Handler) StartSignIn(_ context.Context, _ gen.StartSignInRequestObject) (gen.StartSignInResponseObject, error) {
	if h.signIn == nil {
		return nil, httpapi.ErrNotImplemented
	}
	loc, err := h.signIn.Start()
	if err != nil {
		return nil, err
	}
	return gen.StartSignIn302Response{Headers: gen.StartSignIn302ResponseHeaders{Location: &loc}}, nil
}

func (h *Handler) CompleteSignIn(ctx context.Context, req gen.CompleteSignInRequestObject) (gen.CompleteSignInResponseObject, error) {
	if h.signIn == nil {
		return nil, httpapi.ErrNotImplemented
	}
	token, err := h.signIn.Complete(ctx, req.Params.Code, req.Params.State)
	if errors.Is(err, domain.ErrSignInRefused) {
		code := domain.CodeSessionRequired
		detail := "Sign-in was refused; start it again."
		return gen.CompleteSignIn401ApplicationProblemPlusJSONResponse{
			Type: "about:blank", Title: "Unauthorized", Status: http.StatusUnauthorized, Detail: &detail, Code: &code,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	loc := "/"
	return gen.CompleteSignIn302Response{Headers: gen.CompleteSignIn302ResponseHeaders{Location: &loc, SetCookie: cookieHeader(token, int(30*24*time.Hour/time.Second))}}, nil
}

func (h *Handler) SignOut(ctx context.Context, _ gen.SignOutRequestObject) (gen.SignOutResponseObject, error) {
	if h.signIn == nil {
		return nil, httpapi.ErrNotImplemented
	}
	err := h.signIn.SignOut(ctx, tokenOf(ctx))
	if errors.Is(err, app.ErrNoSession) {
		return gen.SignOut401ApplicationProblemPlusJSONResponse{SessionRequiredApplicationProblemPlusJSONResponse: sessionRequired()}, nil
	}
	if err != nil {
		return nil, err
	}
	return gen.SignOut204Response{Headers: gen.SignOut204ResponseHeaders{SetCookie: cookieHeader("", -1)}}, nil
}

func (h *Handler) GetAssociationCard(_ context.Context, _ gen.GetAssociationCardRequestObject) (gen.GetAssociationCardResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) DownloadConfirmationExtract(_ context.Context, _ gen.DownloadConfirmationExtractRequestObject) (gen.DownloadConfirmationExtractResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) ListMemberships(_ context.Context, _ gen.ListMembershipsRequestObject) (gen.ListMembershipsResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) ApproveApplication(_ context.Context, _ gen.ApproveApplicationRequestObject) (gen.ApproveApplicationResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) RejectApplication(_ context.Context, _ gen.RejectApplicationRequestObject) (gen.RejectApplicationResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}
