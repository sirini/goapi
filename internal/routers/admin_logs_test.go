package routers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/handlers"
	"net/http/httptest"
	"testing"
)

type removedDeletionAdmin struct{ handlers.AdminHandler }

func TestAdminRecordDeletionRoutesAreNotRegistered(t *testing.T) {
	app := fiber.New()
	RegisterAdminRouters(app, &handlers.Handler{Admin: &removedDeletionAdmin{}, CanAuthenticate: func(uint) bool { return true }})
	for _, route := range app.GetRoutes() {
		if route.Method == "DELETE" && route.Path == "/admin/system/records/:kind" {
			t.Fatal("record deletion is still registered")
		}
	}
	for _, kind := range []string{"visits", "audit"} {
		response, err := app.Test(httptest.NewRequest("DELETE", "/admin/system/records/"+kind+"?date=2020-01-01", nil))
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		// No matching API exists; authorization middleware may also reject an anonymous request.
		if response.StatusCode != 404 && response.StatusCode != 401 {
			t.Fatalf("removed API status %d", response.StatusCode)
		}
	}
}
