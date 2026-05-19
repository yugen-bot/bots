package config

import (
	"testing"
)

func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DISCORD_TOKEN", "tok")
	t.Setenv("DISCORD_APP_ID", "app")
}

func TestLoad_Defaults(t *testing.T) {
	setRequiredEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Env != "development" {
		t.Errorf("Env = %q, want %q", cfg.Env, "development")
	}

	if cfg.SyncCommands {
		t.Error("SyncCommands should default to false")
	}

	if cfg.APIPort != "8080" {
		t.Errorf("APIPort = %q, want %q", cfg.APIPort, "8080")
	}

	if cfg.APIHost != "0.0.0.0" {
		t.Errorf("APIHost = %q, want %q", cfg.APIHost, "0.0.0.0")
	}
}

func TestLoad_RequiredFields(t *testing.T) {
	t.Run("missing DISCORD_TOKEN", func(t *testing.T) {
		t.Setenv("DISCORD_APP_ID", "app")

		_, err := Load()
		if err == nil {
			t.Error("expected error for missing DISCORD_TOKEN")
		}
	})

	t.Run("missing DISCORD_APP_ID", func(t *testing.T) {
		t.Setenv("DISCORD_TOKEN", "tok")

		_, err := Load()
		if err == nil {
			t.Error("expected error for missing DISCORD_APP_ID")
		}
	})
}

func TestLoad_OwnerIDFromOwnerIDs(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("OWNER_IDS", "111,222,333")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.OwnerID != "111" {
		t.Errorf("OwnerID = %q, want %q", cfg.OwnerID, "111")
	}

	if len(cfg.OwnerIDs) != 3 {
		t.Errorf("OwnerIDs len = %d, want 3", len(cfg.OwnerIDs))
	}
}

func TestLoad_OwnerIDNotOverwritten(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("OWNER_ID", "explicit")
	t.Setenv("OWNER_IDS", "111,222")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.OwnerID != "explicit" {
		t.Errorf("OwnerID = %q, want %q", cfg.OwnerID, "explicit")
	}
}

func TestLoad_Values(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("ENV", "production")
	t.Setenv("SYNC_COMMANDS", "true")
	t.Setenv("API_LISTEN_PORT", "9090")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Env != "production" {
		t.Errorf("Env = %q, want %q", cfg.Env, "production")
	}

	if !cfg.SyncCommands {
		t.Error("SyncCommands should be true")
	}

	if cfg.APIPort != "9090" {
		t.Errorf("APIPort = %q, want %q", cfg.APIPort, "9090")
	}
}
