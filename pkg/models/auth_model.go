package models

// 비밀번호 초기화 파라미터
type ResetPasswordParam struct {
	Email string `json:"email"`
}

// 네이티브 앱의 리프레시 토큰 회전에 필요한 파라미터다.
type MobileRefreshParam struct {
	Refresh string `json:"refresh"`
}

// 쿠키를 사용할 수 없는 네이티브 앱에 새 토큰 쌍을 전달한다.
type AuthTokenPair struct {
	Token   string `json:"token"`
	Refresh string `json:"refresh"`
}

// Apple 로그인 전에 서버가 발급하는 일회성 nonce다.
type AppleNonceResult struct {
	Nonce string `json:"nonce"`
}

// Apple ID 토큰 로그인과 기존 계정 연결에 공통으로 쓰는 요청이다.
type AppleAuthParam struct {
	IdentityToken string `json:"identityToken"`
	Nonce         string `json:"nonce"`
	Name          string `json:"name"`
}

// 서버에서 검증을 마친 Apple ID 토큰의 계정 식별 정보다.
type AppleIdentity struct {
	Subject       string
	Audience      string
	Email         string
	EmailVerified bool
}

type OAuthIdentityStatus struct {
	Linked bool `json:"linked"`
}

// 인증 완료하기 파라미터
type VerifyParam struct {
	Target   uint   `json:"target"`
	Code     string `json:"code"`
	ID       string `json:"id"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// 구글 OAuth 응답
type GoogleUser struct {
	ID            string `json:"id"`
	Audience      string `json:"aud"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// 네이버 OAuth 응답
type NaverUser struct {
	Response struct {
		Email        string `json:"email"`
		Nickname     string `json:"nickname"`
		ProfileImage string `json:"profile_image"`
	} `json:"response"`
}

// 카카오 OAuth 응답
type KakaoUser struct {
	ID           int64 `json:"id"`
	KakaoAccount struct {
		Email   string `json:"email"`
		Profile struct {
			Nickname        string `json:"nickname"`
			ProfileImageUrl string `json:"profile_image_url"`
		} `json:"profile"`
	} `json:"kakao_account"`
}

// 인증 메일 발송에 필요한 파라미터 정의
type SignupParam struct {
	ID       string `json:"id"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Invite   string `json:"invite"`
}

// 회원가입 시 리턴 타입
type SignupResult struct {
	Target               uint `json:"target"`
	RequiresVerification bool `json:"requiresVerification"`
	Completed            bool `json:"completed"`
}

type SignupStatus struct {
	Mode                     string `json:"mode"`
	MailConfigured           bool   `json:"mailConfigured"`
	OAuthRegistrationAllowed bool   `json:"oauthRegistrationAllowed"`
}

type SignupInviteCreateParam struct {
	Email       string `json:"email"`
	ExpiresDays uint   `json:"expiresDays"`
}

type SignupInvite struct {
	Uid       uint   `json:"uid"`
	Email     string `json:"email"`
	Created   int64  `json:"created"`
	Expires   int64  `json:"expires"`
	Used      int64  `json:"used"`
	Revoked   bool   `json:"revoked"`
	CreatedBy uint   `json:"createdBy"`
}

type SignupInviteCreated struct {
	SignupInvite
	Token string `json:"token"`
	URL   string `json:"url"`
}

// JWT 컨텍스트 키값 설정
type ContextKey string

var JwtClaimsKey = ContextKey("jwtClaims")

// JWT 오류 코드 정의
const (
	JWT_EMPTY_TOKEN = -10 + iota
	JWT_NOT_BEARER
	JWT_INVALID_TOKEN
	JWT_NO_CLAIMS
	JWT_NO_UID
)

// 로그인 시 입력 구조 정의
type SigninParam struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}

// 이메일이 사용중인지 확인할 때의 파라미터 정의
type CheckEmailParam struct {
	Email string `json:"email"`
}

// 닉네임이 사용중인지 확인할 때의 파라미터 정의
type CheckNameParam struct {
	Name string `json:"name"`
}

const AUTH_TOKEN = "nubo-auth-token"
const REFRESH_TOKEN = "nubo-refresh-token"
const OAUTH_STATE = "nubo-oauth-state"
