package gdrive

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
)

const (
	// driveToken is the only access token the fake Drive accepts.
	driveToken = "at-1"
	// driveFolder is the association's folder the adapters under test upload to.
	driveFolder = "folder-1"

	driveUploadPath = "/upload/drive/v3/files"
	driveFilesPath  = "/drive/v3/files/"
)

// driveRequest is one request the fake Drive received.
type driveRequest struct {
	Method      string
	Path        string
	Query       url.Values
	ContentType string
	Body        []byte
}

// driveRestriction is one item of a file's contentRestrictions.
type driveRestriction struct {
	ReadOnly bool   `json:"readOnly"`
	Reason   string `json:"reason,omitempty"`
}

// driveFile is a file as the fake Drive holds it.
type driveFile struct {
	ID                  string
	HeadRevisionID      string
	Name                string
	MimeType            string
	Parents             []string
	SHA256Checksum      string // what Drive reports, wrong under reportWrongSHA256
	Size                int64
	Trashed             bool
	ContentRestrictions []driveRestriction
	Data                []byte
}

// driveSession is an upload session opened by the POST and consumed by the PUT.
type driveSession struct {
	name     string
	mimeType string
	parents  []string
	fields   string
}

// driveAnswer is a response computed under the lock and written after it.
type driveAnswer struct {
	status      int
	location    string
	contentType string
	body        []byte
}

// driveFake behaves like Google Drive for the five calls of the adapter:
// a resumable upload in one piece, files.get (metadata and alt=media),
// files.update and files.delete. Like Drive it answers only kind, id, name
// and mimeType unless fields asks for more.
type driveFake struct {
	srv       *httptest.Server
	closeOnce sync.Once

	mu          sync.Mutex
	seq         int
	sessions    map[string]driveSession
	files       map[string]*driveFile
	requests    []driveRequest
	wrongSHA256 bool
	status      int
}

// newDriveFake starts an empty fake Drive that is closed with the test.
func newDriveFake(t *testing.T) *driveFake {
	t.Helper()
	d := &driveFake{
		sessions: map[string]driveSession{},
		files:    map[string]*driveFile{},
	}
	d.srv = httptest.NewServer(http.HandlerFunc(d.serve))
	t.Cleanup(d.close)
	return d
}

// bearerTransport authorizes every request the way an association's token
// source does in production.
type bearerTransport struct {
	token string
	base  http.RoundTripper
}

func (b bearerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", "Bearer "+b.token)
	return b.base.RoundTrip(r)
}

// client is the only HTTP client the fake answers: it adds the access token.
func (d *driveFake) client() *http.Client {
	return &http.Client{Transport: bearerTransport{token: driveToken, base: d.srv.Client().Transport}}
}

// config is the adapter configuration of the association that owns driveFolder.
func (d *driveFake) config() Config {
	return Config{
		HTTP:      d.client(),
		APIURL:    d.srv.URL + "/drive/v3",
		UploadURL: d.srv.URL + "/upload/drive/v3",
		FolderID:  driveFolder,
	}
}

// trash moves a file to the trash, as a person in the Drive UI would; it
// reports whether the file exists.
func (d *driveFake) trash(id string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	f, ok := d.files[id]
	if ok {
		f.Trashed = true
	}
	return ok
}

// len is the number of files Drive holds, trashed ones included.
func (d *driveFake) len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.files)
}

// file returns a copy of the file Drive holds under id.
func (d *driveFake) file(id string) (driveFile, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	f, ok := d.files[id]
	if !ok {
		return driveFile{}, false
	}
	c := *f
	c.Parents = append([]string(nil), f.Parents...)
	c.ContentRestrictions = append([]driveRestriction(nil), f.ContentRestrictions...)
	c.Data = append([]byte(nil), f.Data...)
	return c, true
}

// reportWrongSHA256 switches the fault: while on, an uploaded file is stored
// with a sha256Checksum that differs from the bytes received.
func (d *driveFake) reportWrongSHA256(on bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.wrongSHA256 = on
}

// answerEverything makes Drive answer status to every request and change
// nothing; 0 switches the fault off.
func (d *driveFake) answerEverything(status int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.status = status
}

// close shuts the server down; what it holds stays readable through file and len.
func (d *driveFake) close() {
	d.closeOnce.Do(d.srv.Close)
}

// recorded returns every request received so far, oldest first.
func (d *driveFake) recorded() []driveRequest {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]driveRequest(nil), d.requests...)
}

func (d *driveFake) serve(w http.ResponseWriter, r *http.Request) {
	body, readErr := io.ReadAll(r.Body)
	a := d.answer(r, body, readErr)
	if a.location != "" {
		w.Header().Set("Location", a.location)
	}
	if a.contentType != "" {
		w.Header().Set("Content-Type", a.contentType)
	}
	w.WriteHeader(a.status)
	_, _ = w.Write(a.body)
}

func (d *driveFake) answer(r *http.Request, body []byte, readErr error) driveAnswer {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.requests = append(d.requests, driveRequest{
		Method:      r.Method,
		Path:        r.URL.Path,
		Query:       r.URL.Query(),
		ContentType: r.Header.Get("Content-Type"),
		Body:        body,
	})
	switch {
	case d.status != 0:
		return driveError(d.status, "the fake Drive answers this to everything")
	case r.Header.Get("Authorization") != "Bearer "+driveToken:
		return driveError(http.StatusUnauthorized, "invalid credentials")
	case readErr != nil:
		return driveError(http.StatusBadRequest, "the request body was cut: "+readErr.Error())
	case r.URL.Path == driveUploadPath:
		return d.answerUpload(r, body)
	case strings.HasPrefix(r.URL.Path, driveFilesPath):
		return d.answerFile(r, body)
	default:
		return driveError(http.StatusNotFound, "no such path: "+r.URL.Path)
	}
}

