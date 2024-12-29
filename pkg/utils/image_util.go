package utils

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/h2non/bimg"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

//                                                                //
// 고품질의 이미지 생성을 위해 libvips 라이브러리를 사용하는 bimg 기반으로 구현   //
// macOS(homebrew): brew install vips                             //
// Ubuntu Linux: sudo apt install libvips-dev                     //
//                                                                //

// OpenAI의 API를 이용해서 사진에 대한 설명 가져오기
func AskImageDescription(path string) (string, error) {
	if len(configs.Env.OpenaiKey) < 1 {
		return "", fmt.Errorf("api key of openai is empty")
	}
	jpgTempPath, err := MakeTempJpeg(path)
	if err != nil {
		return "", err
	}
	defer os.Remove(jpgTempPath)
	encoded, err := EncodeImage(jpgTempPath)
	if err != nil {
		return "", err
	}

	client := openai.NewClient(option.WithAPIKey(configs.Env.OpenaiKey))
	result, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Model: openai.F("gpt-4o"),
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.ChatCompletionUserMessageParam{
				Role: openai.F(openai.ChatCompletionUserMessageParamRoleUser),
				Content: openai.F([]openai.ChatCompletionContentPartUnionParam{
					openai.ChatCompletionContentPartTextParam{
						Type: openai.F(openai.ChatCompletionContentPartTextTypeText),
						Text: openai.F("Describe the content of this image in Korean."),
					},
					openai.ChatCompletionContentPartImageParam{
						Type: openai.F(openai.ChatCompletionContentPartImageTypeImageURL),
						ImageURL: openai.F(openai.ChatCompletionContentPartImageImageURLParam{
							URL:    openai.F(fmt.Sprintf("data:image/jpeg;base64,%s", encoded)),
							Detail: openai.F(openai.ChatCompletionContentPartImageImageURLDetailLow),
						}),
					},
				}),
			},
		}),
	})
	if err != nil {
		return "", err
	}
	return result.Choices[0].Message.Content, nil
}

// URL로부터 이미지 경로를 받아서 지정된 크기로 줄이고 .webp 형식으로 저장
func DownloadImage(imageUrl string, outputPath string, width uint) error {
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

// 이미지를 Base64로 인코딩해서 문자열로 반환
func EncodeImage(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	base64Data := base64.StdEncoding.EncodeToString(fileData)
	return base64Data, nil
}

// EXIF 정보 추출
func ExtractExif(imagePath string) models.BoardExif {
	result := models.BoardExif{}
	f, err := os.Open(imagePath)
	if err != nil {
		return result
	}

	x, err := exif.Decode(f)
	if err != nil {
		return result
	}

	make, err := x.Get(exif.Make)
	if err == nil {
		result.Make, _ = make.StringVal()
	}

	model, err := x.Get(exif.Model)
	if err == nil {
		result.Model, _ = model.StringVal()
	}

	aperture, err := x.Get(exif.FNumber)
	if err == nil {
		numerator, denominator, _ := aperture.Rat2(0)
		result.Aperture = uint(float32(numerator) / float32(denominator) * models.EXIF_APERTURE_FACTOR)
	}

	iso, err := x.Get(exif.ISOSpeedRatings)
	if err == nil {
		isoNum, _ := iso.Int(0)
		result.ISO = uint(isoNum)
	}

	focalLength, err := x.Get(exif.FocalLengthIn35mmFilm)
	if err == nil {
		fl, _ := focalLength.Int(0)
		result.FocalLength = uint(fl)
	}

	exposure, err := x.Get(exif.ExposureTime)
	if err == nil {
		numerator, denominator, _ := exposure.Rat2(0)
		result.Exposure = uint(float32(numerator) / float32(denominator) * models.EXIF_EXPOSURE_FACTOR)
	}

	width, err := x.Get(exif.PixelXDimension)
	if err == nil {
		w, _ := width.Int(0)
		result.Width = uint(w)
	}

	height, _ := x.Get(exif.PixelYDimension)
	if err == nil {
		h, _ := height.Int(0)
		result.Height = uint(h)
	}

	date, err := x.Get(exif.DateTime)
	if err == nil {
		timeStr, _ := date.StringVal()
		result.Date = ConvUnixMilli(timeStr)
	}
	return result
}

// 이미지 비전용으로 잠시 사용하고 삭제할 고압축 미니 썸네일 생성
func MakeTempJpeg(path string) (string, error) {
	buffer, err := bimg.Read(path)
	if err != nil {
		return "", err
	}

	jpgTempPath := strings.ReplaceAll(path, ".webp", ".jpg")
	options := bimg.Options{
		Width:   int(configs.SIZE_PROFILE.Number()),
		Height:  0,
		Quality: 60,
		Type:    bimg.JPEG,
	}

	processed, err := bimg.NewImage(buffer).Process(options)
	if err != nil {
		return "", err
	}
	err = bimg.Write(jpgTempPath, processed)
	if err != nil {
		return "", err
	}
	return jpgTempPath, nil
}

// 이미지를 주어진 크기로 줄여서 .webp 형식으로 저장하기
func ResizeImage(inputPath string, outputPath string, width uint) error {
	buffer, err := bimg.Read(inputPath)
	if err != nil {
		return err
	}
	SaveImage(buffer, outputPath, width)
	return nil
}

// 바이트 버퍼 이미지를 지정된 크기로 줄여서 .webp 형식으로 저장
func SaveImage(inputBuffer []byte, outputPath string, width uint) error {
	options := bimg.Options{
		Width:   int(width),
		Height:  0,
		Quality: 90,
		Type:    bimg.WEBP,
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

// 본문 삽입용 이미지 저장하고 경로 반환
func SaveInsertImage(inputPath string) (string, error) {
	result := ""
	savePath, err := MakeSavePath(models.UPLOAD_IMAGE)
	if err != nil {
		return result, err
	}

	result = fmt.Sprintf("%s/%s.webp", savePath, uuid.New().String()[:8])
	err = ResizeImage(inputPath, result, configs.SIZE_CONTENT_INSERT.Number())
	if err != nil {
		return result, err
	}
	return result, nil
}

// 프로필 이미지 저장하고 경로 반환
func SaveProfileImage(inputPath string) (string, error) {
	result := ""
	savePath, err := MakeSavePath(models.UPLOAD_PROFILE)
	if err != nil {
		return result, err
	}

	result = fmt.Sprintf("%s/%s.webp", savePath, uuid.New().String()[:8])
	err = ResizeImage(inputPath, result, configs.SIZE_PROFILE.Number())
	if err != nil {
		return result, err
	}
	return result, nil
}

// 썸네일 이미지 저장하고 경로 반환
func SaveThumbnailImage(inputPath string) (models.BoardThumbnail, error) {
	result := models.BoardThumbnail{}
	savePath, err := MakeSavePath(models.UPLOAD_THUMB)
	if err != nil {
		return result, err
	}

	randName := uuid.New().String()[:8]
	result.Small = fmt.Sprintf("%s/t%s.webp", savePath, randName)
	result.Large = fmt.Sprintf("%s/f%s.webp", savePath, randName)

	err = ResizeImage(inputPath, result.Small, configs.SIZE_THUMBNAIL.Number())
	if err != nil {
		return result, err
	}
	err = ResizeImage(inputPath, result.Large, configs.SIZE_FULL.Number())
	if err != nil {
		return result, err
	}
	return result, nil
}
