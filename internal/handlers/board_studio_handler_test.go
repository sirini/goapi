package handlers

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

type studioHandlerBoardService struct {
	services.BoardService
	param       models.BoardStudioParam
	result      models.BoardStudioResult
	err         error
	studioCalls int
}

func (s *studioHandlerBoardService) GetBoardUid(id string) uint {
	if id == "photo" {
		return 7
	}
	return 0
}

func (s *studioHandlerBoardService) GetStudio(param models.BoardStudioParam) (models.BoardStudioResult, error) {
	s.param = param
	s.studioCalls++
	return s.result, s.err
}

func withStudioHandlerToken(t *testing.T, uid uint) string {
	t.Helper()
	previous := configs.Env.JWTSecretKey
	configs.Env.JWTSecretKey = "studio-handler-secret"
	t.Cleanup(func() { configs.Env.JWTSecretKey = previous })
	token, err := utils.GenerateAccessToken(uid, 1)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func studioHandlerResponse(t *testing.T, handler *NuboBoardHandler, target string, token string) (int, models.ResponseCommon) {
	t.Helper()
	app := fiber.New()
	app.Get("/board/my/studio", handler.MyStudioHandler)
	request := httptest.NewRequest("GET", target, nil)
	request.Header.Set(models.AUTH_KEY, "Bearer "+token)
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	var body models.ResponseCommon
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	return response.StatusCode, body
}

func TestMyStudioHandlerUsesDefaultsAndJWTUserOnly(t *testing.T) {
	token := withStudioHandlerToken(t, 19)
	board := &studioHandlerBoardService{result: models.BoardStudioResult{
		Posts: models.BoardStudioPosts{Items: make([]models.BoardStudioPostItem, 0)},
	}}
	handler := NewNuboBoardHandler(&services.Service{Board: board})

	status, response := studioHandlerResponse(t, handler, "/board/my/studio?id=photo&userUid=999", token)
	if status != fiber.StatusOK || !response.Success || response.Code != models.CODE_SUCCESS {
		t.Fatalf("studio response status=%d body=%+v", status, response)
	}
	want := models.BoardStudioParam{
		BoardUid: 7,
		UserUid:  19,
		Page:     1,
		Limit:    20,
		Sort:     models.BOARD_STUDIO_SORT_RECENT,
	}
	if board.param != want {
		t.Fatalf("studio param = %+v, want %+v", board.param, want)
	}
}

func TestMyStudioHandlerRejectsInvalidParameters(t *testing.T) {
	token := withStudioHandlerToken(t, 19)
	for _, target := range []string{
		"/board/my/studio",
		"/board/my/studio?id=missing",
		"/board/my/studio?id=photo&page=0",
		"/board/my/studio?id=photo&page=nope",
		"/board/my/studio?id=photo&limit=0",
		"/board/my/studio?id=photo&limit=51",
		"/board/my/studio?id=photo&sort=uid%20DESC",
	} {
		t.Run(target, func(t *testing.T) {
			board := &studioHandlerBoardService{}
			handler := NewNuboBoardHandler(&services.Service{Board: board})
			status, response := studioHandlerResponse(t, handler, target, token)
			if status != fiber.StatusOK || response.Success || response.Code != models.CODE_INVALID_PARAMETER || response.Result != nil {
				t.Fatalf("invalid response status=%d body=%+v", status, response)
			}
			if board.studioCalls != 0 {
				t.Fatalf("repository-facing service called %d times", board.studioCalls)
			}
		})
	}
}

func TestMyStudioHandlerUsesFailedOperationEnvelopeForServiceErrors(t *testing.T) {
	token := withStudioHandlerToken(t, 19)
	board := &studioHandlerBoardService{err: errors.New("studio query failed")}
	handler := NewNuboBoardHandler(&services.Service{Board: board})

	status, response := studioHandlerResponse(t, handler, "/board/my/studio?id=photo", token)
	if status != fiber.StatusOK || response.Success || response.Code != models.CODE_FAILED_OPERATION || response.Error != "studio query failed" || response.Result != nil {
		t.Fatalf("repository error response status=%d body=%+v", status, response)
	}
}
