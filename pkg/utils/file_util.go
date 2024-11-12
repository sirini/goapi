package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"time"

	"github.com/sirini/goapi/pkg/models"
)

// 파일 저장 경로 만들기
func MakeSavePath(target models.UploadCategory) (string, error) {
	today := time.Now()
	year := today.Format("2006")
	month := today.Format("01")
	day := today.Format("02")

	finalPath := fmt.Sprintf("./upload/%s/%s/%s/%s", string(target), year, month, day)
	err := os.MkdirAll(finalPath, os.ModePerm)
	if err != nil {
		return "", err
	}
	return finalPath, nil
}

// 업로드 된 파일 저장하고 경로 반환 (맨 앞 . 제거)
func SaveUploadedFile(target models.UploadCategory, file multipart.File, filename string) string {
	dirPath, err := MakeSavePath(target)
	if err != nil {
		return ""
	}

	savePath := fmt.Sprintf("%s/%s", dirPath, filename)
	dest, err := os.Create(savePath)
	if err != nil {
		return ""
	}
	defer dest.Close()

	if _, err := io.Copy(dest, file); err != nil {
		return ""
	}
	return savePath[1:]
}

// 파일의 크기 반환
func GetFileSize(path string) uint {
	info, err := os.Stat("." + path)
	if err != nil {
		return 0
	}
	return uint(info.Size())
}
