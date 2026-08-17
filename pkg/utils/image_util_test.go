package utils

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func imageTestClient(status int, body string) *http.Client {
	return &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: status,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})}
}

func TestDownloadImageRejectsHTTPError(t *testing.T) {
	output := filepath.Join(t.TempDir(), "profile.webp")
	if err := downloadImage(imageTestClient(http.StatusNotFound, "not found"), "https://example.test/image", output, 64); err == nil {
		t.Fatal("HTTP error response was accepted as an image")
	}
}

func TestDownloadImagePropagatesInvalidImageError(t *testing.T) {
	output := filepath.Join(t.TempDir(), "profile.webp")
	if err := downloadImage(imageTestClient(http.StatusOK, "not an image"), "https://example.test/image", output, 64); err == nil {
		t.Fatal("image conversion error was ignored")
	}
}

func TestEncodeImageReturnsJPEGDataURL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "image.jpg")
	payload := []byte{0xff, 0xd8, 0xff, 0xd9}
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}

	encoded, err := EncodeImage(path)
	if err != nil {
		t.Fatal(err)
	}
	const prefix = "data:image/jpeg;base64,"
	if !strings.HasPrefix(encoded, prefix) {
		t.Fatalf("encoded image is missing JPEG data URL prefix: %q", encoded)
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(encoded, prefix))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decoded, payload) {
		t.Fatal("encoded image payload changed")
	}
}

func TestMakeTempJpegUsesBoundedVisionSize(t *testing.T) {
	source := image.NewNRGBA(image.Rect(0, 0, 1024, 512))
	for y := 0; y < source.Bounds().Dy(); y++ {
		for x := 0; x < source.Bounds().Dx(); x++ {
			source.Set(x, y, color.NRGBA{R: uint8(x), G: uint8(y), B: 120, A: 255})
		}
	}
	inputPath := filepath.Join(t.TempDir(), "source.png")
	input, err := os.Create(inputPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(input, source); err != nil {
		_ = input.Close()
		t.Fatal(err)
	}
	if err := input.Close(); err != nil {
		t.Fatal(err)
	}

	outputPath, err := MakeTempJpeg(inputPath)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(outputPath)
	if outputPath == inputPath {
		t.Fatal("temporary JPEG reused the source path")
	}
	output, err := os.Open(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	config, err := jpeg.DecodeConfig(output)
	if err != nil {
		t.Fatal(err)
	}
	if config.Width != imageDescriptionWidth || config.Height != imageDescriptionWidth/2 {
		t.Fatalf("vision JPEG dimensions = %dx%d, want %dx%d", config.Width, config.Height, imageDescriptionWidth, imageDescriptionWidth/2)
	}
}
