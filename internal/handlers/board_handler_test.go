package handlers

import (
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
)

type originalImageBoardService struct {
	services.BoardService
	result models.BoardOriginalImageResult
}

func (s originalImageBoardService) GetOriginalImage(uint, uint, uint) (models.BoardOriginalImageResult, error) {
	return s.result, nil
}

func TestConsumeDownloadTokenIsOneTime(t *testing.T) {
	h := &NuboBoardHandler{downloadTokenStorage: make(map[string]DownloadToken)}
	h.storeDownloadToken("token", DownloadToken{Name: "photo.jpg", Path: "/photo.jpg", Expiry: time.Now().Add(time.Minute)})

	var successes atomic.Int32
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, ok := h.consumeDownloadToken("token", time.Now()); ok {
				successes.Add(1)
			}
		}()
	}
	wg.Wait()
	if got := successes.Load(); got != 1 {
		t.Fatalf("token was consumed %d times, want exactly once", got)
	}
}

func TestConsumeDownloadTokenRejectsExpired(t *testing.T) {
	h := &NuboBoardHandler{downloadTokenStorage: make(map[string]DownloadToken)}
	now := time.Now()
	h.storeDownloadToken("expired", DownloadToken{Expiry: now.Add(-time.Second)})
	if _, ok := h.consumeDownloadToken("expired", now); ok {
		t.Fatal("expired token was accepted")
	}
}

func TestOriginalImageHandlerDoesNotExposeStoragePath(t *testing.T) {
	h := NewNuboBoardHandler(&services.Service{
		Board: originalImageBoardService{result: models.BoardOriginalImageResult{Path: "/upload/attachments/private.jpg"}},
	})
	app := fiber.New()
	app.Get("/board/original", h.OriginalImageHandler)

	response, err := app.Test(httptest.NewRequest("GET", "/board/original?boardUid=1&fileUid=2", nil))
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if response.StatusCode != fiber.StatusOK || strings.Contains(text, "/upload/") {
		t.Fatalf("original response status=%d body=%s", response.StatusCode, text)
	}
	if !strings.Contains(text, "/board/original/transfer?token=") || len(h.originalImageTokens) != 1 {
		t.Fatalf("original response did not contain a tokenized stream URL: %s", text)
	}
}

func TestOriginalImageTransferSupportsReusableByteRanges(t *testing.T) {
	oldUploadDir := configs.Env.UploadDir
	configs.Env.UploadDir = t.TempDir()
	t.Cleanup(func() { configs.Env.UploadDir = oldUploadDir })

	publicPath := "/upload/attachments/photo.jpg"
	filePath := filepath.Join(configs.Env.UploadDir, "attachments", "photo.jpg")
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filePath, []byte("original-image"), 0o644); err != nil {
		t.Fatal(err)
	}

	h := &NuboBoardHandler{originalImageTokens: make(map[string]OriginalImageToken)}
	h.storeOriginalImageToken("image-token", OriginalImageToken{Path: publicPath, Expiry: time.Now().Add(time.Minute)})
	app := fiber.New()
	app.Get("/board/original/transfer", h.OriginalImageTransferHandler)

	for range 2 {
		request := httptest.NewRequest("GET", "/board/original/transfer?token=image-token", nil)
		request.Header.Set("Range", "bytes=2-5")
		response, err := app.Test(request)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != fiber.StatusPartialContent || string(body) != "igin" {
			t.Fatalf("range response status=%d body=%q", response.StatusCode, body)
		}
		if disposition := response.Header.Get("Content-Disposition"); disposition != "inline" {
			t.Fatalf("content disposition = %q", disposition)
		}
		if cache := response.Header.Get("Cache-Control"); cache != "private, no-store" {
			t.Fatalf("cache control = %q", cache)
		}
	}
}
