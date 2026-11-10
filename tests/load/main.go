// Command load runs the three load scenarios of PRD §6 against a running core API and exits
// non-zero when a threshold is missed. Defaults are the PRD §6 numbers and the local stack;
// every parameter is a flag with an environment fallback (LOAD_<NAME>).
//
//	mise run load                                  # the PRD numbers, 5 minutes of reads
//	mise run load -- -duration 20s -read-p95 1ms   # short run, impossible threshold: must fail
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

func envOr(name, def string) string {
	if v := os.Getenv("LOAD_" + name); v != "" {
		return v
	}
	return def
}

func envDur(name string, def time.Duration) time.Duration {
	d, err := time.ParseDuration(envOr(name, def.String()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "LOAD_%s: %v\n", name, err)
		os.Exit(2)
	}
	return d
}

func envInt(name string, def int) int {
	n, err := strconv.Atoi(envOr(name, strconv.Itoa(def)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "LOAD_%s: %v\n", name, err)
		os.Exit(2)
	}
	return n
}

func envFloat(name string, def float64) float64 {
	n, err := strconv.ParseFloat(envOr(name, strconv.FormatFloat(def, 'f', -1, 64)), 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "LOAD_%s: %v\n", name, err)
		os.Exit(2)
	}
	return n
}

func main() {
	base := flag.String("base-url", envOr("BASE_URL", "http://localhost:8080"), "core API address (the stack must run with SPILNO_DEV_SIGN_IN=true)")
	operator := flag.String("operator-email", envOr("OPERATOR_EMAIL", "operator@example.test"), "the оператор's address (SPILNO_OPERATOR_EMAIL of the stack)")
	duration := flag.Duration("duration", envDur("DURATION", 5*time.Minute), "length of the read and write scenarios")
	readRate := flag.Float64("read-rps", envFloat("READ_RPS", 50), "read mix request rate; also the throughput that must be reached")
	writeRate := flag.Float64("write-rps", envFloat("WRITE_RPS", 5), "write mix request rate")
	readP95 := flag.Duration("read-p95", envDur("READ_P95", 200*time.Millisecond), "read p95 limit")
	writeP95 := flag.Duration("write-p95", envDur("WRITE_P95", 500*time.Millisecond), "write p95 limit")
	filingP95 := flag.Duration("filing-p95", envDur("FILING_P95", 5*time.Second), "filing with 5 MB extract p95 limit")
	nFilings := flag.Int("filings", envInt("FILINGS", 20), "number of 5 MB filings")
	filingConc := flag.Int("filing-concurrency", envInt("FILING_CONCURRENCY", 1), "filings in flight at once")
	assocs := flag.Int("associations", envInt("ASSOCIATIONS", 20), "associations to seed")
	apps := flag.Int("applications", envInt("APPLICATIONS", 200), "applications to seed (including the ones that become associations)")
	runID := flag.String("run-id", envOr("RUN_ID", strconv.FormatInt(time.Now().Unix()%1000000, 10)), "makes the seeded people unique per run")
	edrpouBase := flag.Int("edrpou-base", envInt("EDRPOU_BASE", 0), "first 8-digit code to seed (0 = derived from the run id; must be unused in the database)")
	only := flag.String("only", envOr("ONLY", ""), "run one scenario: reads, writes or filings (default all)")
	flag.Parse()

	if *edrpouBase == 0 {
		n, _ := strconv.Atoi(*runID)
		*edrpouBase = 10_000_000 + (n%3000)*20_000
	}
	cfg := seedConfig{associations: *assocs, applications: *apps, edrpouBase: *edrpouBase, runID: *runID, operatorEmail: *operator}
	a := newAPI(*base)

	fmt.Printf("load: %s, run %s, %s of reads at %.0f req/s and writes at %.0f req/s, %d filings\n", *base, *runID, *duration, *readRate, *writeRate, *nFilings)
	if r := a.get(nil, "/system/health"); r.err != nil || r.status != 200 {
		fmt.Fprintf(os.Stderr, "load: the API at %s is not healthy: status %d %v\n", *base, r.status, r.err)
		os.Exit(2)
	}
	start := time.Now()
	w, err := seed(a, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load: %v\n", err)
		os.Exit(2)
	}
	fmt.Printf("seeded %d associations and %d applications with extracts in %s\n", len(w.heads), *apps, time.Since(start).Round(time.Millisecond))

	var verdicts []verdict
	run := func(name string) bool { return *only == "" || *only == name }
	if run("reads") {
		s, el := readMix(a, w, *readRate, *duration)
		// the throughput bar is the rate itself, with 2% for the tail of the last second
		verdicts = append(verdicts, s.judge(el, *readP95, *readRate*0.98))
	}
	if run("writes") {
		s, el := writeMix(a, w, *writeRate, *duration)
		verdicts = append(verdicts, s.judge(el, *writeP95, 0))
	}
	if run("filings") {
		s, el := filings(a, cfg, *nFilings, *filingConc, cfg.edrpouBase+*apps+10, 5<<20)
		verdicts = append(verdicts, s.judge(el, *filingP95, 0))
	}
	failed := false
	for _, v := range verdicts {
		v.print()
		failed = failed || len(v.failures) > 0
	}
	if failed {
		fmt.Println("\nload: FAILED, a threshold was missed")
		os.Exit(1)
	}
	fmt.Println("\nload: all thresholds met")
}
