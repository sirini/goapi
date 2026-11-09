package securitylog

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func canonical(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	parent := absolute
	tail := []string{}
	for {
		resolved, err := filepath.EvalSymlinks(parent)
		if err == nil {
			for i := len(tail) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, tail[i])
			}
			return resolved, nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		next := filepath.Dir(parent)
		if next == parent {
			return "", err
		}
		tail = append(tail, filepath.Base(parent))
		parent = next
	}
}

// ValidatePrivateDirectory prevents accidentally serving security records as uploads.
func ValidatePrivateDirectory(logDir, uploadDir string) error {
	logPath, err := canonical(logDir)
	if err != nil {
		return err
	}
	uploadPath, err := canonical(uploadDir)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(uploadPath, logPath)
	if err != nil {
		return err
	}
	if relative == "." || relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return errors.New("security logs must be outside public upload directory")
	}
	return nil
}
