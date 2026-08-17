package handlers

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

type statusPinger struct{ err error }

func (p statusPinger) PingContext(context.Context) error { return p.err }

func TestStatusEndpoints(t *testing.T) {
	app := fiber.New()
	ready := newNuboStatusHandler(statusPinger{}, "1.2.1")
	unavailable := newNuboStatusHandler(statusPinger{err: errors.New("private database address")}, "1.2.1")
	app.Get("/health", ready.HealthHandler)
	app.Get("/ready", ready.ReadyHandler)
	app.Get("/unavailable", unavailable.ReadyHandler)
	app.Get("/version", ready.VersionHandler)

	tests := []struct {
		path       string
		wantStatus int
		wantBody   string
	}{
		{path: "/health", wantStatus: fiber.StatusOK, wantBody: `"status":"ok"`},
		{path: "/ready", wantStatus: fiber.StatusOK, wantBody: `"status":"ok"`},
		{path: "/unavailable", wantStatus: fiber.StatusServiceUnavailable, wantBody: `"status":"unavailable"`},
		{path: "/version", wantStatus: fiber.StatusOK, wantBody: `"apiContract":"1"`},
	}
	for _, tt := range tests {
		resp, err := app.Test(httptest.NewRequest("GET", tt.path, nil))
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != tt.wantStatus || !strings.Contains(string(body), tt.wantBody) {
			t.Fatalf("GET %s = %d %s", tt.path, resp.StatusCode, body)
		}
		if strings.Contains(string(body), "private database address") {
			t.Fatal("readiness response leaked an internal database error")
		}
	}
}
