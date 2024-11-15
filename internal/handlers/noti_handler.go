package handlers

import (
	"net/http"
	"strconv"

	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/utils"
)

// 알림 모두 확인하기 처리
func CheckedAllNoti(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userUid := utils.GetUserUidFromToken(r)
		if userUid < 1 {
			utils.Error(w, "Unable to get an user uid from token")
			return
		}

		s.Noti.CheckedAllNoti(userUid, 10)
		utils.Success(w, nil)
	}
}

// 알림 목록 가져오기
func LoadNotiListHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userUid := utils.GetUserUidFromToken(r)
		if userUid < 1 {
			utils.Error(w, "Unable to get an user uid from token")
			return
		}

		limit, err := strconv.ParseUint(r.FormValue("limit"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid limit, not a valid number")
			return
		}

		notis, err := s.Noti.GetUserNoti(userUid, uint(limit))
		if err != nil {
			utils.Error(w, "Failed to load your notifications")
			return
		}
		utils.Success(w, notis)
	}
}
