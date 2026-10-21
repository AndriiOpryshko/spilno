// Package httpapi holds cross-BC HTTP plumbing: RFC 9457 problem responses
// and OpenAPI request validation wired to emit them.
package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	middleware "github.com/oapi-codegen/nethttp-middleware"
)

// FieldError points at one offending input location.
type FieldError struct {
	Pointer string `json:"pointer"`
	Detail  string `json:"detail"`
}

// Problem is the wire shape of RFC 9457 problem details. It mirrors the
// spec's Problem schema; each BC's gen package has its own copy of the type.
type Problem struct {
	Type     string       `json:"type"`
	Title    string       `json:"title"`
	Status   int          `json:"status"`
	Detail   string       `json:"detail,omitempty"`
	Instance string       `json:"instance,omitempty"`
	Errors   []FieldError `json:"errors,omitempty"`
}

// WriteProblem writes p as application/problem+json.
func WriteProblem(w http.ResponseWriter, p Problem) {
	if p.Type == "" {
		p.Type = "about:blank"
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)
	_ = json.NewEncoder(w).Encode(p)
}

func init() {
	// kin-openapi does not check format: uuid unless the format is defined.
	openapi3.DefineStringFormatValidator("uuid", openapi3.NewRegexpFormatValidator(openapi3.FormatOfStringForUUIDOfRFC9562))
}

// RequestValidator returns chi-compatible middleware validating requests
// against spec and answering failures with problem+json. Body/param schema
// violations become 422; unknown routes 404; wrong method 405.
func RequestValidator(spec *openapi3.T) func(http.Handler) http.Handler {
	spec.Servers = nil // avoid Host matching; see nethttp-middleware README
	return middleware.OapiRequestValidatorWithOptions(spec, &middleware.Options{
		Options: openapi3filter.Options{
			MultiError:         true,
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
		},
		ErrorHandlerWithOpts: validationErrorHandler,
	})
}

func validationErrorHandler(_ context.Context, err error, w http.ResponseWriter, r *http.Request, opts middleware.ErrorHandlerOpts) {
	switch {
	case opts.MatchedRoute == nil && opts.StatusCode == http.StatusNotFound:
		WriteProblem(w, Problem{Title: "Not Found", Status: http.StatusNotFound, Detail: err.Error()})
		return
	case opts.StatusCode == http.StatusMethodNotAllowed:
		WriteProblem(w, Problem{Title: "Method Not Allowed", Status: http.StatusMethodNotAllowed})
		return
	case opts.StatusCode != http.StatusBadRequest:
		WriteProblem(w, Problem{Title: http.StatusText(opts.StatusCode), Status: opts.StatusCode})
		return
	}
	WriteProblem(w, Problem{
		Title:    "Validation failed",
		Status:   http.StatusUnprocessableEntity,
		Instance: r.URL.Path,
		Errors:   fieldErrors(err),
	})
}

// fieldErrors flattens kin-openapi's error tree into JSON-pointer field errors.
// It switches on concrete types: errors.As would see through RequestError ->
// SchemaError -> MultiError and drop the parameter prefix.
func fieldErrors(err error) []FieldError {
	var out []FieldError
	var walk func(error, string)
	walk = func(e error, prefix string) {
		switch x := e.(type) {
		case openapi3.MultiError:
			for _, sub := range x {
				walk(sub, prefix)
			}
		case *openapi3filter.RequestError:
			p := prefix
			if x.Parameter != nil {
				p = "/" + x.Parameter.In + "/" + x.Parameter.Name
			}
			if x.Err == nil {
				out = append(out, FieldError{Pointer: p, Detail: x.Reason})
				return
			}
			walk(x.Err, p)
		case *openapi3.SchemaError:
			ptr := strings.Join(x.JSONPointer(), "/")
			if ptr != "" {
				ptr = "/" + ptr
			}
			out = append(out, FieldError{Pointer: prefix + ptr, Detail: x.Reason})
		default:
			out = append(out, FieldError{Pointer: prefix, Detail: e.Error()})
		}
	}
	walk(err, "")
	return out
}

// StrictRequestError is for StrictHTTPServerOptions.RequestErrorHandlerFunc
// (body decode / param binding failures that slip past the validator).
func StrictRequestError(w http.ResponseWriter, _ *http.Request, err error) {
	WriteProblem(w, Problem{Title: "Bad Request", Status: http.StatusBadRequest, Detail: err.Error()})
}

// StrictResponseError is for StrictHTTPServerOptions.ResponseErrorHandlerFunc.
func StrictResponseError(w http.ResponseWriter, _ *http.Request, _ error) {
	WriteProblem(w, Problem{Title: "Internal Server Error", Status: http.StatusInternalServerError})
}

// ParamBindError is for ChiServerOptions.ErrorHandlerFunc (generated wrapper
// failing to bind a path/query param the validator let through).
func ParamBindError(w http.ResponseWriter, _ *http.Request, err error) {
	WriteProblem(w, Problem{Title: "Validation failed", Status: http.StatusUnprocessableEntity, Detail: err.Error()})
}
