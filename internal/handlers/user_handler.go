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
		targetUserUidStr := r.FormValue("targetUserUid")
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
		userUidStr := r.FormValue("userUid")
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

// 비밀번호 변경하기
func ChangePasswordHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targetStr := r.FormValue("target")
		userCode := r.FormValue("code")
		newPassword := r.FormValue("password")

		if len(userCode) != 6 || len(newPassword) != 64 {
			utils.ResponseError(w, "Failed to change your password, invalid inputs")
			return
		}

		verifyUid, err := strconv.ParseUint(targetStr, 10, 32)
		if err != nil {
			utils.ResponseError(w, "Invalid target, not a valid number")
			return
		}

		result := s.UserService.ChangePassword(uint(verifyUid), userCode, newPassword)
		if !result {
			utils.ResponseError(w, "Unable to change your password, internal error")
			return
		}
		utils.ResponseSuccess(w, nil)
	}
}
