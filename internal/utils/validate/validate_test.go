package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/supchaser/test_task/internal/utils/errs"
)

func TestValidateObjectLimit(t *testing.T) {
	tests := []struct {
		name           string
		currentObjects int
		expectedError  error
	}{
		{
			name:           "belowLimit",
			currentObjects: 2,
			expectedError:  nil,
		},
		{
			name:           "atLimit",
			currentObjects: 3,
			expectedError:  errs.ErrMaxObjectsReached,
		},
		{
			name:           "aboveLimit",
			currentObjects: 4,
			expectedError:  errs.ErrMaxObjectsReached,
		},
		{
			name:           "zeroObjects",
			currentObjects: 0,
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateObjectLimit(tt.currentObjects)
			assert.ErrorIs(t, err, tt.expectedError)
		})
	}
}

func TestValidateFileExtension(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		expectedError error
	}{
		{
			name:          "validPdfExtension",
			url:           "document.pdf",
			expectedError: nil,
		},
		{
			name:          "validJpegExtension",
			url:           "image.jpeg",
			expectedError: nil,
		},
		{
			name:          "validJpgExtension",
			url:           "photo.jpg",
			expectedError: nil,
		},
		{
			name:          "uppercaseExtension",
			url:           "image.JPEG",
			expectedError: nil,
		},
		{
			name:          "mixedCaseExtension",
			url:           "document.PdF",
			expectedError: nil,
		},
		{
			name:          "invalidExtension",
			url:           "script.js",
			expectedError: errs.ErrInvalidFileType,
		},
		{
			name:          "noExtension",
			url:           "README",
			expectedError: errs.ErrInvalidFileType,
		},
		{
			name:          "emptyString",
			url:           "",
			expectedError: errs.ErrInvalidFileType,
		},
		{
			name:          "dotFileNoExtension",
			url:           ".gitignore",
			expectedError: errs.ErrInvalidFileType,
		},
		{
			name:          "urlWithQueryParams",
			url:           "document.pdf?token=abc123",
			expectedError: errs.ErrInvalidFileType,
		},
		{
			name:          "urlWithPath",
			url:           "/path/to/document.pdf",
			expectedError: nil,
		},
		{
			name:          "fullUrl",
			url:           "https://example.com/images/photo.jpg",
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFileExtension(tt.url)
			assert.ErrorIs(t, err, tt.expectedError)
		})
	}
}

func TestAllowedExtensions(t *testing.T) {
	assert.True(t, allowedExtensions[".pdf"])
	assert.True(t, allowedExtensions[".jpeg"])
	assert.True(t, allowedExtensions[".jpg"])
	assert.False(t, allowedExtensions[".png"])
}

func TestMaxObjectsPerTaskConstant(t *testing.T) {
	assert.Equal(t, 3, maxObjectsPerTask)
}
