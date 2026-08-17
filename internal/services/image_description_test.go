package services

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/utils"
)

func TestImageDescriptionCandidatesRequireOptInAndRespectPerPostLimit(t *testing.T) {
	files := []*multipart.FileHeader{
		{Filename: "first.jpg"},
		{Filename: "notes.txt"},
		{Filename: "second.png"},
		{Filename: "third.webp"},
	}

	disabled := newNuboBoardService(nil, configs.ImageDescriptionConfig{
		Enabled: false, MaxPerPost: 2, MaxConcurrent: 1,
	}, nil)
	if candidates := disabled.imageDescriptionCandidates(files); len(candidates) != 0 {
		t.Fatalf("disabled feature selected %d candidates", len(candidates))
	}

	enabled := newNuboBoardService(nil, configs.ImageDescriptionConfig{
		Enabled: true, MaxPerPost: 2, MaxConcurrent: 1,
	}, nil)
	candidates := enabled.imageDescriptionCandidates(files)
	if len(candidates) != 2 {
		t.Fatalf("selected %d candidates, want 2", len(candidates))
	}
	if _, ok := candidates[files[0]]; !ok {
		t.Error("first image was not selected")
	}
	if _, ok := candidates[files[2]]; !ok {
		t.Error("second image was not selected")
	}
	if _, ok := candidates[files[3]]; ok {
		t.Error("image beyond the per-post limit was selected")
	}
}

func TestRequestImageDescriptionRespectsConcurrencyGate(t *testing.T) {
	calls := 0
	service := newNuboBoardService(nil, configs.ImageDescriptionConfig{
		Enabled: true, Model: "test-model", MaxPerPost: 1, MaxConcurrent: 1,
	}, func(_ context.Context, _, model string) (utils.ImageDescriptionResult, error) {
		calls++
		return utils.ImageDescriptionResult{Description: "description", Model: model}, nil
	})

	service.imageDescriptionSlots <- struct{}{}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := service.requestImageDescription(cancelled, "image.webp"); !errors.Is(err, context.Canceled) {
		t.Fatalf("request error = %v, want context cancellation", err)
	}
	if calls != 0 {
		t.Fatalf("description generator called %d times while concurrency slot was occupied", calls)
	}
	<-service.imageDescriptionSlots

	result, err := service.requestImageDescription(context.Background(), "image.webp")
	if err != nil {
		t.Fatal(err)
	}
	if result.Description != "description" || result.Model != "test-model" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if calls != 1 {
		t.Fatalf("description generator called %d times, want 1", calls)
	}
}
