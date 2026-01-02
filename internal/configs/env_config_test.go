package configs

import "testing"

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
