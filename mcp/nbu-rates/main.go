// Command nbu-rates is an MCP server with the official exchange rates of the
// National Bank of Ukraine. It speaks stdio by default and Streamable HTTP
// with -http.
package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/AndriiOpryshko/spilno/mcp/nbu-rates/internal/nbu"
	"github.com/AndriiOpryshko/spilno/mcp/nbu-rates/internal/server"
)

func main() {
	httpAddr := flag.String("http", os.Getenv("NBU_RATES_HTTP_ADDR"), "serve Streamable HTTP on this address (e.g. 127.0.0.1:3336) instead of stdio")
	flag.Parse()

	// stdout belongs to the protocol on stdio; logs go to stderr.
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := server.New(logged{nbu.NewClient(os.Getenv("NBU_BASE_URL"))})
	var err error
	if *httpAddr == "" {
		err = srv.Run(ctx, &mcp.StdioTransport{})
	} else {
		err = serveHTTP(ctx, *httpAddr, srv)
	}
	if err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("nbu-rates stopped", "err", err)
		os.Exit(1)
	}
}

func serveHTTP(ctx context.Context, addr string, srv *mcp.Server) error {
	mux := http.NewServeMux()
	// Stateless: every tool is a pure read, so no session has to survive between requests.
	mux.Handle("/mcp", mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return srv },
		&mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true}))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok\n")) })

	hs := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = hs.Shutdown(shutdown)
	}()
	slog.InfoContext(ctx, "nbu-rates listening", "addr", addr, "endpoint", "/mcp")
	if err := hs.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// logged keeps the upstream failure detail in the server log; callers of the
// tool only see a stable error code.
type logged struct{ src *nbu.Client }

func (l logged) Rate(ctx context.Context, currency string, date time.Time) (nbu.Rate, error) {
	r, err := l.src.Rate(ctx, currency, date)
	if errors.Is(err, nbu.ErrUnavailable) {
		slog.WarnContext(ctx, "nbu request failed", "currency", currency, "date", date.Format(time.DateOnly), "err", err)
	}
	return r, err
}
