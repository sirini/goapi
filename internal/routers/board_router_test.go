package routers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/handlers"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type studioRouterBoardService struct {
	services.BoardService
	param models.BoardStudioParam
}

func (s *studioRouterBoardService) GetBoardUid(id string) uint {
	if id == "photo" {
		return 7
	}
	return 0
}

func (s *studioRouterBoardService) GetStudio(param models.BoardStudioParam) (models.BoardStudioResult, error) {
	s.param = param
	return models.BoardStudioResult{Posts: models.BoardStudioPosts{Items: make([]models.BoardStudioPostItem, 0)}}, nil
}

func TestMyStudioRouteRequiresJWTAndUsesItsUID(t *testing.T) {
	previous := configs.Env.JWTSecretKey
	configs.Env.JWTSecretKey = "studio-router-secret"
	t.Cleanup(func() { configs.Env.JWTSecretKey = previous })

	board := &studioRouterBoardService{}
	boardHandler := handlers.NewNuboBoardHandler(&services.Service{Board: board})
	app := fiber.New()
	RegisterBoardRouters(app, &handlers.Handler{
		CanAuthenticate: func(uid uint) bool { return uid == 23 },
		Board:           boardHandler,
	})

	for _, authorization := range []string{"", "Bearer invalid.jwt.token"} {
		request := httptest.NewRequest("GET", "/board/my/studio?id=photo", nil)
		if authorization != "" {
			request.Header.Set(models.AUTH_KEY, authorization)
		}
		response, err := app.Test(request)
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != fiber.StatusUnauthorized {
			t.Fatalf("authorization %q status = %d, want 401", authorization, response.StatusCode)
		}
	}

	token, err := utils.GenerateAccessToken(23, 1)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest("GET", "/board/my/studio?id=photo&page=2&limit=10&sort=comments&userUid=99", nil)
	request.Header.Set(models.AUTH_KEY, "Bearer "+token)
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("authorized studio status = %d", response.StatusCode)
	}
	want := models.BoardStudioParam{
		BoardUid: 7,
		UserUid:  23,
		Page:     2,
		Limit:    10,
		Sort:     models.BOARD_STUDIO_SORT_COMMENTS,
	}
	if board.param != want {
		t.Fatalf("route studio param = %+v, want %+v", board.param, want)
	}
}
