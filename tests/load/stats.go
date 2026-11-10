package main

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// sample collects the outcomes of one scenario.
type sample struct {
	mu      sync.Mutex
	name    string
	took    []time.Duration
	errs    int
	errNote map[string]int
	byOp    map[string][]time.Duration
	refused int // expected refusals, not errors
}

func newSample(name string) *sample {
	return &sample{name: name, errNote: map[string]int{}, byOp: map[string][]time.Duration{}}
}

func (s *sample) ok(op string, d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.took = append(s.took, d)
	s.byOp[op] = append(s.byOp[op], d)
}

func (s *sample) refusal(op string, d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refused++
	s.took = append(s.took, d)
	s.byOp[op] = append(s.byOp[op], d)
}

func (s *sample) fail(op, why string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errs++
	s.errNote[op+": "+why]++
}

func pct(d []time.Duration, q float64) time.Duration {
	if len(d) == 0 {
		return 0
	}
	c := append([]time.Duration(nil), d...)
	sort.Slice(c, func(i, j int) bool { return c[i] < c[j] })
	i := int(float64(len(c))*q+0.999999) - 1
	if i < 0 {
		i = 0
	}
	return c[i]
}

// verdict is the result of one scenario against its thresholds.
type verdict struct {
	name     string
	requests int
	errors   int
	rps      float64
	p50, p95 time.Duration
	max      time.Duration
	failures []string
	detail   []string
}

func (s *sample) judge(elapsed time.Duration, p95Max time.Duration, minRPS float64) verdict {
	s.mu.Lock()
	defer s.mu.Unlock()
	v := verdict{name: s.name, requests: len(s.took) + s.errs, errors: s.errs}
	if elapsed > 0 {
		v.rps = float64(len(s.took)) / elapsed.Seconds()
	}
	v.p50, v.p95 = pct(s.took, 0.5), pct(s.took, 0.95)
	v.max = pct(s.took, 1)
	if v.p95 > p95Max {
		v.failures = append(v.failures, fmt.Sprintf("p95 %s exceeds the limit %s", v.p95.Round(time.Millisecond), p95Max))
	}
	if s.errs > 0 {
		v.failures = append(v.failures, fmt.Sprintf("%d request(s) failed", s.errs))
	}
	if minRPS > 0 && v.rps < minRPS {
		v.failures = append(v.failures, fmt.Sprintf("throughput %.1f req/s is below %.1f", v.rps, minRPS))
	}
	if len(s.took) == 0 {
		v.failures = append(v.failures, "no request completed")
	}
	for k, n := range s.errNote {
		v.detail = append(v.detail, fmt.Sprintf("failed x%d  %s", n, k))
	}
	ops := make([]string, 0, len(s.byOp))
	for k := range s.byOp {
		ops = append(ops, k)
	}
	sort.Strings(ops)
	for _, k := range ops {
		d := s.byOp[k]
		v.detail = append(v.detail, fmt.Sprintf("%-28s n=%-5d p50=%-8s p95=%-8s max=%s", k, len(d), pct(d, .5).Round(time.Microsecond*100), pct(d, .95).Round(time.Microsecond*100), pct(d, 1).Round(time.Microsecond*100)))
	}
	if s.refused > 0 {
		v.detail = append(v.detail, fmt.Sprintf("expected refusals (not errors): %d", s.refused))
	}
	return v
}

func (v verdict) print() {
	status := "PASS"
	if len(v.failures) > 0 {
		status = "FAIL"
	}
	fmt.Printf("\n[%s] %s: %d requests, %d errors, %.1f req/s, p50 %s, p95 %s, max %s\n", status, v.name, v.requests, v.errors, v.rps, v.p50.Round(time.Microsecond*100), v.p95.Round(time.Microsecond*100), v.max.Round(time.Microsecond*100))
	for _, d := range v.detail {
		fmt.Println("    " + d)
	}
	for _, f := range v.failures {
		fmt.Println("    threshold missed: " + f)
	}
}
