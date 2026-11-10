// Package gdrive is ports.Files on Google Drive: an association's files live
// in its own Drive folder (ADR-0005).
package gdrive

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/ports"
)

// ReadOnlyReason is the reason Drive shows on a file spilno made read-only.
const ReadOnlyReason = "locked by spilno"

// Drive answers only kind, id, name and mimeType unless fields asks for more.
const (
	uploadFields = "id,headRevisionId,sha256Checksum"
	statFields   = "id,headRevisionId,sha256Checksum,size,mimeType,trashed,contentRestrictions"
)

const jsonContentType = "application/json; charset=UTF-8"

// Config is one association's Drive storage.
type Config struct {
	HTTP      *http.Client // authorized as the association's Google account
	APIURL    string       // "https://www.googleapis.com/drive/v3" in production; no trailing slash
	UploadURL string       // "https://www.googleapis.com/upload/drive/v3" in production; no trailing slash
	FolderID  string       // the association's spilno folder; every upload goes there
}

// Files is ports.Files on Google Drive.
type Files struct {
	cfg Config
}

var _ ports.Files = (*Files)(nil)

// New returns the storage of one association.
func New(cfg Config) *Files {
	return &Files{cfg: cfg}
}

// fileResource is the part of a Drive file the adapter reads.
type fileResource struct {
	ID                  string `json:"id"`
	HeadRevisionID      string `json:"headRevisionId"`
	SHA256Checksum      string `json:"sha256Checksum"`
	Size                int64  `json:"size,string"` // Drive sends int64 as a decimal string
	MimeType            string `json:"mimeType"`
	Trashed             bool   `json:"trashed"`
	ContentRestrictions []struct {
		ReadOnly bool `json:"readOnly"`
	} `json:"contentRestrictions"`
}

// Upload streams body into the association's folder.
func (f *Files) Upload(ctx context.Context, name, mediaType string, body io.Reader) (ports.StoredFile, error) {
	session, err := f.openSession(ctx, name, mediaType)
	if err != nil {
		return ports.StoredFile{}, err
	}

	src := &hashingReader{r: body, hash: sha256.New()}
	put, err := f.send(ctx, http.MethodPut, session, mediaType, src)
	sum, size, readErr := src.result()
	if readErr != nil {
		if err == nil {
			discard(put)
		}
		return ports.StoredFile{}, fmt.Errorf("gdrive: read the upload body: %w", readErr)
	}
	if err != nil {
		return ports.StoredFile{}, err
	}
	file, err := readFile(ctx, put)
	if err != nil {
		return ports.StoredFile{}, err
	}
	if !strings.EqualFold(file.SHA256Checksum, sum) {
		if err := f.remove(ctx, file.ID); err != nil {
			return ports.StoredFile{}, fmt.Errorf("%w; deleting %s failed: %v", ports.ErrChecksumMismatch, file.ID, err)
		}
		return ports.StoredFile{}, ports.ErrChecksumMismatch
	}
	return ports.StoredFile{
		ID:        file.ID,
		Revision:  file.HeadRevisionID,
		SHA256:    sum,
		Size:      size,
		MediaType: mediaType,
	}, nil
}

// openSession starts a resumable upload of a file with that name and media
// type in the association's folder and returns the URL its bytes are PUT to.
func (f *Files) openSession(ctx context.Context, name, mediaType string) (string, error) {
	meta, err := json.Marshal(map[string]any{
		"name":     name,
		"mimeType": mediaType,
		"parents":  []string{f.cfg.FolderID},
	})
	if err != nil {
		return "", fmt.Errorf("gdrive: encode the upload metadata: %w", err)
	}
	sessionURL := f.cfg.UploadURL + "/files?uploadType=resumable&fields=" + url.QueryEscape(uploadFields)
	opened, err := f.send(ctx, http.MethodPost, sessionURL, jsonContentType, bytes.NewReader(meta))
	if err != nil {
		return "", err
	}
	discard(opened)
	session := opened.Header.Get("Location")
	if session == "" {
		return "", answerError(opened, "no Location of the upload session")
	}
	return session, nil
}

