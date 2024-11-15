package handlers

import (
	"net/http"
	"strconv"

	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

// 오고 간 쪽지들의 목록 가져오기
func LoadChatListHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actionUserUid := utils.GetUserUidFromToken(r)
		if actionUserUid < 1 {
			utils.Error(w, "Unable to get an user uid from token")
			return
		}

		limit, err := strconv.ParseUint(r.FormValue("limit"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid limit, not a valid number")
			return
		}

		chatItems, err := s.Chat.GetChattingList(actionUserUid, uint(limit))
		if err != nil {
			utils.Error(w, err.Error())
			return
		}
		utils.Success(w, chatItems)
	}
}

// 특정인과 나눈 최근 쪽지들의 내용 가져오기
func LoadChatHistoryHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actionUserUid := utils.GetUserUidFromToken(r)
		if actionUserUid < 1 {
			utils.Error(w, "Unable to get an user uid from token")
			return
		}

		targetUserUid, err := strconv.ParseUint(r.FormValue("userUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid (target) user uid, not a valid number")
			return
		}

		limit, err := strconv.ParseUint(r.FormValue("limit"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid limit, not a valid number")
			return
		}

		chatHistories, err := s.Chat.GetChattingHistory(actionUserUid, uint(targetUserUid), uint(limit))
		if err != nil {
			utils.Error(w, err.Error())
			return
		}
		utils.Success(w, chatHistories)
	}
}

// 쪽지 내용 저장하기
func SaveChatHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actionUserUid := utils.GetUserUidFromToken(r)
		if actionUserUid < 1 {
			utils.Error(w, "Unable to get an user uid from token")
			return
		}

		message := r.FormValue("message")
		if len(message) < 2 {
			utils.Error(w, "Your message is too short, aborted")
			return
		}

		targetUserUid64, err := strconv.ParseUint(r.FormValue("targetUserUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid targetUserUid parameter")
			return
		}

		message = utils.Escape(message)
		targetUserUid := uint(targetUserUid64)

		if isPerm := s.Auth.CheckUserPermission(actionUserUid, models.USER_ACTION_SEND_CHAT); !isPerm {
			utils.Error(w, "You don't have permission to send a chat message")
			return
		}

		insertId := s.Chat.SaveChatMessage(actionUserUid, targetUserUid, message)
		if insertId < 1 {
			utils.Error(w, "Failed to send a message")
			return
		}
		utils.Success(w, insertId)
	}
}
