package utils

import (
	"bytes"
	"encoding/base64"
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

func TestEncodeImageReturnsWebPDataURL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "thumbnail.webp")
	payload := []byte("RIFF\x00\x00\x00\x00WEBP")
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}

	encoded, err := EncodeImage(path)
	if err != nil {
		t.Fatal(err)
	}
	const prefix = "data:image/webp;base64,"
	if !strings.HasPrefix(encoded, prefix) {
		t.Fatalf("encoded image is missing WebP data URL prefix: %q", encoded)
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(encoded, prefix))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decoded, payload) {
		t.Fatal("encoded image payload changed")
	}
}

func TestEncodeImageRejectsUnsupportedFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "image.bmp")
	if err := os.WriteFile(path, []byte("bitmap"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := EncodeImage(path); err == nil {
		t.Fatal("unsupported OpenAI image format was accepted")
	}
}
