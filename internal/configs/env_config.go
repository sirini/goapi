package configs

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Version           string
	Port              string
	URL               string
	ProfileSize       string
	ContentInsertSize string
	ThumbnailSize     string
	FullSize          string
	FileSizeLimit     string
	DBHost            string
	DBUser            string
	DBPass            string
	DBName            string
	Prefix            string
	DBPort            string
	JWTSecretKey      string
	GmailID           string
	GmailAppPassword  string
	OAuthGoogleID     string
	OAuthGoogleSecret string
	OAuthNaverID      string
	OAuthNaverSecret  string
	OAuthKakaoID      string
	OAuthKakaoSecret  string
	OpenaiKey         string
}

// 환경변수에 기본값을 설정해주는 함수
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// 설정 저장한 변수
var Env Config

// .env 파일에서 설정 내용 불러오기
func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found. Please make sure that this goapi binary is locate in tsboard.git directory.")
	}

	Env = Config{
		Version:           getEnv("GOAPI_VERSION", "1.0.0-beta1"),
		Port:              getEnv("GOAPI_PORT", "3003"),
		URL:               getEnv("GOAPI_URL", "http://localhost:3003"),
		ProfileSize:       getEnv("GOAPI_PROFILE_SIZE", "256"),
		ContentInsertSize: getEnv("GOAPI_CONTENT_INSERT_SIZE", "640"),
		ThumbnailSize:     getEnv("GOAPI_THUMBNAIL_SIZE", "512"),
		FullSize:          getEnv("GOAPI_FULL_SIZE", "2400"),
		FileSizeLimit:     getEnv("GOAPI_FILE_SIZE_LIMIT", "104857600"),
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBUser:            getEnv("DB_USER", ""),
		DBPass:            getEnv("DB_PASS", ""),
		DBName:            getEnv("DB_NAME", "tsboard"),
		Prefix:            getEnv("DB_TABLE_PREFIX", "tsb_"),
		DBPort:            getEnv("DB_PORT", "3306"),
		JWTSecretKey:      getEnv("JWT_SECRET_KEY", ""),
		GmailID:           getEnv("GMAIL_ID", "sirini@gmail.com"),
		GmailAppPassword:  getEnv("GMAIL_APP_PASSWORD", ""),
		OAuthGoogleID:     getEnv("OAUTH_GOOGLE_CLIENT_ID", ""),
		OAuthGoogleSecret: getEnv("OAUTH_GOOGLE_SECRET", ""),
		OAuthNaverID:      getEnv("OAUTH_NAVER_CLIENT_ID", ""),
		OAuthNaverSecret:  getEnv("OAUTH_NAVER_SECRET", ""),
		OAuthKakaoID:      getEnv("OAUTH_KAKAO_CLIENT_ID", ""),
		OAuthKakaoSecret:  getEnv("OAUTH_KAKAO_SECRET", ""),
		OpenaiKey:         getEnv("OPENAI_API_KEY", ""),
	}
}

// 숫자 형태로 반환이 필요한 항목 정의
type Size uint8

const (
	SIZE_PROFILE = iota
	SIZE_CONTENT_INSERT
	SIZE_THUMBNAIL
	SIZE_FULL
	SIZE_FILE
)

// 사이즈 반환하기
func (c *Config) Number(s Size) int {
	var target string
	var defaultValue int

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

	size, err := strconv.ParseInt(target, 10, 32)
	if err != nil {
		return defaultValue
	}
	return int(size)
}