// answerUpload is the resumable upload: the POST opens a session, the PUT
// stores the bytes as the file of that session.
func (d *driveFake) answerUpload(r *http.Request, body []byte) driveAnswer {
	q := r.URL.Query()
	switch r.Method {
	case http.MethodPost:
		if q.Get("uploadType") != "resumable" {
			return driveError(http.StatusBadRequest, "uploadType must be resumable")
		}
		var meta struct {
			Name     string   `json:"name"`
			MimeType string   `json:"mimeType"`
			Parents  []string `json:"parents"`
		}
		if err := json.Unmarshal(body, &meta); err != nil {
			return driveError(http.StatusBadRequest, "the metadata is not JSON: "+err.Error())
		}
		d.seq++
		n := strconv.Itoa(d.seq)
		d.sessions[n] = driveSession{
			name:     meta.Name,
			mimeType: meta.MimeType,
			parents:  meta.Parents,
			fields:   q.Get("fields"),
		}
		return driveAnswer{
			status:   http.StatusOK,
			location: "http://" + r.Host + driveUploadPath + "?uploadType=resumable&upload_id=" + n,
		}
	case http.MethodPut:
		n := q.Get("upload_id")
		s, ok := d.sessions[n]
		if !ok {
			return driveError(http.StatusNotFound, "no such upload session: "+n)
		}
		delete(d.sessions, n)
		reported := body
		if d.wrongSHA256 {
			reported = append([]byte("not the bytes received"), body...)
		}
		sum := sha256.Sum256(reported)
		f := &driveFile{
			ID:             "file-" + n,
			HeadRevisionID: "rev-" + n,
			Name:           s.name,
			MimeType:       s.mimeType,
			Parents:        s.parents,
			SHA256Checksum: hex.EncodeToString(sum[:]),
			Size:           int64(len(body)),
			Data:           body,
		}
		d.files[f.ID] = f
		return f.answer(s.fields)
	default:
		return driveError(http.StatusMethodNotAllowed, r.Method+" is not an upload call")
	}
}

// answerFile is files.get, files.update and files.delete on one file.
func (d *driveFake) answerFile(r *http.Request, body []byte) driveAnswer {
	id := strings.TrimPrefix(r.URL.Path, driveFilesPath)
	f, ok := d.files[id]
	if !ok {
		return driveError(http.StatusNotFound, "file not found: "+id)
	}
	q := r.URL.Query()
	switch r.Method {
	case http.MethodGet:
		if q.Get("alt") == "media" {
			return driveAnswer{status: http.StatusOK, contentType: f.MimeType, body: f.Data}
		}
		return f.answer(q.Get("fields"))
	case http.MethodPatch:
		var patch struct {
			Trashed             *bool               `json:"trashed"`
			ContentRestrictions *[]driveRestriction `json:"contentRestrictions"`
		}
		if err := json.Unmarshal(body, &patch); err != nil {
			return driveError(http.StatusBadRequest, "the patch is not JSON: "+err.Error())
		}
		if patch.Trashed != nil {
			f.Trashed = *patch.Trashed
		}
		if patch.ContentRestrictions != nil {
			f.ContentRestrictions = *patch.ContentRestrictions
		}
		return f.answer(q.Get("fields"))
	case http.MethodDelete:
		delete(d.files, id)
		return driveAnswer{status: http.StatusNoContent}
	default:
		return driveError(http.StatusMethodNotAllowed, r.Method+" is not a files call")
	}
}

// answer is the file resource cut down to kind, id, name, mimeType and the
// top-level names fields lists.
func (f *driveFile) answer(fields string) driveAnswer {
	all := map[string]any{
		"kind":           "drive#file",
		"id":             f.ID,
		"name":           f.Name,
		"mimeType":       f.MimeType,
		"parents":        f.Parents,
		"headRevisionId": f.HeadRevisionID,
		"sha256Checksum": f.SHA256Checksum,
		"size":           strconv.FormatInt(f.Size, 10), // Drive sends int64 as a decimal string
		"trashed":        f.Trashed,
	}
	if len(f.ContentRestrictions) > 0 { // Drive omits the list of a file with no restriction
		all["contentRestrictions"] = f.ContentRestrictions
	}
	out := map[string]any{}
	for _, name := range append([]string{"kind", "id", "name", "mimeType"}, fieldNames(fields)...) {
		if name == "*" {
			out = all
			break
		}
		if v, ok := all[name]; ok {
			out[name] = v
		}
	}
	b, err := json.Marshal(out)
	if err != nil {
		return driveError(http.StatusInternalServerError, err.Error())
	}
	return driveAnswer{status: http.StatusOK, contentType: "application/json; charset=UTF-8", body: b}
}

// fieldNames returns the top-level names of a Drive fields selector:
// "id,size,contentRestrictions(readOnly)" gives id, size, contentRestrictions.
func fieldNames(fields string) []string {
	var names []string
	add := func(part string) {
		name, _, _ := strings.Cut(part, "(")
		name, _, _ = strings.Cut(name, "/")
		if name = strings.TrimSpace(name); name != "" {
			names = append(names, name)
		}
	}
	depth, start := 0, 0
	for i, c := range fields {
		switch c {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				add(fields[start:i])
				start = i + 1
			}
		}
	}
	add(fields[start:])
	return names
}

// driveError is an error answer in Drive's JSON shape.
func driveError(status int, message string) driveAnswer {
	b, _ := json.Marshal(map[string]any{
		"error": map[string]any{"code": status, "message": message},
	})
	return driveAnswer{status: status, contentType: "application/json; charset=UTF-8", body: b}
}
