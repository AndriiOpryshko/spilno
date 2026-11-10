package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"time"
)

// person is a signed-in session: the cookie is sent by hand because it is Secure and the
// local stack speaks plain http.
type person struct {
	email  string
	cookie string
}

type reply struct {
	status int
	body   []byte
	took   time.Duration
	err    error
}

// problemCode is the `code` member of an application/problem+json body, "" if none.
func (r reply) problemCode() string {
	var p struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal(r.body, &p)
	return p.Code
}

type api struct {
	base string
	http *http.Client
}

func newAPI(base string) *api {
	return &api{
		base: base,
		http: &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns: 512, MaxIdleConnsPerHost: 512, MaxConnsPerHost: 0,
				IdleConnTimeout: 90 * time.Second,
			},
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
	}
}

func (a *api) do(p *person, method, path, contentType string, body []byte) reply {
	var rd io.Reader
	if body != nil {
		rd = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(context.Background(), method, a.base+path, rd)
	if err != nil {
		return reply{err: err}
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if p != nil {
		req.Header.Set("Cookie", p.cookie)
	}
	start := time.Now()
	resp, err := a.http.Do(req)
	if err != nil {
		return reply{err: err, took: time.Since(start)}
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	return reply{status: resp.StatusCode, body: b, took: time.Since(start), err: err}
}

func (a *api) get(p *person, path string) reply { return a.do(p, http.MethodGet, path, "", nil) }

func (a *api) postJSON(p *person, path string, v any) reply {
	b, _ := json.Marshal(v)
	return a.do(p, http.MethodPost, path, "application/json", b)
}

// signIn goes through the development shim: the code is the verified email address.
func (a *api) signIn(email string) (*person, error) {
	req, _ := http.NewRequest(http.MethodGet, a.base+"/identity/sign-in", nil)
	resp, err := a.http.Do(req)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	loc, err := url.Parse(resp.Header.Get("Location"))
	if err != nil || loc.Query().Get("state") == "" {
		return nil, fmt.Errorf("sign-in: status %d, Location %q carries no state (is SPILNO_DEV_SIGN_IN on?)", resp.StatusCode, resp.Header.Get("Location"))
	}
	cb, _ := http.NewRequest(http.MethodGet, a.base+"/identity/sign-in/callback?"+url.Values{"code": {email}, "state": {loc.Query().Get("state")}}.Encode(), nil)
	for _, c := range resp.Cookies() {
		cb.AddCookie(c)
	}
	r2, err := a.http.Do(cb)
	if err != nil {
		return nil, err
	}
	r2.Body.Close()
	for _, c := range r2.Cookies() {
		if c.Name == "spilno_session" && c.Value != "" {
			return &person{email: email, cookie: c.Name + "=" + c.Value}, nil
		}
	}
	return nil, fmt.Errorf("sign-in of %s: status %d, no spilno_session cookie", email, r2.StatusCode)
}

type filing struct {
	edrpou, name, applicant, termStart string
}

// file posts a заявка на підключення with the extract bytes.
func (a *api) file(p *person, f filing, extract []byte) reply {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range map[string]string{
		"kind": "connection", "edrpou": f.edrpou, "association_name": f.name,
		"association_address": "Test Street 1", "applicant_full_name": f.applicant,
		"position": "head", "term_start": f.termStart,
	} {
		_ = w.WriteField(k, v)
	}
	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", `form-data; name="extract"; filename="extract.pdf"`)
	h.Set("Content-Type", "application/pdf")
	fw, _ := w.CreatePart(h)
	_, _ = fw.Write(extract)
	_ = w.Close()
	return a.do(p, http.MethodPost, "/identity/applications", w.FormDataContentType(), buf.Bytes())
}

// pdf is the extract fixture: n bytes that start like a PDF.
func pdf(n int) []byte {
	b := bytes.Repeat([]byte{'x'}, n)
	copy(b, "%PDF-1.4\n")
	b[n-1] = 0xff
	return b
}
