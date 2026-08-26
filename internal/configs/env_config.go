package configs

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	GoPort                  string
	GoHost                  string
	GoapiBase               string
	Domain                  string
	Title                   string
	Version                 string
	ProfileSize             string
	ContentInsertSize       string
	ThumbnailSize           string
	FullSize                string
	FileSizeLimit           string
	UploadDir               string
	DBHost                  string
	DBUser                  string
	DBPass                  string
	DBName                  string
	Prefix                  string
	DBPort                  string
	DBSocket                string
	DBMaxIdle               string
	DBMaxOpen               string
	AdminID                 string
	AdminPW                 string
	JWTSecretKey            string
	SyncSecretKey           string
	JWTAccessHours          string
	JWTRefreshDays          string
	ResendKey               string
	ResendFromEmail         string
	ResendFromName          string
	ResendReplyToEmail      string
	SignupMode              string
	OAuthGoogleID           string
	OAuthGoogleSecret       string
	OAuthGoogleAndroidID    string
	OAuthNaverID            string
	OAuthNaverSecret        string
	OAuthKakaoID            string
	OAuthKakaoSecret        string
	OpenaiKey               string
	FirebaseProjectID       string
	FirebaseCredentialsFile string
	ImageDescription        ImageDescriptionEnv
}

type ImageDescriptionEnv struct {
	Enabled     string
	Model       string
	MaxPerPost  string
	Concurrency string
}

type ImageDescriptionConfig struct {
	Enabled       bool
	Model         string
	MaxPerPost    int
	MaxConcurrent int
}

const EnvironmentFileVariable = "NUBO_ENV_FILE"

// EnvironmentFilePath는 명시적인 런타임 설정 경로를 반환하며,
// 별도 지정이 없으면 기존 방식대로 작업 디렉터리의 .env를 사용한다.
func EnvironmentFilePath() string {
	if path := strings.TrimSpace(os.Getenv(EnvironmentFileVariable)); path != "" {
		return path
	}
	return ".env"
}

func GetImageDescriptionConfig() ImageDescriptionConfig {
	enabled, err := strconv.ParseBool(strings.TrimSpace(Env.ImageDescription.Enabled))
	if err != nil {
		enabled = false
	}

	model := strings.TrimSpace(Env.ImageDescription.Model)
	if model == "" {
		model = "gpt-5.6-luna"
	}

	return ImageDescriptionConfig{
		Enabled:       enabled && strings.TrimSpace(Env.OpenaiKey) != "",
		Model:         model,
		MaxPerPost:    parseBoundedInt(Env.ImageDescription.MaxPerPost, 3, 0, 100),
		MaxConcurrent: parseBoundedInt(Env.ImageDescription.Concurrency, 1, 1, 10),
	}
}

func parseBoundedInt(value string, fallback, minimum, maximum int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed < minimum || parsed > maximum {
		return fallback
	}
	return parsed
}

func GetSignupMode() string {
	mode := strings.ToLower(strings.TrimSpace(Env.SignupMode))
	switch mode {
	case "invite_only", "disabled":
		return mode
	default:
		return "verified_email"
	}
}

// 환경변수, 설정 파일, 기본값 순서로 설정값을 반환한다.
func getConfigValue(fileValues map[string]string, key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	if value, exists := fileValues[key]; exists {
		return value
	}
	return defaultValue
}

// 설정 저장한 변수
var Env Config

