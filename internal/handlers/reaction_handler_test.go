package handlers

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

// 리액션 쓰기 핸들러는 reaction 필드의 누락과 명시적 null을 구분해야 한다.
type reactionBoardService struct {
	services.BoardService
	param    models.BoardReactionParam
	setCalls int
}

func (s *reactionBoardService) SetPostReaction(param models.BoardReactionParam) (models.ReactionState, error) {
	s.param = param
	s.setCalls++
	return models.ReactionState{}, nil
}

type reactionCommentService struct {
	services.CommentService
	param    models.CommentReactionParam
	setCalls int
}

func (s *reactionCommentService) SetReaction(param models.CommentReactionParam) (models.ReactionState, error) {
	s.param = param
	s.setCalls++
	return models.ReactionState{}, nil
}

func withReactionHandlerToken(t *testing.T, uid uint) string {
	t.Helper()
	previous := configs.Env.JWTSecretKey
	configs.Env.JWTSecretKey = "reaction-handler-secret"
	t.Cleanup(func() { configs.Env.JWTSecretKey = previous })
	token, err := utils.GenerateAccessToken(uid, 1)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestBoardSetReactionHandlerDistinguishesMissingAndNull(t *testing.T) {
	token := withReactionHandlerToken(t, 9)
	cases := []struct {
		name        string
		payload     string
		wantService bool
		wantCancel  bool
		wantValue   string
	}{
		{"missing reaction", `{"boardUid":1,"postUid":2}`, false, false, ""},
		{"explicit null cancels", `{"boardUid":1,"postUid":2,"reaction":null}`, true, true, ""},
		{"empty string", `{"boardUid":1,"postUid":2,"reaction":""}`, true, false, ""},
		{"unknown kind", `{"boardUid":1,"postUid":2,"reaction":"downvote"}`, true, false, "downvote"},
		{"known kind", `{"boardUid":1,"postUid":2,"reaction":"best"}`, true, false, "best"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			board := &reactionBoardService{}
			handler := &NuboBoardHandler{service: &services.Service{Board: board}}
			app := fiber.New()
			app.Patch("/board/reaction", handler.SetReactionHandler)
			request := httptest.NewRequest("PATCH", "/board/reaction", strings.NewReader(tc.payload))
			request.Header.Set(models.AUTH_KEY, "Bearer "+token)
			request.Header.Set("Content-Type", "application/json")
			response, err := app.Test(request)
			if err != nil {
				t.Fatal(err)
			}
			if tc.wantService {
				if response.StatusCode != fiber.StatusOK {
					t.Fatalf("status = %d, want 200", response.StatusCode)
				}
				if board.setCalls != 1 {
					t.Fatalf("service called %d times, want 1", board.setCalls)
				}
				if board.param.BoardUid != 1 || board.param.PostUid != 2 || board.param.UserUid != 9 {
					t.Fatalf("unexpected param: %+v", board.param)
				}
				if board.param.ReactionIsNull != tc.wantCancel || board.param.Reaction != tc.wantValue {
					t.Fatalf("reaction mapping = %+v, want cancel=%v value=%q", board.param, tc.wantCancel, tc.wantValue)
				}
			} else {
				if response.StatusCode != fiber.StatusOK {
					t.Fatalf("error responses must use the 200 envelope, got %d", response.StatusCode)
				}
				var body models.ResponseCommon
				if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if body.Success || body.Code != models.CODE_INVALID_PARAMETER || !strings.Contains(body.Error, "reaction field is required") {
					t.Fatalf("missing reaction was not rejected: %+v", body)
				}
				if board.setCalls != 0 {
					t.Fatal("missing reaction must not reach the service")
				}
			}
		})
	}

	t.Run("zero uid is rejected", func(t *testing.T) {
		board := &reactionBoardService{}
		handler := &NuboBoardHandler{service: &services.Service{Board: board}}
		app := fiber.New()
		app.Patch("/board/reaction", handler.SetReactionHandler)
		request := httptest.NewRequest("PATCH", "/board/reaction", strings.NewReader(`{"boardUid":0,"postUid":2,"reaction":"like"}`))
		request.Header.Set(models.AUTH_KEY, "Bearer "+token)
		request.Header.Set("Content-Type", "application/json")
		response, err := app.Test(request)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), "boardUid and postUid are required") {
			t.Fatalf("zero uid response: %s", body)
		}
		if board.setCalls != 0 {
			t.Fatal("zero uid must not reach the service")
		}
	})
}

func TestCommentSetReactionHandlerDistinguishesMissingAndNull(t *testing.T) {
	token := withReactionHandlerToken(t, 9)
	cases := []struct {
		name        string
		payload     string
		wantService bool
		wantCancel  bool
	}{
		{"missing reaction", `{"boardUid":1,"commentUid":3}`, false, false},
		{"explicit null cancels", `{"boardUid":1,"commentUid":3,"reaction":null}`, true, true},
		{"empty string", `{"boardUid":1,"commentUid":3,"reaction":""}`, true, false},
		{"known kind", `{"boardUid":1,"commentUid":3,"reaction":"hmm"}`, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			comments := &reactionCommentService{}
			handler := NewNuboCommentHandler(&services.Service{Comment: comments})
			app := fiber.New()
			app.Patch("/comment/reaction", handler.SetReactionHandler)
			request := httptest.NewRequest("PATCH", "/comment/reaction", strings.NewReader(tc.payload))
			request.Header.Set(models.AUTH_KEY, "Bearer "+token)
			request.Header.Set("Content-Type", "application/json")
			response, err := app.Test(request)
			if err != nil {
				t.Fatal(err)
			}
			if !tc.wantService {
				var body models.ResponseCommon
				if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if body.Success || !strings.Contains(body.Error, "reaction field is required") {
					t.Fatalf("missing reaction was not rejected: %+v", body)
				}
				if comments.setCalls != 0 {
					t.Fatal("missing reaction must not reach the service")
				}
				return
			}
			if response.StatusCode != fiber.StatusOK || comments.setCalls != 1 {
				t.Fatalf("status=%d serviceCalls=%d", response.StatusCode, comments.setCalls)
			}
			if comments.param.ReactionIsNull != tc.wantCancel {
				t.Fatalf("cancel mapping = %+v, want cancel=%v", comments.param, tc.wantCancel)
			}
			if comments.param.CommentUid != 3 || comments.param.BoardUid != 1 || comments.param.UserUid != 9 {
				t.Fatalf("unexpected param: %+v", comments.param)
			}
		})
	}
}
