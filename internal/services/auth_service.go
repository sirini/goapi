package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/templates"
	"github.com/sirini/goapi/pkg/utils"
)

var ErrMailNotConfigured = errors.New("Resend is not configured; add a Resend API key to enable email features")
var ErrMailRateLimited = errors.New("please wait before requesting another email")
var ErrSignupDisabled = errors.New("new account registration is currently disabled")
var ErrInvalidInvite = errors.New("a valid invitation issued for this email is required")

const verificationRequestCooldown = time.Minute

type AuthService interface {
	CanAuthenticate(userUid uint) bool
	CheckEmailExists(id string) bool
	CheckNameExists(name string, userUid uint) bool
	CheckRefreshToken(userUid uint, refreshToken string) bool
	CheckUserPermission(userUid uint, action models.UserAction) bool
	ChangeHashForPassword(userUid uint, newBcryptHash string) error
	GetMyInfo(userUid uint) models.MyInfoResult
	GetUserAndHash(id string) (models.MyInfoResult, string)
	Logout(userUid uint)
	ResetPassword(param models.ResetPasswordParam) error
	Signin(id string, pw string) models.MyInfoResult
	SignupStatus() models.SignupStatus
	Signup(param models.SignupParam) (models.SignupResult, error)
	CreateSignupInvite(param models.SignupInviteCreateParam, createdBy uint) (models.SignupInviteCreated, error)
	ListSignupInvites(limit uint) ([]models.SignupInvite, error)
	RevokeSignupInvite(uid uint) error
	CanRegisterOAuthUser() bool
	SaveTokensInCookie(c fiber.Ctx, userUid uint) (string, string, error)
	RotateTokensInCookie(c fiber.Ctx, userUid uint, oldRefreshToken string) (string, error)
	VerifyEmail(param models.VerifyParam) (bool, error)
}

// 토큰의 사용자가 현재도 존재하며 로그인 가능한 상태인지 확인한다.
func (s *NuboAuthService) CanAuthenticate(userUid uint) bool {
	user := s.repos.Auth.FindMyInfoByUid(userUid)
	return user.Uid == userUid && !user.Blocked
}

func (s *NuboAuthService) RotateTokensInCookie(c fiber.Ctx, userUid uint, oldRefreshToken string) (string, error) {
	accessHours, refreshDays := configs.GetJWTAccessRefresh()
	authToken, err := utils.GenerateAccessToken(userUid, accessHours)
	if err != nil {
		return "", err
	}
	refreshToken, err := utils.GenerateRefreshToken(userUid, refreshDays)
	if err != nil {
		return "", err
	}
	if !s.repos.Auth.RotateRefreshToken(userUid, oldRefreshToken, refreshToken) {
		return "", fmt.Errorf("refresh token is no longer valid")
	}
	utils.SaveCookie(c, models.AUTH_TOKEN, authToken, accessHours)
	utils.SaveCookie(c, models.REFRESH_TOKEN, refreshToken, refreshDays*24)
	return authToken, nil
}

type NuboAuthService struct {
	repos  *repositories.Repository
	mailer utils.Mailer
}

// 리포지토리 묶음 주입받기
func NewNuboAuthService(repos *repositories.Repository) *NuboAuthService {
	return newNuboAuthService(repos, utils.NewResendMailer())
}

func newNuboAuthService(repos *repositories.Repository, mailer utils.Mailer) *NuboAuthService {
	return &NuboAuthService{repos: repos, mailer: mailer}
}

// 이메일 중복 체크
func (s *NuboAuthService) CheckEmailExists(id string) bool {
	return s.repos.User.IsEmailDuplicated(id)
}

// 이름 중복 체크
func (s *NuboAuthService) CheckNameExists(name string, userUid uint) bool {
	return s.repos.User.IsNameDuplicated(name, userUid)
}

func (s *NuboAuthService) SignupStatus() models.SignupStatus {
	mode := configs.GetSignupMode()
	return models.SignupStatus{
		Mode:                     mode,
		MailConfigured:           s.mailer.Configured(),
		OAuthRegistrationAllowed: mode == "verified_email",
	}
}

