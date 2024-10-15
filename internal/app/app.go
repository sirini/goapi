package app

import (
	"net/http"

	"github.com/sirini/goapi/internal/app/router"
)

func SetupRouter() *http.ServeMux {
	mux := http.NewServeMux()
	router.SetupRoutes(mux)

	return mux
}
