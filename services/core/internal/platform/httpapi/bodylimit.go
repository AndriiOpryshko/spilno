package httpapi

import (
	"mime"
	"net/http"
	"strings"
)

// BodyLimit refuses a request body over the limit of its kind with a 413 problem+json before any
// later middleware (the request validator buffers multipart bodies) reads it. multipart applies to
// multipart/form-data requests, other to everything else.
func BodyLimit(multipart, other int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			limit := other
			if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err == nil && strings.EqualFold(mt, "multipart/form-data") {
				limit = multipart
			}
			if r.ContentLength > limit {
				w.Header().Set("Connection", "close")
				WriteProblem(w, Problem{Title: "Request Entity Too Large", Status: http.StatusRequestEntityTooLarge})
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}