func (s *NuboAuthService) CanRegisterOAuthUser() bool {
	return configs.GetSignupMode() == "verified_email"
}

// 리프레시 토큰이 유효할 경우 새로운 액세스 토큰 발급하기
func (s *NuboAuthService) CheckRefreshToken(userUid uint, refreshToken string) bool {
	return s.repos.Auth.CheckRefreshToken(userUid, refreshToken)
}

// 사용자 권한 확인하기
func (s *NuboAuthService) CheckUserPermission(userUid uint, action models.UserAction) bool {
	return s.repos.Auth.CheckPermissionForAction(userUid, action)
}

// 사용자 비밀번호를 SHA256 해시값에서 Bcrypt 해시값으로 변경하기
func (s *NuboAuthService) ChangeHashForPassword(userUid uint, newBcryptHash string) error {
	return s.repos.Auth.UpdateUserPasswordHash(userUid, newBcryptHash)
}

// 로그인 한 내 정보 가져오기
func (s *NuboAuthService) GetMyInfo(userUid uint) models.MyInfoResult {
	return s.repos.Auth.FindMyInfoByUid(userUid)
}

// 사용자의 정보와 함께 기존에 저장된 비밀번호 해시값 가져오기
func (s *NuboAuthService) GetUserAndHash(id string) (models.MyInfoResult, string) {
	userUid := s.repos.Auth.FindUserUidById(id)
	userInfo := s.repos.Auth.FindMyInfoByUid(userUid)
	storedHash := s.repos.Auth.FindUserPasswordByUid(userUid)
	return userInfo, storedHash
}

// 로그아웃하기
func (s *NuboAuthService) Logout(userUid uint) {
	s.repos.Auth.ClearRefreshToken(userUid)
}

// 비밀번호 초기화하기
func (s *NuboAuthService) ResetPassword(param models.ResetPasswordParam) error {
	if !s.mailer.Configured() {
		return ErrMailNotConfigured
	}
	userUid := s.repos.Auth.FindUserUidById(param.Email)
	if userUid < 1 {
		return nil
	}
	if s.repos.Auth.VerificationRecentlyIssued(param.Email, verificationRequestCooldown) {
		return nil
	}

	code, err := generateVerificationCode()
	if err != nil {
		return fmt.Errorf("generate password reset code: %w", err)
	}
	verifyUid := s.repos.Auth.SaveVerificationCode(param.Email, code)
	if verifyUid < 1 {
		return fmt.Errorf("save password reset code")
	}
	resetURL := fmt.Sprintf("%s/auth/change-password/%d/%s", siteURL(), verifyUid, code)
	html, text, err := templates.RenderTransactionalMail(templates.MailContent{
		SiteName:    configs.Env.Title,
		SiteURL:     siteURL(),
		Preheader:   "비밀번호 재설정 요청을 확인해 주세요.",
		Label:       "Security",
		Heading:     "비밀번호를 재설정해 주세요",
		Body:        "아래 버튼을 눌러 새 비밀번호를 설정할 수 있습니다. 이 링크는 10분 동안 한 번만 사용할 수 있습니다.",
		ActionLabel: "비밀번호 재설정",
		ActionURL:   resetURL,
		Notice:      "본인이 요청하지 않았다면 이 메일을 무시해 주세요. 비밀번호는 변경되지 않습니다.",
	})
	if err != nil {
		s.repos.Auth.DeleteVerificationCode(verifyUid)
		return fmt.Errorf("render password reset email: %w", err)
	}
	delivery, err := s.mailer.Send(models.MailMessage{
		To:             param.Email,
		Subject:        fmt.Sprintf("[%s] 비밀번호 재설정 안내", configs.Env.Title),
		HTML:           html,
		Text:           text,
		IdempotencyKey: mailIdempotencyKey("password-reset", verifyUid, code),
		Tags:           map[string]string{"type": "password-reset"},
	})
	if err != nil {
		s.repos.Auth.DeleteVerificationCode(verifyUid)
		log.Printf("mail: password reset delivery failed for user %d: %v", userUid, err)
		return fmt.Errorf("send password reset email: %w", err)
	}
	log.Printf("mail: password reset accepted by %s as %s", delivery.Provider, delivery.MessageID)
	return nil
}

