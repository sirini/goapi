package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"time"

	"github.com/sirini/goapi/pkg/models"
)

// 파일의 크기 반환
func GetFileSize(path string) uint {
	info, err := os.Stat("." + path)
	if err != nil {
		return 0
	}
	return uint(info.Size())
}

// 파일 저장 경로 만들기 (맨 앞 `.` 은 DB에 넣을 때 빼줘야함)
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

// 업로드 된 파일을 임시 폴더에 저장하고 경로 반환
func SaveUploadedFile(file multipart.File, filename string) (string, error) {
	result := ""
	tempDir := fmt.Sprintf("./upload/%s", models.UPLOAD_TEMP)
	err := os.MkdirAll(tempDir, os.ModePerm)
	if err != nil {
		return result, err
	}

	result = fmt.Sprintf("%s/%s", tempDir, filename)
	dest, err := os.Create(result)
	if err != nil {
		return result, err
	}
	defer dest.Close()

	if _, err := io.Copy(dest, file); err != nil {
		return result, err
	}
	return result, nil
}