// 명시한 환경 파일 또는 기본 .env 파일에서 설정 내용을 불러온다.
func LoadConfig() error {
	path := EnvironmentFilePath()
	fileValues, err := godotenv.Read(path)
	if err != nil {
		return fmt.Errorf("load environment file %q: %w", path, err)
	}
	getEnv := func(key, defaultValue string) string {
		return getConfigValue(fileValues, key, defaultValue)
	}

	Env = Config{
		Version:                 getEnv("GOAPI_VERSION", "1.3.0"),
		GoapiBase:               getEnv("GOAPI_BASE", "goapi"),
		GoHost:                  getEnv("GOAPI_HOST", "0.0.0.0"),
		GoPort:                  getEnv("GOAPI_PORT", "3006"),
		Domain:                  getEnv("GOAPI_DOMAIN", "http://localhost"),
		Title:                   getEnv("GOAPI_TITLE", "NUBO"),
		ProfileSize:             getEnv("GOAPI_PROFILE_SIZE", "256"),
		ContentInsertSize:       getEnv("GOAPI_CONTENT_INSERT_SIZE", "640"),
		ThumbnailSize:           getEnv("GOAPI_THUMBNAIL_SIZE", "512"),
		FullSize:                getEnv("GOAPI_FULL_SIZE", "2400"),
		FileSizeLimit:           getEnv("GOAPI_FILE_SIZE_LIMIT", "104857600"),
		UploadDir:               getEnv("NUBO_UPLOAD_DIR", "./upload"),
		DBHost:                  getEnv("DB_HOST", "localhost"),
		DBUser:                  getEnv("DB_USER", ""),
		DBPass:                  getEnv("DB_PASS", ""),
		DBName:                  getEnv("DB_NAME", "nubo"),
		Prefix:                  getEnv("DB_TABLE_PREFIX", "nubo_"),
		DBPort:                  getEnv("DB_PORT", "3306"),
		DBSocket:                getEnv("DB_UNIX_SOCKET", ""),
		DBMaxIdle:               getEnv("DB_MAX_IDLE", "10"),
		DBMaxOpen:               getEnv("DB_MAX_OPEN", "10"),
		AdminID:                 getEnv("ADMIN_ID", ""),
		AdminPW:                 getEnv("ADMIN_PW", ""),
		JWTSecretKey:            getEnv("JWT_SECRET_KEY", ""),
		SyncSecretKey:           getEnv("SYNC_SECRET_KEY", ""),
		JWTAccessHours:          getEnv("JWT_ACCESS_HOURS", "2"),
		JWTRefreshDays:          getEnv("JWT_REFRESH_DAYS", "30"),
		ResendKey:               getEnv("RESEND_API_KEY", ""),
		ResendFromEmail:         getEnv("RESEND_FROM_EMAIL", ""),
		ResendFromName:          getEnv("RESEND_FROM_NAME", ""),
		ResendReplyToEmail:      getEnv("RESEND_REPLY_TO_EMAIL", ""),
		SignupMode:              getEnv("SIGNUP_MODE", "verified_email"),
		OAuthGoogleID:           getEnv("OAUTH_GOOGLE_CLIENT_ID", ""),
		OAuthGoogleSecret:       getEnv("OAUTH_GOOGLE_SECRET", ""),
		OAuthGoogleAndroidID:    getEnv("OAUTH_GOOGLE_ANDROID_CLIENT_ID", ""),
		OAuthNaverID:            getEnv("OAUTH_NAVER_CLIENT_ID", ""),
		OAuthNaverSecret:        getEnv("OAUTH_NAVER_SECRET", ""),
		OAuthKakaoID:            getEnv("OAUTH_KAKAO_CLIENT_ID", ""),
		OAuthKakaoSecret:        getEnv("OAUTH_KAKAO_SECRET", ""),
		OpenaiKey:               getEnv("OPENAI_API_KEY", ""),
		FirebaseProjectID:       getEnv("FIREBASE_PROJECT_ID", ""),
		FirebaseCredentialsFile: getEnv("FIREBASE_CREDENTIALS_FILE", ""),
		ImageDescription: ImageDescriptionEnv{
			Enabled:     getEnv("OPENAI_IMAGE_DESCRIPTION_ENABLED", "false"),
			Model:       getEnv("OPENAI_IMAGE_DESCRIPTION_MODEL", "gpt-5.6-luna"),
			MaxPerPost:  getEnv("OPENAI_IMAGE_DESCRIPTION_MAX_PER_POST", "3"),
			Concurrency: getEnv("OPENAI_IMAGE_DESCRIPTION_CONCURRENCY", "1"),
		},
	}
	return nil
}

// GetGoogleAndroidClientID는 Android ID 토큰의 audience를 반환한다.
// 전용 설정이 없는 기존 배포에서는 웹 OAuth client ID를 그대로 사용한다.
func GetGoogleAndroidClientID() string {
	if clientID := strings.TrimSpace(Env.OAuthGoogleAndroidID); clientID != "" {
		return clientID
	}
	return strings.TrimSpace(Env.OAuthGoogleID)
}

// 숫자 형태로 반환이 필요한 항목 정의
type ImageSize uint8

const (
	SIZE_PROFILE ImageSize = iota
	SIZE_CONTENT_INSERT
	SIZE_THUMBNAIL
	SIZE_FULL
	SIZE_FILE
)

// 사이즈 반환하기
func (s ImageSize) Number() uint {
	var target string
	var defaultValue uint

	switch s {
	case SIZE_CONTENT_INSERT:
		target = Env.ContentInsertSize
		defaultValue = 640
	case SIZE_THUMBNAIL:
		target = Env.ThumbnailSize
		defaultValue = 512
	case SIZE_FULL:
		target = Env.FullSize
		defaultValue = 2400
	case SIZE_FILE:
		target = Env.FileSizeLimit
		defaultValue = 104857600
	default:
		target = Env.ProfileSize
		defaultValue = 256
	}

	size, err := strconv.ParseUint(target, 10, 32)
	if err != nil {
		return defaultValue
	}
	return uint(size)
}

// HTTP 요청 크기 제한값 가져오기
func GetFileSizeLimit() int {
	size, err := strconv.ParseInt(Env.FileSizeLimit, 10, 32)
	if err != nil {
		return 10485760 /* 10MB */
	}
	return int(size)
}

// JWT 유효 기간 (access: hours, refresh: days) 반환
func GetJWTAccessRefresh() (int, int) {
	access := 2
	if accessHours, err := strconv.ParseInt(Env.JWTAccessHours, 10, 32); err == nil {
		access = int(accessHours)
	}
	refresh := 30
	if refreshDays, err := strconv.ParseInt(Env.JWTRefreshDays, 10, 32); err == nil {
		refresh = int(refreshDays)
	}
	return access, refresh
}