// 사용자 로그인 처리하기
func (s *NuboAuthService) Signin(id string, pw string) models.MyInfoResult {
	user := s.repos.Auth.FindMyInfoByIDPW(id, pw)
	if user.Uid < 1 {
		return user
	}

	accessHours, refreshDays := configs.GetJWTAccessRefresh()
	accessToken, err := utils.GenerateAccessToken(user.Uid, accessHours)
	if err != nil {
		return user
	}

	refreshToken, err := utils.GenerateRefreshToken(user.Uid, refreshDays)
	if err != nil {
		return user
	}

	user.Token = accessToken
	user.Refresh = refreshToken
	s.repos.Auth.SaveRefreshToken(user.Uid, refreshToken)
	s.repos.Auth.UpdateUserSignin(user.Uid)
	return user
}

// 신규 회원 바로 가입 혹은 인증 메일 발송
func (s *NuboAuthService) Signup(param models.SignupParam) (models.SignupResult, error) {
	mode := configs.GetSignupMode()
	signupResult := models.SignupResult{}
	if mode == "disabled" {
		return signupResult, ErrSignupDisabled
	}
	if mode == "invite_only" && strings.TrimSpace(param.Invite) == "" {
		return signupResult, ErrInvalidInvite
	}
	isDupId := s.repos.User.IsEmailDuplicated(param.ID)
	var target uint
	if isDupId {
		return signupResult, fmt.Errorf("email(%s) is already in use", param.ID)
	}

	name := utils.Escape(param.Name)
	isDupName := s.repos.User.IsNameDuplicated(name, 0)
	if isDupName {
		return signupResult, fmt.Errorf("name(%s) is already in use", name)
	}
	if !utils.IsValidSignupPassword(param.Password) || len(name) < 2 || len(name) > 30 {
		return signupResult, fmt.Errorf("invalid signup credentials")
	}

	if mode == "invite_only" {
		digest := sha256.Sum256([]byte(strings.TrimSpace(param.Invite)))
		_, err := s.repos.SignupInvite.ConsumeInviteAndCreateUser(
			hex.EncodeToString(digest[:]), param.ID, param.Password, name, time.Now().UnixMilli(),
		)
		if err != nil {
			if errors.Is(err, repositories.ErrInviteInvalid) {
				return signupResult, ErrInvalidInvite
			}
			return signupResult, err
		}
		return models.SignupResult{Completed: true}, nil
	}

	if !s.mailer.Configured() {
		return signupResult, ErrMailNotConfigured
	}
	if s.repos.Auth.VerificationRecentlyIssued(param.ID, verificationRequestCooldown) {
		return signupResult, ErrMailRateLimited
	}
	code, err := generateVerificationCode()
	if err != nil {
		return signupResult, fmt.Errorf("failed to generate verification code")
	}
	target = s.repos.Auth.SaveVerificationCode(param.ID, code)
	if target < 1 {
		return signupResult, fmt.Errorf("failed to save verification code")
	}
	html, text, err := templates.RenderTransactionalMail(templates.MailContent{
		SiteName:  configs.Env.Title,
		SiteURL:   siteURL(),
		Preheader: fmt.Sprintf("%s 가입 인증 코드", configs.Env.Title),
		Label:     "Welcome",
		Heading:   "이메일 주소를 확인해 주세요",
		Body:      fmt.Sprintf("아래 인증 코드를 %s 가입 화면에 입력해 주세요. 코드는 10분 동안 한 번만 사용할 수 있습니다.", configs.Env.Title),
		Highlight: code,
		Notice:    "본인이 가입을 요청하지 않았다면 이 메일을 무시해 주세요.",
	})
	if err != nil {
		s.repos.Auth.DeleteVerificationCode(target)
		return signupResult, fmt.Errorf("failed to render verification email")
	}
	delivery, err := s.mailer.Send(models.MailMessage{
		To:             param.ID,
		Subject:        fmt.Sprintf("[%s] 이메일 주소를 확인해 주세요", configs.Env.Title),
		HTML:           html,
		Text:           text,
		IdempotencyKey: mailIdempotencyKey("signup-verification", target, code),
		Tags:           map[string]string{"type": "signup-verification"},
	})
	if err != nil {
		s.repos.Auth.DeleteVerificationCode(target)
		log.Printf("mail: signup verification delivery failed: %v", err)
		return signupResult, fmt.Errorf("failed to send verification email")
	}
	log.Printf("mail: signup verification accepted by %s as %s", delivery.Provider, delivery.MessageID)

	signupResult = models.SignupResult{
		Target:               target,
		RequiresVerification: true,
	}
	return signupResult, nil
}

