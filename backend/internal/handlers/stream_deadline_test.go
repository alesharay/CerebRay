package handlers_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/aray/cerebray/backend/internal/metrics"
	mw "github.com/aray/cerebray/backend/internal/middleware"
)

// The server sets an absolute WriteTimeout, which would otherwise cut off a long
// SSE stream partway through (a promoted note can generate for over a minute).
// SendMessage clears its own write deadline via http.ResponseController, which
// only works if every ResponseWriter wrapper in the chain exposes Unwrap.
// These tests pin both halves of that: the chain stays unwrappable, and a stream
// that outlives WriteTimeout still completes.

const (
	testWriteTimeout = 300 * time.Millisecond
	testStreamTime   = 900 * time.Millisecond
	testChunks       = 6
)

// streamHandler writes testChunks SSE events spread over testStreamTime,
// clearing the write deadline first when clearDeadline is set.
func streamHandler(t *testing.T, clearDeadline bool, deadlineErr chan<- error) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Error("ResponseWriter does not implement http.Flusher")
			return
		}

		if clearDeadline {
			err := http.NewResponseController(w).SetWriteDeadline(time.Time{})
			deadlineErr <- err
			if err != nil {
				return
			}
		}

		for i := range testChunks {
			fmt.Fprintf(w, "data: {\"delta\":\"chunk-%d\"}\n\n", i)
			flusher.Flush()
			time.Sleep(testStreamTime / testChunks)
		}
		fmt.Fprint(w, "data: {\"done\":true}\n\n")
		flusher.Flush()
	}
}

// metrics.New registers against the default Prometheus registry, so it can only
// run once per process.
var testMetrics = sync.OnceValue(metrics.New)

// newStreamServer builds a server with the same middleware chain and absolute
// WriteTimeout as cmd/server, so the wrappers under test are the real ones.
func newStreamServer(t *testing.T, h http.Handler) *httptest.Server {
	t.Helper()

	r := chi.NewRouter()
	r.Use(metrics.HTTPMiddleware(testMetrics()))
	r.Use(mw.Logger)
	r.Method(http.MethodGet, "/stream", h)

	srv := httptest.NewUnstartedServer(r)
	srv.Config.WriteTimeout = testWriteTimeout
	srv.Start()
	t.Cleanup(srv.Close)
	return srv
}

func TestSSEStreamSurvivesWriteTimeout(t *testing.T) {
	deadlineErr := make(chan error, 1)
	srv := newStreamServer(t, streamHandler(t, true, deadlineErr))

	resp, err := srv.Client().Get(srv.URL + "/stream")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("stream was cut off mid-read after %d bytes: %v", len(body), err)
	}

	if err := <-deadlineErr; err != nil {
		t.Fatalf("SetWriteDeadline failed, so a ResponseWriter in the chain is missing Unwrap: %v", err)
	}

	for i := range testChunks {
		want := fmt.Sprintf("chunk-%d", i)
		if !contains(string(body), want) {
			t.Errorf("missing %q in stream body: %q", want, body)
		}
	}
	if !contains(string(body), `"done":true`) {
		t.Errorf("stream did not reach the done event: %q", body)
	}
}

// TestSSEStreamCutWithoutDeadlineClear is the control: it demonstrates the
// original bug, so the test above is not passing for some unrelated reason.
func TestSSEStreamCutWithoutDeadlineClear(t *testing.T) {
	srv := newStreamServer(t, streamHandler(t, false, nil))

	resp, err := srv.Client().Get(srv.URL + "/stream")
	if err != nil {
		// Some stacks surface the truncation on the request itself.
		return
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr == nil && contains(string(body), `"done":true`) {
		t.Fatalf("expected the stream to be cut at WriteTimeout, but it completed: %q", body)
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
