package imageprocessor

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/h2non/bimg"
)

func TestBimgProcessorCreatesRequestedVariants(t *testing.T) {
	source := image.NewNRGBA(image.Rect(0, 0, 8, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 8; x++ {
			source.Set(x, y, color.NRGBA{R: uint8(x * 20), G: uint8(y * 30), B: 120, A: 255})
		}
	}
	var input bytes.Buffer
	if err := png.Encode(&input, source); err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	webpPath := filepath.Join(dir, "image.webp")
	jpegPath := filepath.Join(dir, "image.jpg")
	processor := NewBimgProcessor()
	if err := processor.ProcessBuffer(input.Bytes(), []Variant{
		{Path: webpPath, Width: 4, Quality: 90, Format: FormatWebP},
		{Path: jpegPath, Width: 2, Quality: 60, Format: FormatJPEG},
	}); err != nil {
		t.Fatal(err)
	}

	webp, err := os.ReadFile(webpPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(webp) < 12 || string(webp[:4]) != "RIFF" || string(webp[8:12]) != "WEBP" {
		t.Fatal("processor did not create a WebP variant")
	}
	webpSize, err := bimg.Size(webp)
	if err != nil {
		t.Fatal(err)
	}
	if webpSize.Width != 4 || webpSize.Height != 2 {
		t.Fatalf("unexpected WebP dimensions: %dx%d", webpSize.Width, webpSize.Height)
	}
	jpeg, err := os.ReadFile(jpegPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(jpeg) < 2 || jpeg[0] != 0xff || jpeg[1] != 0xd8 {
		t.Fatal("processor did not create a JPEG variant")
	}
	jpegSize, err := bimg.Size(jpeg)
	if err != nil {
		t.Fatal(err)
	}
	if jpegSize.Width != 2 || jpegSize.Height != 1 {
		t.Fatalf("unexpected JPEG dimensions: %dx%d", jpegSize.Width, jpegSize.Height)
	}
}

func TestBimgProcessorRejectsUnknownOutputFormat(t *testing.T) {
	processor := NewBimgProcessor()
	err := processor.ProcessBuffer([]byte("not decoded before format validation"), []Variant{{
		Path:   filepath.Join(t.TempDir(), "image.unknown"),
		Width:  1,
		Format: Format(99),
	}})
	if err == nil {
		t.Fatal("unknown output format was accepted")
	}
}
