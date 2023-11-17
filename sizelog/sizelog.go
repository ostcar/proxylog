package sizelog

import (
	"context"
	_ "embed" // For embedding
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
	"time"
)

//go:embed page.html
var page []byte

// SizeLog logs sizes
type SizeLog struct {
	size atomic.Uint64
}

// LogSize is called with a size.
func (sl *SizeLog) LogSize(size int) {
	sl.size.Add(uint64(size))
}

// Run starts the log
func (sl *SizeLog) Run(ctx context.Context, addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/data", sl.serveData)
	mux.HandleFunc("/", sl.serveHTML)

	httpSRV := &http.Server{
		Addr:        addr,
		Handler:     mux,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}

	wait := make(chan error)
	go func() {
		// Wait for the context to be closed.
		<-ctx.Done()

		if err := httpSRV.Shutdown(context.WithoutCancel(ctx)); err != nil {
			wait <- fmt.Errorf("HTTP server shutdown: %w", err)
			return
		}
		wait <- nil
	}()

	fmt.Printf("Listen webserver on: %s\n", addr)
	if err := httpSRV.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("HTTP Server failed: %v", err)
	}

	return <-wait
}

func (sl *SizeLog) serveHTML(w http.ResponseWriter, r *http.Request) {
	w.Write(page)
}

func (sl *SizeLog) serveData(w http.ResponseWriter, r *http.Request) {
	// TODO: This only supports one client. If more should be supported, the resetting of sl.size has to be done in a different goroutine
	w.Header().Add("Content-Type", "text/event-stream")
	w.Header().Add("Content-Disposition", "inline")

	ctx := r.Context()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			size := sl.size.Load()
			if size != 0 {
				sl.size.Add(-size)
			}

			fmt.Fprintf(w, "data: %d\n\n", size)
			w.(http.Flusher).Flush()

		case <-ctx.Done():
			return
		}
	}
}
