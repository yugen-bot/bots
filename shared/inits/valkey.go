package inits

import (
	"fmt"

	"github.com/valkey-io/valkey-go"

	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/utils"
)

// InitValkey creates a Valkey client from cfg.ValkeyURL.
// Returns nil, nil when ValkeyURL is not configured — callers must handle nil.
func InitValkey(cfg *config.Config) (valkey.Client, error) {
	if cfg.ValkeyURL == "" {
		return nil, nil
	}

	utils.Logger.Info("Connecting to Valkey")

	opts, err := valkey.ParseURL(cfg.ValkeyURL)
	if err != nil {
		return nil, fmt.Errorf("valkey: parse URL: %w", err)
	}

	client, err := valkey.NewClient(opts)
	if err != nil {
		return nil, fmt.Errorf("valkey: new client: %w", err)
	}

	return client, nil
}
