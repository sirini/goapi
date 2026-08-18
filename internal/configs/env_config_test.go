package configs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// GONOSUMDB=* GOPROXY=off go test ./internal/configs

func TestEnvironmentFilePathDefaultsAndAcceptsOverride(t *testing.T) {
	t.Setenv(EnvironmentFileVariable, "")
	if path := EnvironmentFilePath(); path != ".env" {
		t.Fatalf("default environment path = %q, want .env", path)
	}

	externalPath := filepath.Join(t.TempDir(), "nubo.env")
	t.Setenv(EnvironmentFileVariable, "  "+externalPath+"  ")
	if path := EnvironmentFilePath(); path != externalPath {
		t.Fatalf("environment path = %q, want %q", path, externalPath)
	}
}

func TestLoadConfigReadsExternalFileAndPreservesProcessPrecedence(t *testing.T) {
	original := Env
	t.Cleanup(func() { Env = original })

	environmentPath := filepath.Join(t.TempDir(), "nubo.env")
	contents := "GOAPI_TITLE=File Title\nGOAPI_HOST=127.0.0.1\nGOAPI_PORT=4310\nDB_NAME=external_db\nNUBO_UPLOAD_DIR=/srv/nubo/upload\nADMIN_ID=admin@example.com\nADMIN_PW=admin-password\n"
	if err := os.WriteFile(environmentPath, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(EnvironmentFileVariable, environmentPath)
	t.Setenv("GOAPI_TITLE", "Process Title")
	unsetEnvironmentForTest(t, "GOAPI_PORT")
	unsetEnvironmentForTest(t, "GOAPI_HOST")
	unsetEnvironmentForTest(t, "DB_NAME")
	unsetEnvironmentForTest(t, "NUBO_UPLOAD_DIR")
	unsetEnvironmentForTest(t, "ADMIN_ID")
	unsetEnvironmentForTest(t, "ADMIN_PW")

	if err := LoadConfig(); err != nil {
		t.Fatal(err)
	}
	if Env.Title != "Process Title" {
		t.Fatalf("title = %q, want process override", Env.Title)
	}
	if Env.GoPort != "4310" {
		t.Fatalf("port = %q, want file value", Env.GoPort)
	}
	if Env.GoHost != "127.0.0.1" {
		t.Fatalf("host = %q, want loopback file value", Env.GoHost)
	}
	if Env.DBName != "external_db" {
		t.Fatalf("database = %q, want file value", Env.DBName)
	}
	if Env.UploadDir != "/srv/nubo/upload" {
		t.Fatalf("upload directory = %q, want file value", Env.UploadDir)
	}
	if Env.AdminID != "admin@example.com" || Env.AdminPW != "admin-password" {
		t.Fatalf("admin values were not loaded from the external environment file")
	}
}

func TestLoadConfigMissingFileDoesNotReplaceCurrentConfig(t *testing.T) {
	original := Env
	t.Cleanup(func() { Env = original })
	Env = Config{Title: "keep-current"}
	t.Setenv(EnvironmentFileVariable, filepath.Join(t.TempDir(), "missing.env"))

	if err := LoadConfig(); err == nil {
		t.Fatal("LoadConfig succeeded with a missing explicit environment file")
	}
	if Env.Title != "keep-current" {
		t.Fatalf("failed load replaced current config: %+v", Env)
	}
}

func TestLoadConfigKeepsLegacyListenDefault(t *testing.T) {
	original := Env
	t.Cleanup(func() { Env = original })

	environmentPath := filepath.Join(t.TempDir(), "nubo.env")
	if err := os.WriteFile(environmentPath, []byte("GOAPI_PORT=3006\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(EnvironmentFileVariable, environmentPath)
	unsetEnvironmentForTest(t, "GOAPI_HOST")

	if err := LoadConfig(); err != nil {
		t.Fatal(err)
	}
	if Env.GoHost != "0.0.0.0" {
		t.Fatalf("legacy host = %q, want 0.0.0.0", Env.GoHost)
	}
}

func unsetEnvironmentForTest(t *testing.T, key string) {
	t.Helper()
	value, exists := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if exists {
			_ = os.Setenv(key, value)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}

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
