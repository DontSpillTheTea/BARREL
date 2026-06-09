package security

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
)

func ValidateUpload(r *http.Request, maxUploadMB int64) error {
	r.Body = http.MaxBytesReader(nil, r.Body, maxUploadMB<<20)
	if err := r.ParseMultipartForm(maxUploadMB << 20); err != nil {
		return fmt.Errorf("file too large or invalid multipart form")
	}
	return nil
}

func IsAllowedExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".png" || ext == ".jpg" || ext == ".jpeg"
}
