package handlers

import (
	"net/http"
	"strconv"

	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/utils"
)

// 사용자 정보 열람
func LoadUserInfoHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targetUserUidStr := r.URL.Query().Get("targetUserUid")
		if targetUserUidStr == "" {
			utils.ResponseError(w, "Missing targetUserUid parameter")
			return
		}

		targetUserUid, err := strconv.ParseUint(targetUserUidStr, 10, 32)
		if err != nil {
			utils.ResponseError(w, "Invalid targetUserUid parameter")
			return
		}

		userInfo, err := s.UserService.GetUserInfo(uint(targetUserUid))
		if err != nil {
			utils.ResponseError(w, "User not found")
			return
		}

		utils.ResponseSuccess(w, userInfo)
	}
}

// 사용자 신고하기
func ReportUserHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userUidStr := r.URL.Query().Get("userUid")
		targetUserUidStr := r.FormValue("targetUserUid")
		contentStr := r.FormValue("content")
		checkedBlackListStr := r.FormValue("checkedBlackList")

		if userUidStr == "" || targetUserUidStr == "" || contentStr == "" || checkedBlackListStr == "" {
			utils.ResponseError(w, "Invalid parameters, unable to parse form")
			return
		}

		userUid, err := strconv.ParseUint(userUidStr, 10, 32)
		if err != nil {
			utils.ResponseError(w, "Invalid userUid parameter")
			return
		}

		targetUserUid, err := strconv.ParseUint(targetUserUidStr, 10, 32)
		if err != nil {
			utils.ResponseError(w, "Invalid targetUserUid parameter")
			return
		}

		checkedBlackList, err := strconv.ParseUint(checkedBlackListStr, 10, 32)
		if err != nil {
			utils.ResponseError(w, "Invalid checkedBlackList parameter")
			return
		}

		result := s.UserService.ReportTargetUser(uint(userUid), uint(targetUserUid), checkedBlackList > 0, contentStr)
		if !result {
			utils.ResponseError(w, "You have no permission to report other user")
			return
		}

		utils.ResponseSuccess(w, nil)
	}
}

// 로그인 하기
func SigninHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.FormValue("id")
		pw := r.FormValue("password")

		if len(pw) != 64 || !utils.IsValidEmail(id) {
			utils.ResponseError(w, "Failed to sign in, invalid ID or password")
			return
		}

		user := s.UserService.Signin(id, pw)
		if user.Uid < 1 {
			utils.ResponseError(w, "Unable to get an information, invalid ID or password")
			return
		}

		utils.ResponseSuccess(w, user)
	}
}
