package handlers

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

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
