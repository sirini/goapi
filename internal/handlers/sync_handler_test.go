package handlers

import (
	"testing"

	"github.com/sirini/goapi/internal/configs"
)

func TestSyncKeyMatches(t *testing.T) {
	tests := []struct {
		name       string
		provided   string
		configured string
		want       bool
	}{
		{name: "matching", provided: "dedicated-secret", configured: "dedicated-secret", want: true},
		{name: "different", provided: "wrong", configured: "dedicated-secret", want: false},
		{name: "empty provided", configured: "dedicated-secret", want: false},
		{name: "empty configuration", provided: "anything", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := syncKeyMatches(tt.provided, tt.configured); got != tt.want {
				t.Fatalf("syncKeyMatches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfiguredSyncKeyPrefersDedicatedSecretWithJWTFallback(t *testing.T) {
	original := configs.Env
	t.Cleanup(func() { configs.Env = original })

	configs.Env.JWTSecretKey = "jwt-secret"
	configs.Env.SyncSecretKey = ""
	if got := configuredSyncKey(); got != "jwt-secret" {
		t.Fatalf("fallback sync key = %q, want JWT secret", got)
	}

	configs.Env.SyncSecretKey = "sync-secret"
	if got := configuredSyncKey(); got != "sync-secret" {
		t.Fatalf("configured sync key = %q, want dedicated secret", got)
	}
}
