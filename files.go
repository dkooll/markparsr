package markparsr

import (
	"fmt"
	"os"
	"path/filepath"
)

type FileValidator struct {
	rootDir string
	files   []string
}

func NewFileValidator(readmePath string) *FileValidator {
	rootDir := filepath.Dir(readmePath)
	files := []string{
		readmePath,
		filepath.Join(rootDir, "outputs.tf"),
		filepath.Join(rootDir, "variables.tf"),
		filepath.Join(rootDir, "terraform.tf"),
		filepath.Join(rootDir, "Makefile"),
	}
	return &FileValidator{rootDir: rootDir, files: files}
}

func (fv *FileValidator) Validate() []error {
	var allErrors []error
	for _, filePath := range fv.files {
		if err := validateFile(filePath); err != nil {
			allErrors = append(allErrors, err)
		}
	}
	return allErrors
}

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
