package handlers

import (
	"net/http"
	"strconv"

	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

// 비밀번호 변경하기
func ChangePasswordHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCode := r.FormValue("code")
		newPassword := r.FormValue("password")

		if len(userCode) != 6 || len(newPassword) != 64 {
			utils.Error(w, "Failed to change your password, invalid inputs")
			return
		}

		verifyUid, err := strconv.ParseUint(r.FormValue("target"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid target, not a valid number")
			return
		}

		result := s.User.ChangePassword(uint(verifyUid), userCode, newPassword)
		if !result {
			utils.Error(w, "Unable to change your password, internal error")
			return
		}
		utils.Success(w, nil)
	}
}

// 사용자 정보 열람
func LoadUserInfoHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targetUserUid, err := strconv.ParseUint(r.FormValue("targetUserUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid targetUserUid, not a valid number")
			return
		}

		userInfo, err := s.User.GetUserInfo(uint(targetUserUid))
		if err != nil {
			utils.Error(w, "User not found")
			return
		}
		utils.Success(w, userInfo)
	}
}

// 사용자 권한 및 리포트 응답 가져오기
func LoadUserPermissionHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targetUserUid, err := strconv.ParseUint(r.FormValue("targetUserUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid target user uid, not a valid number")
			return
		}

		result := s.User.GetUserPermission(uint(targetUserUid))
		utils.Success(w, result)
	}
}

// 사용자 권한 수정하기
func ManageUserPermissionHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actionUserUid := utils.GetUserUidFromToken(r)
		if actionUserUid < 1 {
			utils.Error(w, "Unable to get an user uid from token")
			return
		}
		targetUserUid, err := strconv.ParseUint(r.FormValue("userUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid parameter(userUid), not a valid number")
			return
		}
		writePost, err := strconv.ParseBool(r.FormValue("writePost"))
		if err != nil {
			utils.Error(w, "Invalid parameter(writePost), not a valid boolean")
			return
		}
		writeComment, err := strconv.ParseBool(r.FormValue("writeComment"))
		if err != nil {
			utils.Error(w, "Invalid parameter(writeComment), not a valid boolean")
			return
		}
		sendChat, err := strconv.ParseBool(r.FormValue("sendChatMessage"))
		if err != nil {
			utils.Error(w, "Invalid parameter(sendChatMessage), not a valid boolean")
			return
		}
		sendReport, err := strconv.ParseBool(r.FormValue("sendReport"))
		if err != nil {
			utils.Error(w, "Invalid parameter(sendReport), not a valid boolean")
			return
		}
		login, err := strconv.ParseBool(r.FormValue("login"))
		if err != nil {
			utils.Error(w, "Invalid parameter(login), not a valid boolean")
			return
		}
		response := r.FormValue("response")

		param := models.UserPermissionReportResult{
			UserPermissionResult: models.UserPermissionResult{
				WritePost:       writePost,
				WriteComment:    writeComment,
				SendChatMessage: sendChat,
				SendReport:      sendReport,
			},
			Login:    login,
			UserUid:  uint(targetUserUid),
			Response: response,
		}

		s.User.ChangeUserPermission(actionUserUid, &param)
		utils.Success(w, nil)
	}
}

// 사용자 신고하기
func ReportUserHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		content := r.FormValue("content")
		userUid, err := strconv.ParseUint(r.FormValue("userUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid userUid, not a valid number")
			return
		}
		targetUserUid, err := strconv.ParseUint(r.FormValue("targetUserUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid targetUserUid, not a valid number")
			return
		}
		checkedBlackList, err := strconv.ParseUint(r.FormValue("checkedBlackList"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid checkedBlackList, not a valid number")
			return
		}
		result := s.User.ReportTargetUser(uint(userUid), uint(targetUserUid), checkedBlackList > 0, content)
		if !result {
			utils.Error(w, "You have no permission to report other user")
			return
		}
		utils.Success(w, nil)
	}
}
