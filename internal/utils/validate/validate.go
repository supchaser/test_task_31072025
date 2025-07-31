package validate

import (
	"path/filepath"
	"strings"

	"github.com/supchaser/test_task/internal/utils/errs"
)

const (
	maxObjectsPerTask = 3
)

var allowedExtensions = map[string]bool{
	".pdf":  true,
	".jpeg": true,
	".jpg":  true,
}

func ValidateObjectLimit(currentObjects int) error {
	if currentObjects >= maxObjectsPerTask {
		return errs.ErrMaxObjectsReached
	}

	return nil
}

func ValidateFileExtension(url string) error {
	ext := strings.ToLower(filepath.Ext(url))
	if _, ok := allowedExtensions[ext]; !ok {
		return errs.ErrInvalidFileType
	}

	return nil
}
