package utils

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/sirini/goapi/internal/configs"
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

func TestNormalizeImageDescriptionCollapsesWhitespace(t *testing.T) {
	got := normalizeImageDescription("  해변의\n\n노을  풍경입니다.\t검색어: 해변, 노을  ")
	want := "해변의 노을 풍경입니다. 검색어: 해변, 노을"
	if got != want {
		t.Fatalf("normalizeImageDescription() = %q, want %q", got, want)
	}
}

func TestNormalizeImageDescriptionLimitsDatabaseLength(t *testing.T) {
	got := normalizeImageDescription(strings.Repeat("가", imageDescriptionMaxRunes+20))
	if utf8.RuneCountInString(got) != imageDescriptionMaxRunes {
		t.Fatalf("description length = %d, want %d", utf8.RuneCountInString(got), imageDescriptionMaxRunes)
	}
	if !strings.HasSuffix(got, "…") {
		t.Fatal("truncated description is missing an ellipsis")
	}
}

func TestAskImageDescriptionLive(t *testing.T) {
	imagePath := strings.TrimSpace(os.Getenv("NUBO_OPENAI_LIVE_TEST_IMAGE"))
	if imagePath == "" {
		t.Skip("set NUBO_OPENAI_LIVE_TEST_IMAGE to run the OpenAI vision smoke test")
	}
	if err := configs.LoadConfig(); err != nil {
		t.Fatal(err)
	}
	config := configs.GetImageDescriptionConfig()
	if !config.Enabled {
		t.Fatal("OpenAI image descriptions are not enabled in the active environment file")
	}

	result, err := AskImageDescription(context.Background(), imagePath, config.Model)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(result.Description) == "" {
		t.Fatal("OpenAI returned an empty image description")
	}
	if utf8.RuneCountInString(result.Description) > imageDescriptionMaxRunes {
		t.Fatalf("description exceeds database limit: %d", utf8.RuneCountInString(result.Description))
	}
	t.Logf(
		"model=%s input_tokens=%d output_tokens=%d runes=%d description=%s",
		result.Model,
		result.InputTokens,
		result.OutputTokens,
		utf8.RuneCountInString(result.Description),
		result.Description,
	)
}
