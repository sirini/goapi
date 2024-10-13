package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	DBHost            string
	DBUser            string
	DBPass            string
	DBName            string
	DBTablePrefix     string
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

// .env 파일에서 설정 내용 불러오기
func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found. Please make sure that this goapi binary is locate in tsboard.git directory.")
	}

	config := &Config{
		Port:              getEnv("GOAPI_PORT", "3003"),
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBUser:            getEnv("DB_USER", ""),
		DBPass:            getEnv("DB_PASS", ""),
		DBName:            getEnv("DB_NAME", "tsboard"),
		DBTablePrefix:     getEnv("DB_TABLE_PREFIX", "tsb_"),
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

	return config
}
