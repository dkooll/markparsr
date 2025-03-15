package markparsr

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileValidator checks that required files for a Terraform module exist and aren't empty.
type FileValidator struct {
	rootDir string
	files   []string
}

// NewFileValidator creates a validator that checks for certain files in module directory.
// It accepts both the README path and module path to check files in both locations.
func NewFileValidator(readmePath string, modulePath string) *FileValidator {
	// Use module path for terraform files
	files := []string{
		readmePath,
		filepath.Join(modulePath, "outputs.tf"),
		filepath.Join(modulePath, "variables.tf"),
		filepath.Join(modulePath, "terraform.tf"),
	}
	return &FileValidator{rootDir: modulePath, files: files}
}

// Validate checks that all required files exist and are not empty.
func (fv *FileValidator) Validate() []error {
	var allErrors []error
	for _, filePath := range fv.files {
		if err := validateFile(filePath); err != nil {
			allErrors = append(allErrors, err)
		}
	}
	return allErrors
}

// validateFile checks if a file exists and is not empty.
func validateFile(filePath string) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", filepath.Base(filePath))
		}
		return fmt.Errorf("error accessing file: %s: %w", filepath.Base(filePath), err)
	}
	if fileInfo.Size() == 0 {
		return fmt.Errorf("file is empty: %s", filepath.Base(filePath))
	}
	return nil
}
