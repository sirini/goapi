package configs

import (
	"strings"
	"testing"
)

// GONOSUMDB=* GOPROXY=off go test ./internal/configs

func TestGetJWTAccessRefreshValid(t *testing.T) {
	original := Env
	defer func() { Env = original }()

	Env.JWTAccessHours = "5"
	Env.JWTRefreshDays = "10"

	access, refresh := GetJWTAccessRefresh()

	if access != 5 {
		t.Fatalf("expected access hours to be 5, got %d", access)
	}

	if refresh != 10 {
		t.Fatalf("expected refresh days to be 10, got %d", refresh)
	}
}

func TestGetJWTAccessRefreshInvalid(t *testing.T) {
	original := Env
	defer func() { Env = original }()

	Env.JWTAccessHours = "invalid"
	Env.JWTRefreshDays = "invalid"

	access, refresh := GetJWTAccessRefresh()

	if access != 2 {
		t.Fatalf("expected default access hours 2, got %d", access)
	}

	if refresh != 30 {
		t.Fatalf("expected default refresh days 30, got %d", refresh)
	}
}

func TestGetFileSizeLimitValid(t *testing.T) {
	original := Env
	defer func() { Env = original }()

	Env.FileSizeLimit = "2048"

	size := GetFileSizeLimit()

	if size != 2048 {
		t.Fatalf("expected parsed file size limit 2048, got %d", size)
	}
}

func TestGetFileSizeLimitInvalid(t *testing.T) {
	original := Env
	defer func() { Env = original }()

	Env.FileSizeLimit = "not-a-number"

	size := GetFileSizeLimit()

	if size != 10485760 {
		t.Fatalf("expected default file size limit 10485760, got %d", size)
	}
}

func TestGetImageDescriptionConfigRequiresKeyAndExplicitOptIn(t *testing.T) {
	original := Env
	defer func() { Env = original }()

	Env.OpenaiKey = "configured"
	Env.ImageDescription.Enabled = "false"
	if config := GetImageDescriptionConfig(); config.Enabled {
		t.Fatal("API key alone enabled image descriptions")
	}

	Env.ImageDescription.Enabled = "true"
	Env.OpenaiKey = ""
	if config := GetImageDescriptionConfig(); config.Enabled {
		t.Fatal("image descriptions were enabled without an API key")
	}

	Env.OpenaiKey = "configured"
	if config := GetImageDescriptionConfig(); !config.Enabled {
		t.Fatal("explicitly configured image descriptions were not enabled")
	}
}

func TestGetImageDescriptionConfigUsesSafeDefaultsForInvalidValues(t *testing.T) {
	original := Env
	defer func() { Env = original }()

	Env.OpenaiKey = "configured"
	Env.ImageDescription.Enabled = "not-a-boolean"
	Env.ImageDescription.Model = " "
	Env.ImageDescription.MaxPerPost = "unlimited"
	Env.ImageDescription.Concurrency = "0"

	config := GetImageDescriptionConfig()
	if config.Enabled {
		t.Fatal("invalid opt-in value enabled image descriptions")
	}
	if config.Model != "gpt-5.6-luna" {
		t.Fatalf("model = %q, want gpt-5.6-luna", config.Model)
	}
	if config.MaxPerPost != 3 {
		t.Fatalf("max per post = %d, want 3", config.MaxPerPost)
	}
	if config.MaxConcurrent != 1 {
		t.Fatalf("max concurrent = %d, want 1", config.MaxConcurrent)
	}
}

func TestGetImageDescriptionConfigAcceptsBoundedOverrides(t *testing.T) {
	original := Env
	defer func() { Env = original }()

	Env.OpenaiKey = "configured"
	Env.ImageDescription.Enabled = "true"
	Env.ImageDescription.Model = "gpt-4o-mini-2024-07-18"
	Env.ImageDescription.MaxPerPost = "0"
	Env.ImageDescription.Concurrency = "4"

	config := GetImageDescriptionConfig()
	if !config.Enabled {
		t.Fatal("valid image-description configuration was not enabled")
	}
	if config.Model != "gpt-4o-mini-2024-07-18" {
		t.Fatalf("model = %q", config.Model)
	}
	if config.MaxPerPost != 0 {
		t.Fatalf("max per post = %d, want 0", config.MaxPerPost)
	}
	if config.MaxConcurrent != 4 {
		t.Fatalf("max concurrent = %d, want 4", config.MaxConcurrent)
	}
}

func TestNotificationSenderForeignKeyReferencesUser(t *testing.T) {
	ddl := notificationSenderForeignKeyDDL("nubo_")
	if !strings.Contains(ddl, "REFERENCES nubo_user(uid)") {
		t.Fatalf("notification sender foreign key references the wrong table: %s", ddl)
	}
}
