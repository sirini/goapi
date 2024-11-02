package utils

import (
	"os"
	"path/filepath"
	"time"
)

// 파일 저장 경로 만들기
func MakeSavePath(target string) (string, error) {
	today := time.Now()
	year := today.Format("2006")
	month := today.Format("01")
	day := today.Format("02")

	finalPath := filepath.Join("./upload", target, year, month, day)
	err := os.MkdirAll(finalPath, os.ModePerm)
	if err != nil {
		return "", err
	}
	return finalPath, nil
}
