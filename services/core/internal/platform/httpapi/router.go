package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter returns the root router shared by all bounded contexts. Contexts
// attach their routes in their own r.Group (see each context's module.go):
// the generated handlers register absolute spec paths, so r.Mount does not fit.
func NewRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Recoverer)
	r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		WriteProblem(w, Problem{Title: "Not Found", Status: http.StatusNotFound})
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		WriteProblem(w, Problem{Title: "Method Not Allowed", Status: http.StatusMethodNotAllowed})
	})
	return r
}
