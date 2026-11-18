package main

import (
	"strings"
	"testing"
	"time"
)

func filled(n int, each time.Duration, errs int) *sample {
	s := newSample("reads")
	for i := 0; i < n; i++ {
		s.ok("GET /x", each)
	}
	for i := 0; i < errs; i++ {
		s.fail("GET /x", "status 500")
	}
	return s
}

func hasFailure(v verdict, sub string) bool {
	for _, f := range v.failures {
		if strings.Contains(f, sub) {
			return true
		}
	}
	return false
}

// CI mode: a p95 over the limit is reported next to its limit and does not fail the run.
func TestJudgeReportOnlyDoesNotFailOnP95(t *testing.T) {
	v := filled(100, 300*time.Millisecond, 0).judgeMode(time.Second, 200*time.Millisecond, 50, true)
	if len(v.failures) != 0 {
		t.Fatalf("report-only: failures = %v, want none for a p95 of 300ms over a 200ms limit", v.failures)
	}
	if v.p95 != 300*time.Millisecond || v.p95Limit != 200*time.Millisecond {
		t.Fatalf("report-only: p95 %s limit %s, want 300ms next to 200ms", v.p95, v.p95Limit)
	}
}

// Throughput and errors fail the job in both modes.
func TestJudgeReportOnlyStillFailsOnThroughputAndErrors(t *testing.T) {
	slow := filled(10, time.Millisecond, 0).judgeMode(time.Second, 200*time.Millisecond, 49, true)
	if !hasFailure(slow, "throughput") {
		t.Fatalf("report-only: failures = %v, want a throughput failure at 10 req/s vs 49", slow.failures)
	}
	bad := filled(100, time.Millisecond, 1).judgeMode(time.Second, 200*time.Millisecond, 50, true)
	if !hasFailure(bad, "failed") {
		t.Fatalf("report-only: failures = %v, want a failed-request failure", bad.failures)
	}
}

// Release mode (no flag): a missed p95 still fails.
func TestJudgeEnforcingFailsOnP95(t *testing.T) {
	v := filled(100, 300*time.Millisecond, 0).judgeMode(time.Second, 200*time.Millisecond, 50, false)
	if !hasFailure(v, "p95") {
		t.Fatalf("enforcing: failures = %v, want a p95 failure", v.failures)
	}
	if v.p95Limit != 200*time.Millisecond {
		t.Fatalf("enforcing: p95Limit = %s, want 200ms", v.p95Limit)
	}
}
