package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAttachedImageJSONDoesNotExposeOriginalStoragePath(t *testing.T) {
	encoded, err := json.Marshal(BoardAttachedImage{
		File: BoardFile{Uid: 7, Path: "/upload/attachments/private-original.jpg"},
	})
	if err != nil {
		t.Fatal(err)
	}
	result := string(encoded)
	if strings.Contains(result, "/upload/") || strings.Contains(result, `"path"`) {
		t.Fatalf("attached image exposed its original storage path: %s", result)
	}
	if !strings.Contains(result, `"uid":7`) {
		t.Fatalf("attached image omitted its stable file uid: %s", result)
	}
}
