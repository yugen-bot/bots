package inits

import (
	"net/http"
	_ "net/http/pprof" // registers handlers on http.DefaultServeMux
	"os"

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

	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			utils.Logger.Errorf("pprof: %v", err)
		}
	}()
}
