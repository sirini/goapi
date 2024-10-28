package models

// (공개된) 사용자 정보
type UserInfoResult struct {
	Uid       uint   `json:"uid"`
	Name      string `json:"name"`
	Profile   string `json:"profile"`
	Level     uint   `json:"level"`
	Signature string `json:"signature"`
	Signup    uint64 `json:"signup"`
	Signin    uint64 `json:"signin"`
	Admin     bool   `json:"admin"`
	Blocked   bool   `json:"blocked"`
}

// (로그인 한) 내 정보
type MyInfoResult struct {
	UserInfoResult
	Id      string `json:"id"`
	Point   uint   `json:"point"`
	Token   string `json:"token"`
	Refresh string `json:"refresh"`
}

// 회원가입 시 리턴 타입
type SignupResult struct {
	Sendmail bool `json:"sendmail"`
	Target   uint `json:"target"`
}

// 인증 완료하기 파라미터
type VerifyParameter struct {
	Target   uint
	Code     string
	Id       string
	Password string
	Name     string
}

// 비밀번호 초기화 시 리턴 타입
type ResetPasswordResult struct {
	Sendmail bool `json:"sendmail"`
}

// 권한 확인이 필요한 액션 정의
type Action uint8

// 액션 고유 값들
const (
	WRITE_POST = iota
	WRITE_COMMENT
	SEND_CHAT
	SEND_REPORT
)

// 액션 이름 반환
func (a Action) String() string {
	switch a {
	case WRITE_COMMENT:
		return "write_comment"
	case SEND_CHAT:
		return "send_chat"
	case SEND_REPORT:
		return "send_report"
	default:
		return "write_post"
	}
}