func (s *NuboAuthService) CreateSignupInvite(param models.SignupInviteCreateParam, createdBy uint) (models.SignupInviteCreated, error) {
	result := models.SignupInviteCreated{}
	email := strings.ToLower(strings.TrimSpace(param.Email))
	if !utils.IsValidEmail(email) {
		return result, fmt.Errorf("invalid invitation email")
	}
	if s.repos.User.IsEmailDuplicated(email) {
		return result, fmt.Errorf("email is already in use")
	}
	if param.ExpiresDays < 1 || param.ExpiresDays > 90 {
		return result, fmt.Errorf("invitation expiry must be between 1 and 90 days")
	}
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return result, err
	}
	token := base64.RawURLEncoding.EncodeToString(random)
	digest := sha256.Sum256([]byte(token))
	created := time.Now()
	expires := created.Add(time.Duration(param.ExpiresDays) * 24 * time.Hour)
	uid, err := s.repos.SignupInvite.CreateInvite(email, hex.EncodeToString(digest[:]), createdBy, created.UnixMilli(), expires.UnixMilli())
	if err != nil {
		return result, err
	}
	result.SignupInvite = models.SignupInvite{Uid: uid, Email: email, Created: created.UnixMilli(), Expires: expires.UnixMilli(), CreatedBy: createdBy}
	result.Token = token
	result.URL = fmt.Sprintf("%s/auth/join?invite=%s", siteURL(), token)
	return result, nil
}

func (s *NuboAuthService) ListSignupInvites(limit uint) ([]models.SignupInvite, error) {
	return s.repos.SignupInvite.ListInvites(limit)
}

func (s *NuboAuthService) RevokeSignupInvite(uid uint) error {
	return s.repos.SignupInvite.RevokeInvite(uid)
}

func generateVerificationCode() (string, error) {
	const codeSpace = 1_000_000
	value, err := rand.Int(rand.Reader, big.NewInt(codeSpace))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", value.Int64()), nil
}

func mailIdempotencyKey(kind string, target uint, secret string) string {
	digest := sha256.Sum256([]byte(secret))
	return fmt.Sprintf("%s/%d/%x", kind, target, digest[:8])
}

func siteURL() string {
	return strings.TrimRight(configs.Env.Domain, "/")
}

// 로그인 성공 시 액세스 토큰과 리프레시 토큰들을 쿠키에 보관하기
func (s *NuboAuthService) SaveTokensInCookie(c fiber.Ctx, userUid uint) (string, string, error) {
	accessHours, refreshDays := configs.GetJWTAccessRefresh()
	authToken, err := utils.GenerateAccessToken(userUid, accessHours)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := utils.GenerateRefreshToken(userUid, refreshDays)
	if err != nil {
		return authToken, "", err
	}

	utils.SaveCookie(c, models.AUTH_TOKEN, authToken, accessHours)
	utils.SaveCookie(c, models.REFRESH_TOKEN, refreshToken, refreshDays*24)
	s.repos.Auth.SaveRefreshToken(userUid, refreshToken)
	return authToken, refreshToken, nil
}

// 이메일 인증 완료하기
func (s *NuboAuthService) VerifyEmail(param models.VerifyParam) (bool, error) {
	if configs.GetSignupMode() != "verified_email" {
		return false, ErrSignupDisabled
	}
	if !utils.IsValidSignupPassword(param.Password) {
		return false, fmt.Errorf("invalid signup credentials")
	}
	_, ok := s.repos.Auth.ConsumeVerificationCode(param.Target, param.Code, param.ID)
	if !ok {
		return false, nil
	}
	return s.repos.User.InsertNewUser(param.ID, param.Password, utils.Escape(param.Name)) > 0, nil
}
