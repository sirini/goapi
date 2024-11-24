package routers

import (
	"net/http"

	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/services"
)

// 홈화면에서 사용하는 라우터들 등록하기
func SetupHomeRouter(mux *http.ServeMux, s *services.Service) {
	mux.HandleFunc("GET /goapi/home/tsboard", handlers.ShowVersionHandler)
	mux.HandleFunc("GET /goapi/home/visit", handlers.CountingVisitorHandler(s))
	mux.HandleFunc("GET /goapi/home/latest", handlers.LoadAllPostsHandler(s))
	mux.HandleFunc("GET /goapi/home/latest/post", handlers.LoadPostsByIdHandler(s))
	mux.HandleFunc("GET /goapi/home/sidebar/links", handlers.LoadSidebarLinkHandler(s))
}

// 검색 엔진 최적화용 라우터들 등록하기
func SetupSeoRouter(mux *http.ServeMux, s *services.Service) {
	mux.HandleFunc("GET /goapi/seo/sitemap.xml", handlers.LoadSitemapHandler(s))
}
