package markparsr

import (
	"os"
	"path/filepath"
)

// FileValidator validates the presence of required files
type FileValidator struct {
	files []string
}

// NewFileValidator creates a new FileValidator
func NewFileValidator(readmePath string) *FileValidator {
	rootDir := filepath.Dir(readmePath)
	files := []string{
		readmePath,
		filepath.Join(rootDir, "CONTRIBUTING.md"),
		filepath.Join(rootDir, "CODE_OF_CONDUCT.md"),
		filepath.Join(rootDir, "SECURITY.md"),
		filepath.Join(rootDir, "LICENSE"),
		filepath.Join(rootDir, "outputs.tf"),
		filepath.Join(rootDir, "variables.tf"),
		filepath.Join(rootDir, "terraform.tf"),
		filepath.Join(rootDir, "Makefile"),
		filepath.Join(rootDir, "TESTING.md"),
	}
	return &FileValidator{
		files: files,
	}
}

// Validate checks if required files exist and are not empty
func (fv *FileValidator) Validate() []error {
	var allErrors []error
	for _, filePath := range fv.files {
		allErrors = append(allErrors, validateFile(filePath)...)
	}
	return allErrors
}

// validateFile checks if a file exists and is not empty
func validateFile(filePath string) []error {
	var errors []error
	fileInfo, err := os.Stat(filePath)
	baseName := filepath.Base(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			errors = append(errors, formatError("file does not exist:\n  %s", baseName))
		} else {
			errors = append(errors, formatError("error accessing file:\n  %s\n  %v", baseName, err))
		}
		return errors
	}

	if fileInfo.Size() == 0 {
		errors = append(errors, formatError("file is empty:\n  %s", baseName))
	}

	return errors
}
