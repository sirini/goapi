package configs

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestMakeEnvIncludesCurrentMailAndInstallerSettings(t *testing.T) {
	previousWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previousWorkingDirectory) })

	temporaryDirectory := t.TempDir()
	sample := `DB_HOST=#dbhost#
DB_USER=#dbuser#
DB_PASS=#dbpass#
DB_NAME=#dbname#
DB_TABLE_PREFIX=#dbprefix#
DB_PORT=#dbport#
DB_UNIX_SOCKET=#dbsock#
DB_MAX_IDLE=#dbmaxidle#
DB_MAX_OPEN=#dbmaxopen#
JWT_SECRET_KEY=#jwtsecret#
SYNC_SECRET_KEY=#syncsecret#
ADMIN_ID=#adminid#
ADMIN_PW=#adminpw#
RESEND_API_KEY=
RESEND_FROM_EMAIL=
RESEND_FROM_NAME=
RESEND_REPLY_TO_EMAIL=
SIGNUP_MODE=verified_email
`
	if err := os.WriteFile(filepath.Join(temporaryDirectory, "env.sample"), []byte(sample), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(temporaryDirectory); err != nil {
		t.Fatal(err)
	}
	environmentDirectory := filepath.Join(temporaryDirectory, "config")
	if err := os.Mkdir(environmentDirectory, 0700); err != nil {
		t.Fatal(err)
	}
	environmentPath := filepath.Join(environmentDirectory, "nubo.env")
	t.Setenv(EnvironmentFileVariable, environmentPath)

	dbInfo := DBInfo{
		Host: "db.example.com", User: "nubo", Pass: "secret", Name: "nubo",
		Port: "3307", Prefix: "nubo_", MaxIdle: "10", MaxOpen: "20",
	}
	if !makeEnv(dbInfo, AdminInfo{Id: "admin@example.com", Pw: "password"}) {
		t.Fatal("makeEnv returned false")
	}

	contents, err := os.ReadFile(environmentPath)
	if err != nil {
		t.Fatal(err)
	}
	result := string(contents)
	for _, expected := range []string{
		"DB_PORT=3307",
		"RESEND_API_KEY=",
		"RESEND_FROM_EMAIL=",
		"RESEND_FROM_NAME=",
		"RESEND_REPLY_TO_EMAIL=",
		"SIGNUP_MODE=verified_email",
	} {
		if !strings.Contains(result, expected) {
			t.Errorf("generated .env does not contain %q", expected)
		}
	}
	if strings.Contains(result, "GMAIL_") {
		t.Error("generated .env must not contain legacy Gmail settings")
	}
	if strings.Contains(result, "#syncsecret#") || strings.Contains(result, "#jwtsecret#") {
		t.Error("generated .env contains an unresolved secret placeholder")
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(environmentPath)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0600 {
			t.Errorf("generated .env permissions = %o, want 600", info.Mode().Perm())
		}
	}
}

func TestInstallStateUsesExplicitEnvironmentFile(t *testing.T) {
	environmentPath := filepath.Join(t.TempDir(), "nubo.env")
	t.Setenv(EnvironmentFileVariable, environmentPath)
	if isAlreadyInstalled() {
		t.Fatal("missing explicit environment file was treated as installed")
	}
	if err := os.WriteFile(environmentPath, []byte("GOAPI_PORT=3006\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if !isAlreadyInstalled() {
		t.Fatal("existing explicit environment file was not treated as installed")
	}
}
