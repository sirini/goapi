package utils

import (
	"io"
	"net/http"

	"github.com/h2non/bimg"
)

//                                                                              //
// 고품질의 이미지 생성을 위해 libvips 라이브러리를 사용하는 bimg 기반으로 구현 //
// macOS(homebrew): brew install vips                                           //
// Ubuntu Linux: sudo apt install libvips                                       //
// Windows: https://www.libvips.org/install.html                                //
//                                                                              //

// 바이트 버퍼 이미지를 지정된 크기로 줄여서 .avif 형식으로 저장
func SaveImage(inputBuffer []byte, outputPath string, width int) {
	options := bimg.Options{
		Width:   width,
		Height:  0,
		Quality: 90,
		Type:    bimg.AVIF,
	}

	processed, err := bimg.NewImage(inputBuffer).Process(options)
	if err != nil {
		return
	}

	err = bimg.Write(outputPath, processed)
	if err != nil {
		return
	}
}

// URL로부터 이미지 경로를 받아서 지정된 크기로 줄이고 .avif 형식으로 저장
func DownloadImage(imageUrl string, outputPath string, width int) {
	resp, err := http.Get(imageUrl)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	buffer, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	SaveImage(buffer, outputPath, width)
}

// 이미지를 주어진 크기로 줄여서 .avif 형식으로 저장하기 (`libvips` 필요)
func ResizeImage(inputPath string, outputPath string, width int) {
	buffer, err := bimg.Read(inputPath)
	if err != nil {
		return
	}
	SaveImage(buffer, outputPath, width)
}
