package routers

import (
	"net/http"

	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/middlewares"
	"github.com/sirini/goapi/internal/services"
)

// 일반 게시판 라우터들 등록하기
func SetupBoardRouter(mux *http.ServeMux, s *services.Service) {
	mux.HandleFunc("GET /goapi/board/list", handlers.BoardListHandler(s))
	mux.HandleFunc("GET /goapi/board/view", handlers.BoardViewHandler(s))
	mux.HandleFunc("GET /goapi/board/photo/list", handlers.GalleryListHandler(s))
	mux.HandleFunc("GET /goapi/board/photo/view", handlers.GalleryLoadPhotoHandler(s))
}

// 로그인이 필요한 게시판 라우터들 등록하기
func SetupLoggedInBoardRouter(mux *http.ServeMux, s *services.Service) {
	mux.Handle("GET /goapi/board/download", middlewares.AuthMiddleware(handlers.DownloadHandler(s)))
	mux.Handle("GET /goapi/board/config", middlewares.AuthMiddleware(handlers.GetEditorConfigHandler(s)))
	mux.Handle("GET /goapi/board/move/list", middlewares.AuthMiddleware(handlers.ListForMoveHandler(s)))
	mux.Handle("PATCH /goapi/board/like/post", middlewares.AuthMiddleware(handlers.LikePostHandler(s)))
	mux.Handle("GET /goapi/board/load/images", middlewares.AuthMiddleware(handlers.LoadInsertImageHandler(s)))
	mux.Handle("PUT /goapi/board/move/apply", middlewares.AuthMiddleware(handlers.MovePostHandler(s)))
	mux.Handle("DELETE /goapi/board/remove/post", middlewares.AuthMiddleware(handlers.RemovePostHandler(s)))
	mux.Handle("DELETE /goapi/board/remove/image", middlewares.AuthMiddleware(handlers.RemoveInsertImageHandler(s)))
	mux.Handle("GET /goapi/board/tag/suggestion", middlewares.AuthMiddleware(handlers.SuggestionHashtagHandler(s)))
	mux.Handle("POST /goapi/board/upload/images", middlewares.AuthMiddleware(handlers.UploadInsertImageHandler(s)))
	mux.Handle("POST /goapi/board/write", middlewares.AuthMiddleware(handlers.WritePostHandler(s)))
}