// Open streams back the stored bytes; the caller closes the reader.
func (f *Files) Open(ctx context.Context, id string) (io.ReadCloser, error) {
	resp, err := f.send(ctx, http.MethodGet, f.fileURL(id)+"?alt=media", "", nil)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// Stat reports the file as Drive sees it now.
func (f *Files) Stat(ctx context.Context, id string) (ports.FileStatus, error) {
	resp, err := f.send(ctx, http.MethodGet, f.fileURL(id)+"?fields="+url.QueryEscape(statFields), "", nil)
	if err != nil {
		return ports.FileStatus{}, err
	}
	file, err := readFile(ctx, resp)
	if err != nil {
		return ports.FileStatus{}, err
	}
	st := ports.FileStatus{
		StoredFile: ports.StoredFile{
			ID:        file.ID,
			Revision:  file.HeadRevisionID,
			SHA256:    file.SHA256Checksum,
			Size:      file.Size,
			MediaType: file.MimeType,
		},
		Trashed: file.Trashed,
	}
	for _, r := range file.ContentRestrictions {
		if r.ReadOnly {
			st.ReadOnly = true
		}
	}
	return st, nil
}

// SetReadOnly forbids changing the file's content.
func (f *Files) SetReadOnly(ctx context.Context, id string) error {
	return f.patch(ctx, id, map[string]any{
		"contentRestrictions": []map[string]any{{"readOnly": true, "reason": ReadOnlyReason}},
	})
}

// Untrash restores a trashed file.
func (f *Files) Untrash(ctx context.Context, id string) error {
	return f.patch(ctx, id, map[string]any{"trashed": false})
}

// patch is files.update with change as the body; the answer is not read.
func (f *Files) patch(ctx context.Context, id string, change map[string]any) error {
	body, err := json.Marshal(change)
	if err != nil {
		return fmt.Errorf("gdrive: encode the patch: %w", err)
	}
	resp, err := f.send(ctx, http.MethodPatch, f.fileURL(id), jsonContentType, bytes.NewReader(body))
	if err != nil {
		return err
	}
	discard(resp)
	return nil
}

// remove is files.delete.
func (f *Files) remove(ctx context.Context, id string) error {
	resp, err := f.send(ctx, http.MethodDelete, f.fileURL(id), "", nil)
	if err != nil {
		return err
	}
	discard(resp)
	return nil
}

func (f *Files) fileURL(id string) string {
	return f.cfg.APIURL + "/files/" + url.PathEscape(id)
}

// send does one request and returns its answer when the status is 2xx; the
// caller closes the body. Any other outcome is the error the port names.
func (f *Files) send(ctx context.Context, method, rawURL, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		return nil, &domain.DomainError{Code: domain.CodeStorageUnavailable}
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := f.cfg.HTTP.Do(req)
	if err != nil {
		return nil, unreachable(ctx, req)
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp, nil
	}
	discard(resp)
	return nil, statusError(resp)
}

// statusError is the error the port names for an answer outside 2xx: the file
// is not there, Drive cannot be used right now, or Drive said something the
// adapter does not expect.
func statusError(resp *http.Response) error {
	switch status := resp.StatusCode; {
	case status == http.StatusNotFound:
		return ports.ErrFileNotFound
	case status == http.StatusUnauthorized,
		status == http.StatusForbidden,
		status == http.StatusRequestTimeout,
		status == http.StatusTooManyRequests,
		status >= 500:
		return &domain.DomainError{Code: domain.CodeStorageUnavailable}
	default:
		return answerError(resp, "unexpected status")
	}
}

// readFile decodes the file resource of a 2xx answer and closes it.
func readFile(ctx context.Context, resp *http.Response) (fileResource, error) {
	raw, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return fileResource{}, unreachable(ctx, resp.Request)
	}
	var file fileResource
	if err := json.Unmarshal(raw, &file); err != nil {
		return fileResource{}, answerError(resp, "the answer is not a file: "+err.Error())
	}
	if file.ID == "" || file.HeadRevisionID == "" || file.SHA256Checksum == "" {
		return fileResource{}, answerError(resp, "the answer lacks id, headRevisionId or sha256Checksum")
	}
	return file, nil
}

// unreachable is the error of a request that could not be sent or whose
// answer could not be read: the context's error when the caller gave up,
// storage_unavailable otherwise.
func unreachable(ctx context.Context, req *http.Request) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("gdrive: %s %s: %w", req.Method, req.URL.Path, err)
	}
	return &domain.DomainError{Code: domain.CodeStorageUnavailable}
}

// answerError is the plain error of an answer the adapter cannot use.
func answerError(resp *http.Response, what string) error {
	return fmt.Errorf("gdrive: %s %s: %s: %s", resp.Request.Method, resp.Request.URL.Path, resp.Status, what)
}

// discard drains what is left of an answer and closes it, so the connection
// is reused.
func discard(resp *http.Response) {
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
	_ = resp.Body.Close()
}

// hashingReader hashes and counts what is read through it and keeps the
// source's read error. The HTTP transport may still read the body on its own
// goroutine after the answer arrived, hence the lock.
type hashingReader struct {
	r io.Reader

	mu   sync.Mutex
	hash hash.Hash
	size int64
	err  error
}

func (h *hashingReader) Read(p []byte) (int, error) {
	n, err := h.r.Read(p)
	h.mu.Lock()
	defer h.mu.Unlock()
	h.hash.Write(p[:n])
	h.size += int64(n)
	if err != nil && err != io.EOF {
		h.err = err
	}
	return n, err
}

// result is the lowercase hex sha256 and the size of the bytes read so far,
// and the source's read error if it had one.
func (h *hashingReader) result() (sum string, size int64, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return hex.EncodeToString(h.hash.Sum(nil)), h.size, h.err
}
