package inits

import (
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	"jurien.dev/yugen/shared/utils"
)

// InitPprof starts a net/http pprof server when PPROF_ENABLED=true.
// The listen address defaults to "localhost:6060" and can be overridden via PPROF_ADDR.
// Call this after the logger is initialized.
func InitPprof() {
	if os.Getenv("PPROF_ENABLED") != "true" {
		return
	}

	addr := os.Getenv("PPROF_ADDR")
	if addr == "" {
		addr = "localhost:6060"
	}

	utils.Logger.Infof("pprof: listening on %s", addr)

	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			utils.Logger.Errorf("pprof: %v", err)
		}
	}()
}
