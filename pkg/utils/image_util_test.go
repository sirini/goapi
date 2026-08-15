package utils

import (
	"io"
	"net/http"
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
