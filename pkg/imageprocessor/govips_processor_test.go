package imageprocessor

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	_ "golang.org/x/image/webp"
)

func TestGovipsProcessorCreatesRequestedVariants(t *testing.T) {
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
	processor, err := NewGovipsProcessor()
	if err != nil {
		t.Fatal(err)
	}
	if err := processor.ProcessBuffer(input.Bytes(), []Variant{
		{Path: webpPath, Width: 4, Quality: 90, Format: FormatWebP},
		{Path: jpegPath, Width: 2, Quality: 60, Format: FormatJPEG},
	}); err != nil {
		t.Fatal(err)
	}

	assertVariant(t, webpPath, FormatWebP, 4, 2)
	assertVariant(t, jpegPath, FormatJPEG, 2, 1)
}

func TestGovipsProcessorRejectsUnknownOutputFormatBeforeDecoding(t *testing.T) {
	processor, err := NewGovipsProcessor()
	if err != nil {
		t.Fatal(err)
	}
	err = processor.ProcessBuffer([]byte("not decoded before format validation"), []Variant{{
		Path:   filepath.Join(t.TempDir(), "image.unknown"),
		Width:  1,
		Format: Format(99),
	}})
	if err == nil {
		t.Fatal("unknown output format was accepted")
	}
}

func assertVariant(t *testing.T, path string, format Format, width, height int) {
	t.Helper()
	buffer, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	switch format {
	case FormatWebP:
		if len(buffer) < 12 || string(buffer[:4]) != "RIFF" || string(buffer[8:12]) != "WEBP" {
			t.Fatal("processor did not create a WebP variant")
		}
	case FormatJPEG:
		if len(buffer) < 2 || buffer[0] != 0xff || buffer[1] != 0xd8 {
			t.Fatal("processor did not create a JPEG variant")
		}
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(buffer))
	if err != nil {
		t.Fatal(err)
	}
	if config.Width != width || config.Height != height {
		t.Fatalf("unexpected dimensions: %dx%d", config.Width, config.Height)
	}
}
