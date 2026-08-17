package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

const publicUploadRoot = "/upload"

// UploadDirectory는 설정된 업로드 파일시스템 루트를 반환한다.
// 소스 설치 호환성을 위해 기존 상대 경로를 기본값으로 유지한다.
func UploadDirectory() string {
	directory := strings.TrimSpace(configs.Env.UploadDir)
	if directory == "" {
		directory = "." + publicUploadRoot
	}
	return filepath.Clean(directory)
}

// UploadFilePath는 공개 `/upload` 경로를 설정된 실제 디스크 경로로 변환한다.
func UploadFilePath(publicPath string) (string, error) {
	cleanPath := path.Clean("/" + strings.TrimSpace(publicPath))
	if cleanPath != publicUploadRoot && !strings.HasPrefix(cleanPath, publicUploadRoot+"/") {
		return "", fmt.Errorf("invalid upload path %q", publicPath)
	}
	relativePath := strings.TrimPrefix(cleanPath, publicUploadRoot)
	return filepath.Join(UploadDirectory(), filepath.FromSlash(relativePath)), nil
}

// PublicUploadPath는 설정된 업로드 디렉터리 아래의 파일을
// 기존 설치와 호환되는 URL/DB 경로로 변환한다.
func PublicUploadPath(filePath string) (string, error) {
	rootPath, err := filepath.Abs(UploadDirectory())
	if err != nil {
		return "", err
	}
	absolutePath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}
	relativePath, err := filepath.Rel(rootPath, absolutePath)
	if err != nil || relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("upload file %q is outside %q", filePath, UploadDirectory())
	}
	if relativePath == "." {
		return publicUploadRoot, nil
	}
	return publicUploadRoot + "/" + filepath.ToSlash(relativePath), nil
}

func RemoveUploadFile(publicPath string) error {
	filePath, err := UploadFilePath(publicPath)
	if err != nil {
		return err
	}
	return os.Remove(filePath)
}

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
	filePath, err := UploadFilePath(path)
	if err != nil {
		return 0
	}
	info, err := os.Stat(filePath)
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

	finalPath := filepath.Join(UploadDirectory(), string(target), year, month, day)
	err := os.MkdirAll(finalPath, os.ModePerm)
	if err != nil {
		return "", err
	}
	return finalPath, nil
}

// 업로드 된 파일을 attachments 폴더에 저장하고 경로 반환
func SaveAttachmentFile(file *multipart.FileHeader) (string, error) {
	result := ""
	savePath, err := MakeSavePath(models.UPLOAD_ATTACH)
	if err != nil {
		return result, err
	}
	randName := uuid.New().String()[:8]
	ext := filepath.Ext(file.Filename)
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

// 업로드 된 파일을 임시 폴더에 랜덤한 파일명으로 저장하고 경로 반환
func SaveUploadedFile(file multipart.File, fileName string) (string, error) {
	result := ""
	tempDir := filepath.Join(UploadDirectory(), string(models.UPLOAD_TEMP))
	err := os.MkdirAll(tempDir, os.ModePerm)
	if err != nil {
		return result, err
	}

	ext := filepath.Ext(fileName)
	safeFileName := uuid.New().String() + ext
	result = fmt.Sprintf("%s/%s", tempDir, safeFileName)

	if err = CopyFile(result, file); err != nil {
		return result, err
	}
	return result, nil
}
