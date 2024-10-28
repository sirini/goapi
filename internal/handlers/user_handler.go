package handlers

import (
	"net/http"
	"strconv"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
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

// 회원가입 하기
func SignupHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.FormValue("email")
		pw := r.FormValue("password")
		name := r.FormValue("name")

		if len(pw) != 64 || !utils.IsValidEmail(id) {
			utils.ResponseError(w, "Failed to sign up, invalid ID or password")
			return
		}

		result, err := s.UserService.Signup(id, pw, name, r)
		if err != nil {
			utils.ResponseError(w, err.Error())
			return
		}
		utils.ResponseSuccess(w, result)
	}
}

// (회원가입 시) 이메일 주소가 이미 등록되어 있는지 확인하기
func CheckEmailHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.FormValue("email")

		if !utils.IsValidEmail(id) {
			utils.ResponseError(w, "Invalid email address")
			return
		}

		result := s.UserService.CheckEmailExists(id)
		if result {
			utils.ResponseError(w, "Email address is already in use")
			return
		}
		utils.ResponseSuccess(w, nil)
	}
}

// (회원가입 시) 이름이 이미 등록되어 있는지 확인하기
func CheckNameHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")

		if len(name) < 2 {
			utils.ResponseError(w, "Invalid name, too short")
			return
		}

		result := s.UserService.CheckNameExists(name)
		if result {
			utils.ResponseError(w, "Name is already in use")
			return
		}
		utils.ResponseSuccess(w, nil)
	}
}

// 인증 완료하기
func VerifyCodeHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targetStr := r.FormValue("target")
		code := r.FormValue("code")
		id := r.FormValue("email")
		pw := r.FormValue("password")
		name := r.FormValue("name")

		if len(pw) != 64 || !utils.IsValidEmail(id) {
			utils.ResponseError(w, "Failed to verify, invalid ID or password")
			return
		}

		if len(name) < 2 {
			utils.ResponseError(w, "Invalid name, too short")
			return
		}

		if len(code) != 6 {
			utils.ResponseError(w, "Invalid code, wrong length")
			return
		}

		target, err := strconv.ParseUint(targetStr, 10, 32)
		if err != nil {
			utils.ResponseError(w, "Invalid target, not a valid number")
			return
		}

		result := s.UserService.VerifyEmail(&models.VerifyParameter{
			Target:   uint(target),
			Code:     code,
			Id:       id,
			Password: pw,
			Name:     name,
		})

		if !result {
			utils.ResponseError(w, "Failed to verify code")
			return
		}
		utils.ResponseSuccess(w, nil)
	}
}

// 비밀번호 초기화하기
func ResetPasswordHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.FormValue("email")

		if !utils.IsValidEmail(id) {
			utils.ResponseError(w, "Failed to reset password, invalid ID(email)")
			return
		}

		result := s.UserService.ResetPassword(id, r)
		if !result {
			utils.ResponseError(w, "Unable to reset password, internal error")
			return
		}
		utils.ResponseSuccess(w, &models.ResetPasswordResult{
			Sendmail: configs.Env.GmailAppPassword != "",
		})
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
