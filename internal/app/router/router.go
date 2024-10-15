package router

import (
	"net/http"

	"github.com/sirini/goapi/internal/handler"
)

func SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/goapi/hello", handler.Hello)
}
