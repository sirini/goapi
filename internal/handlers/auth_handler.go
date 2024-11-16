package handlers

import (
	"html"
	"net/http"
	"strconv"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

// (회원가입 시) 이메일 주소가 이미 등록되어 있는지 확인하기
func CheckEmailHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.FormValue("email")
		if !utils.IsValidEmail(id) {
			utils.Error(w, "Invalid email address")
			return
		}

		result := s.Auth.CheckEmailExists(id)
		if result {
			utils.Error(w, "Email address is already in use")
			return
		}
		utils.Success(w, nil)
	}
}

// (회원가입 시) 이름이 이미 등록되어 있는지 확인하기
func CheckNameHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		if len(name) < 2 {
			utils.Error(w, "Invalid name, too short")
			return
		}

		result := s.Auth.CheckNameExists(name)
		if result {
			utils.Error(w, "Name is already in use")
			return
		}
		utils.Success(w, nil)
	}
}

// 로그인 한 사용자의 정보 불러오기
func LoadMyInfoHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userUid, err := strconv.ParseUint(r.FormValue("userUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid user uid, not a valid number")
			return
		}

		myinfo := s.Auth.GetMyInfo(uint(userUid))
		if myinfo.Uid < 1 {
			utils.Error(w, "Unable to load your information")
			return
		}
		utils.Success(w, myinfo)
	}
}

// 로그아웃 처리하기
func LogoutHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userUid := utils.GetUserUidFromToken(r)
		if userUid < 1 {
			utils.Error(w, "Unable to get an user uid from token")
			return
		}
		s.Auth.Logout(userUid)
		utils.Success(w, nil)
	}
}

// 비밀번호 초기화하기
func ResetPasswordHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.FormValue("email")

		if !utils.IsValidEmail(id) {
			utils.Error(w, "Failed to reset password, invalid ID(email)")
			return
		}

		result := s.Auth.ResetPassword(id, r)
		if !result {
			utils.Error(w, "Unable to reset password, internal error")
			return
		}
		utils.Success(w, &models.ResetPasswordResult{
			Sendmail: configs.Env.GmailAppPassword != "",
		})
	}
}

// 로그인 하기
func SigninHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.FormValue("id")
		pw := r.FormValue("password")

		if len(pw) != 64 || !utils.IsValidEmail(id) {
			utils.Error(w, "Failed to sign in, invalid ID or password")
			return
		}

		user := s.Auth.Signin(id, pw)
		if user.Uid < 1 {
			utils.Error(w, "Unable to get an information, invalid ID or password")
			return
		}
		utils.Success(w, user)
	}
}

// 회원가입 하기
func SignupHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.FormValue("email")
		pw := r.FormValue("password")
		name := r.FormValue("name")

		if len(pw) != 64 || !utils.IsValidEmail(id) {
			utils.Error(w, "Failed to sign up, invalid ID or password")
			return
		}

		result, err := s.Auth.Signup(id, pw, name, r)
		if err != nil {
			utils.Error(w, err.Error())
			return
		}
		utils.Success(w, result)
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
			utils.Error(w, "Failed to verify, invalid ID or password")
			return
		}

		if len(name) < 2 {
			utils.Error(w, "Invalid name, too short")
			return
		}

		if len(code) != 6 {
			utils.Error(w, "Invalid code, wrong length")
			return
		}

		target, err := strconv.ParseUint(targetStr, 10, 32)
		if err != nil {
			utils.Error(w, "Invalid target, not a valid number")
			return
		}

		result := s.Auth.VerifyEmail(models.VerifyParameter{
			Target:   uint(target),
			Code:     code,
			Id:       id,
			Password: pw,
			Name:     name,
		})

		if !result {
			utils.Error(w, "Failed to verify code")
			return
		}
		utils.Success(w, nil)
	}
}

// 로그인 한 사용자 정보 업데이트
func UpdateMyInfoHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := html.EscapeString(r.FormValue("name"))
		signature := html.EscapeString(r.FormValue("signature"))
		password := r.FormValue("password")

		if len(name) < 2 {
			utils.Error(w, "Invalid name, too short")
			return
		}

		userUid64, err := strconv.ParseUint(r.FormValue("userUid"), 10, 32)
		if err != nil {
			utils.Error(w, "Invalid access user uid, not a valid number")
			return
		}

		if isDup := s.Auth.CheckNameExists(name); isDup {
			utils.Error(w, "Duplicated name, please choose another one")
			return
		}

		userUid := uint(userUid64)
		userInfo, err := s.User.GetUserInfo(userUid)
		if err != nil {
			utils.Error(w, "Unable to find your information")
			return
		}

		userInfo.Name = name
		userInfo.Signature = signature
		file, handler, _ := r.FormFile("profile")
		if file != nil {
			defer file.Close()
		}

		param := models.UpdateUserInfoParameter{
			UserUid:        userUid,
			Name:           name,
			Signature:      signature,
			Password:       password,
			Profile:        file,
			ProfileHandler: handler,
		}
		s.User.ChangeUserInfo(param, userInfo)
		utils.Success(w, nil)
	}
}
