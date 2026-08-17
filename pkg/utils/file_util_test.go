package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

func TestUploadPathsUseConfiguredDirectoryAndStablePublicPath(t *testing.T) {
	original := configs.Env
	t.Cleanup(func() { configs.Env = original })

	uploadRoot := filepath.Join(t.TempDir(), "existing-site", "upload")
	configs.Env.UploadDir = uploadRoot

	savePath, err := MakeSavePath(models.UPLOAD_ATTACH)
	if err != nil {
		t.Fatal(err)
	}
	filePath := filepath.Join(savePath, "example.txt")
	if err := os.WriteFile(filePath, []byte("nubo"), 0600); err != nil {
		t.Fatal(err)
	}

	publicPath, err := PublicUploadPath(filePath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(publicPath, "/upload/attachments/") {
		t.Fatalf("public path = %q", publicPath)
	}
	resolvedPath, err := UploadFilePath(publicPath)
	if err != nil {
		t.Fatal(err)
	}
	if resolvedPath != filePath {
		t.Fatalf("resolved path = %q, want %q", resolvedPath, filePath)
	}
	if size := GetFileSize(publicPath); size != 4 {
		t.Fatalf("file size = %d, want 4", size)
	}
	if err := RemoveUploadFile(publicPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatalf("removed file stat error = %v", err)
	}
}

func TestUploadFilePathRejectsEscapes(t *testing.T) {
	original := configs.Env
	t.Cleanup(func() { configs.Env = original })
	configs.Env.UploadDir = t.TempDir()

	for _, candidate := range []string{"/etc/passwd", "/upload/../../etc/passwd", "../outside"} {
		if resolved, err := UploadFilePath(candidate); err == nil {
			t.Fatalf("UploadFilePath(%q) = %q, want error", candidate, resolved)
		}
	}
}

func TestUploadDirectoryKeepsLegacyDefault(t *testing.T) {
	original := configs.Env
	t.Cleanup(func() { configs.Env = original })
	configs.Env.UploadDir = ""

	if directory := UploadDirectory(); directory != "upload" {
		t.Fatalf("upload directory = %q, want upload", directory)
	}
}
