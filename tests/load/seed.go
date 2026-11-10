package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// world is what the seed leaves behind: 20 associations and 200 applications with extracts.
type world struct {
	operator *person
	heads    []head         // the approved associations, each with its голова signed in
	pending  []*application // pending applications that carry a recorded check
	decided  []*application // an application that is already decided (refusals are expected)
	assocs   []string
	mu       sync.Mutex
}

type head struct {
	person      *person
	association string
}

type application struct {
	id        string
	filingID  string
	applicant *person
}

type seedConfig struct {
	associations, applications int
	edrpouBase                 int
	runID, operatorEmail       string
}

func decode(r reply, v any) error {
	if r.err != nil {
		return r.err
	}
	if r.status/100 != 2 {
		return fmt.Errorf("status %d: %s", r.status, strings.TrimSpace(string(r.body)))
	}
	return json.Unmarshal(r.body, v)
}

// seed files every application as its own person, records the оператор's check on the
// ones that will be decided, and approves the first cfg.associations of them.
func seed(a *api, cfg seedConfig) (*world, error) {
	w := &world{}
	var err error
	if w.operator, err = a.signIn(cfg.operatorEmail); err != nil {
		return nil, err
	}
	today := time.Now().Format("2006-01-02")
	apps := make([]*application, cfg.applications)
	errs := make(chan error, cfg.applications)
	sem := make(chan struct{}, 8)
	var wg sync.WaitGroup
	for i := range apps {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			email := fmt.Sprintf("load-%s-%03d@example.test", cfg.runID, i)
			p, err := a.signIn(email)
			if err != nil {
				errs <- err
				return
			}
			f := filing{edrpou: fmt.Sprintf("%08d", cfg.edrpouBase+i), name: fmt.Sprintf("OSBB Load %03d", i), applicant: fmt.Sprintf("Test User %03d", i), termStart: "2026-01-01"}
			var app struct {
				ID string `json:"id"`
			}
			if err := decode(a.file(p, f, pdf(64*1024)), &app); err != nil {
				errs <- fmt.Errorf("file %s: %w", email, err)
				return
			}
			var detail struct {
				Filings []struct {
					ID string `json:"id"`
				} `json:"filings"`
			}
			if err := decode(a.get(p, "/identity/applications/"+app.ID), &detail); err != nil || len(detail.Filings) == 0 {
				errs <- fmt.Errorf("read back %s: %v", app.ID, err)
				return
			}
			apps[i] = &application{id: app.ID, filingID: detail.Filings[0].ID, applicant: p}
			// the check the approval needs; the extract names the same head and code
			chk := map[string]any{"filing_id": detail.Filings[0].ID, "signature_valid": true, "signer_full_name": f.applicant, "signed_on": today,
				"head_full_name": f.applicant, "extract_edrpou": f.edrpou, "extract_dated_on": today}
			if r := a.postJSON(w.operator, "/identity/operator/applications/"+app.ID+"/checks", chk); r.err != nil || r.status != 201 {
				errs <- fmt.Errorf("check %s: status %d %v %s", app.ID, r.status, r.err, r.body)
			}
		}()
	}
	wg.Wait()
	close(errs)
	if e := <-errs; e != nil {
		return nil, fmt.Errorf("seed: %w", e)
	}
	for i, app := range apps {
		if i >= cfg.associations {
			w.pending = append(w.pending, app)
			continue
		}
		var dec struct {
			AssociationID string `json:"association_id"`
		}
		if err := decode(a.postJSON(w.operator, "/identity/operator/applications/"+app.id+"/approval", map[string]any{"filing_id": app.filingID}), &dec); err != nil || dec.AssociationID == "" {
			return nil, fmt.Errorf("seed: approve %s: %v", app.id, err)
		}
		w.heads = append(w.heads, head{person: app.applicant, association: dec.AssociationID})
		w.assocs = append(w.assocs, dec.AssociationID)
		w.decided = append(w.decided, app)
	}
	return w, nil
}

// popPending takes an undecided application, or an already decided one once none is left.
func (w *world) popPending(i int) (*application, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if n := len(w.pending); n > 0 {
		app := w.pending[n-1]
		w.pending = w.pending[:n-1]
		return app, true
	}
	return w.decided[i%len(w.decided)], false
}
