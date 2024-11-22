package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/sirini/goapi/pkg/models"
)

// 대상 경로에 파일 복사하기
func CopyFile(destPath string, file multipart.File) error {
	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	if _, err := io.Copy(dest, file); err != nil {
		return err
	}
	return nil
}

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

// 업로드 된 파일을 attachments 폴더에 저장하고 경로 반환
func SaveAttachmentFile(file *multipart.FileHeader, fileName string) (string, error) {
	result := ""
	savePath, err := MakeSavePath(models.UPLOAD_ATTACH)
	if err != nil {
		return result, err
	}
	randName := uuid.New().String()[:8]
	ext := filepath.Ext(fileName)
	result = fmt.Sprintf("%s/%s%s", savePath, randName, ext)

	srcFile, err := file.Open()
	if err != nil {
		return result, err
	}
	defer srcFile.Close()

	if err = CopyFile(result, srcFile); err != nil {
		return result, err
	}
	return result, nil
}

// 업로드 된 파일을 임시 폴더에 저장하고 경로 반환
func SaveUploadedFile(file multipart.File, fileName string) (string, error) {
	result := ""
	tempDir := fmt.Sprintf("./upload/%s", models.UPLOAD_TEMP)
	err := os.MkdirAll(tempDir, os.ModePerm)
	if err != nil {
		return result, err
	}
	result = fmt.Sprintf("%s/%s", tempDir, fileName)

	if err = CopyFile(result, file); err != nil {
		return result, err
	}
	return result, nil
}
