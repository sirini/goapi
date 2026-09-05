package handlers

import (
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"net/http/httptest"
	"testing"
)

type publicSummaryBoard struct {
	studioHandlerBoardService
	config models.BoardConfig
}

func (s *publicSummaryBoard) GetBoardConfig(uid uint) models.BoardConfig { return s.config }

type publicSummaryUser struct {
	services.UserService
	user models.UserInfoResult
	err  error
}

func (s *publicSummaryUser) GetUserInfo(uid uint) (models.UserInfoResult, error) {
	return s.user, s.err
}

func TestPublicSummaryAccessAndIdentity(t *testing.T) {
	for _, name := range []string{"public", "private-list", "private-view", "blocked", "missing-user", "wrong-user", "missing-board", "invalid-user", "query-failure"} {
		t.Run(name, func(t *testing.T) {
			board := &publicSummaryBoard{config: models.BoardConfig{Uid: 7}}
			user := &publicSummaryUser{user: models.UserInfoResult{Uid: 42}}
			target := "/board/user/summary?id=photo&targetUserUid=42&userUid=999"
			switch name {
			case "private-list":
				board.config.Level.List = 1
			case "private-view":
				board.config.Level.View = 1
			case "blocked":
				user.user.Blocked = true
			case "missing-user":
				user.err = errors.New("missing")
			case "wrong-user":
				user.user.Uid = 43
			case "missing-board":
				target = "/board/user/summary?id=missing&targetUserUid=42"
			case "invalid-user":
				target = "/board/user/summary?id=photo&targetUserUid=0"
			case "query-failure":
				board.err = errors.New("query failed")
			}
			h := NewNuboBoardHandler(&services.Service{Board: board, User: user})
			app := fiber.New()
			app.Get("/board/user/summary", h.PublicUserSummaryHandler)
			response, err := app.Test(httptest.NewRequest("GET", target, nil))
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			var body models.ResponseCommon
			if err = json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body.Success != (name == "public") {
				t.Fatalf("unexpected response: %+v", body)
			}
			if name == "public" || name == "query-failure" {
				if board.param.UserUid != 42 || board.param.BoardUid != 7 || !board.param.PublicOnly || !board.param.SummaryOnly {
					t.Fatalf("invalid scope: %+v", board.param)
				}
			} else if board.studioCalls != 0 {
				t.Fatal("denied request reached aggregation")
			}
		})
	}
}
