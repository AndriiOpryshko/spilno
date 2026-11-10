package main

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// openLoop issues fn at a fixed rate for d, whatever the latency of earlier calls, and
// waits for all of them. It returns the wall time from the first call to the last answer.
func openLoop(rate float64, d time.Duration, fn func(i int)) time.Duration {
	start := time.Now()
	interval := time.Duration(float64(time.Second) / rate)
	var wg sync.WaitGroup
	for i := 0; ; i++ {
		at := start.Add(time.Duration(i) * interval)
		if at.Sub(start) >= d {
			break
		}
		time.Sleep(time.Until(at))
		wg.Add(1)
		go func() {
			defer wg.Done()
			fn(i)
		}()
	}
	wg.Wait()
	return time.Since(start)
}

// record classifies one reply: ok statuses count as served, refusal codes as expected
// refusals, anything else (transport error, 5xx, unexpected 4xx) as a failure.
func record(s *sample, op string, r reply, okStatus int, refusal map[int]string) {
	switch {
	case r.err != nil:
		s.fail(op, "transport: "+r.err.Error())
	case r.status == okStatus:
		s.ok(op, r.took)
	case refusal[r.status] != "" && r.problemCode() == refusal[r.status]:
		s.refusal(op, r.took)
	default:
		s.fail(op, fmt.Sprintf("status %d %s", r.status, r.problemCode()))
	}
}

// readMix: the applicant's list, the memberships, the оператор's queue and association
// list, and the association card — the read path of PRD §6.
func readMix(a *api, w *world, rate float64, d time.Duration) (*sample, time.Duration) {
	s := newSample("reads (application list, queue, operators' associations, memberships, card)")
	elapsed := openLoop(rate, d, func(i int) {
		h := w.heads[rand.IntN(len(w.heads))]
		switch i % 5 {
		case 0:
			record(s, "GET my applications", a.get(h.person, "/identity/applications"), http.StatusOK, nil)
		case 1:
			record(s, "GET operator queue", a.get(w.operator, "/identity/operator/applications?state=pending"), http.StatusOK, nil)
		case 2:
			record(s, "GET operator associations", a.get(w.operator, "/identity/operator/associations"), http.StatusOK, nil)
		case 3:
			record(s, "GET memberships", a.get(h.person, "/identity/memberships"), http.StatusOK, nil)
		default:
			record(s, "GET association card", a.get(h.person, "/identity/associations/"+h.association), http.StatusOK, nil)
		}
	})
	return s, elapsed
}

// writeMix: approvals, rejections, disablings and enablings. A decision on an already
// decided заявка and a repeated state change are refused by design and not errors.
func writeMix(a *api, w *world, rate float64, d time.Duration) (*sample, time.Duration) {
	s := newSample("writes (approve, reject, disable, enable)")
	notPending := map[int]string{http.StatusConflict: "identity.application_not_pending"}
	unchanged := map[int]string{http.StatusConflict: "identity.association_state_unchanged"}
	var disabled sync.Map
	elapsed := openLoop(rate, d, func(i int) {
		switch i % 4 {
		case 0:
			app, _ := w.popPending(i)
			record(s, "POST approval", a.postJSON(w.operator, "/identity/operator/applications/"+app.id+"/approval", map[string]any{"filing_id": app.filingID}), http.StatusOK, notPending)
		case 1:
			app, _ := w.popPending(i)
			record(s, "POST rejection", a.postJSON(w.operator, "/identity/operator/applications/"+app.id+"/rejection", map[string]any{"filing_id": app.filingID, "reason": "load test: the extract names another head"}), http.StatusOK, notPending)
		case 2:
			id := w.assocs[rand.IntN(len(w.assocs))]
			disabled.Store(id, true)
			record(s, "POST disabling", a.postJSON(w.operator, "/identity/operator/associations/"+id+"/disabling", map[string]any{"reason": "load test"}), http.StatusOK, unchanged)
		default:
			id := w.assocs[rand.IntN(len(w.assocs))]
			record(s, "POST enabling", a.do(w.operator, http.MethodPost, "/identity/operator/associations/"+id+"/enabling", "", nil), http.StatusOK, unchanged)
		}
	})
	return s, elapsed
}

// filings: n заявки with the 5 MB extract, each by a person of its own.
func filings(a *api, cfg seedConfig, n, concurrency, edrpouBase int, size int) (*sample, time.Duration) {
	s := newSample(fmt.Sprintf("filings with a %d MB extract", size>>20))
	extract := pdf(size)
	var next atomic.Int64
	var wg sync.WaitGroup
	start := time.Now()
	for range concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				i := int(next.Add(1)) - 1
				if i >= n {
					return
				}
				p, err := a.signIn(fmt.Sprintf("load-%s-filer-%03d@example.test", cfg.runID, i))
				if err != nil {
					s.fail("sign-in", err.Error())
					continue
				}
				f := filing{edrpou: fmt.Sprintf("%08d", edrpouBase+i), name: fmt.Sprintf("OSBB Filing %03d", i), applicant: fmt.Sprintf("Filing User %03d", i), termStart: "2026-01-01"}
				record(s, "POST filing (5 MB)", a.file(p, f, extract), http.StatusCreated, nil)
			}
		}()
	}
	wg.Wait()
	return s, time.Since(start)
}
