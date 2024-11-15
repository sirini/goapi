package utils

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/h2non/bimg"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

//                                                                              //
// 고품질의 이미지 생성을 위해 libvips 라이브러리를 사용하는 bimg 기반으로 구현 //
// macOS(homebrew): brew install vips                                           //
// Ubuntu Linux: sudo apt install libvips                                       //
// Windows: https://www.libvips.org/install.html                                //
//                                                                              //

// 바이트 버퍼 이미지를 지정된 크기로 줄여서 .avif 형식으로 저장
func SaveImage(inputBuffer []byte, outputPath string, width int) error {
	options := bimg.Options{
		Width:   width,
		Height:  0,
		Quality: 90,
		Type:    bimg.AVIF,
	}

	processed, err := bimg.NewImage(inputBuffer).Process(options)
	if err != nil {
		return err
	}

	err = bimg.Write(outputPath, processed)
	if err != nil {
		return err
	}
	return nil
}

// URL로부터 이미지 경로를 받아서 지정된 크기로 줄이고 .avif 형식으로 저장
func DownloadImage(imageUrl string, outputPath string, width int) error {
	resp, err := http.Get(imageUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buffer, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	SaveImage(buffer, outputPath, width)
	return nil
}

// 주어진 파일 경로가 이미지 파일인지 아닌지 확인하기
func IsImage(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".avif":
		return true
	case ".jpg":
		return true
	case ".jpeg":
		return true
	case ".png":
		return true
	case ".bmp":
		return true
	case ".webp":
		return true
	case ".gif":
		return true
	default:
		return false
	}
}

// 이미지를 주어진 크기로 줄여서 .avif 형식으로 저장하기
func ResizeImage(inputPath string, outputPath string, width int) error {
	buffer, err := bimg.Read(inputPath)
	if err != nil {
		return err
	}
	SaveImage(buffer, outputPath, width)
	return nil
}

// 이미지 저장하고 경로 반환
func SaveProfileImage(inputPath string) string {
	savePath, err := MakeSavePath(models.UPLOAD_PROFILE)
	if err != nil {
		return ""
	}

	outputPath := fmt.Sprintf("%s/%s.avif", savePath, uuid.New().String()[:8])
	err = ResizeImage("."+inputPath, outputPath, configs.Env.Number(configs.SIZE_PROFILE))
	if err != nil {
		return ""
	}

	return outputPath[1:]
}
