package markparsr

import (
	"fmt"
	"os"
	"path/filepath"
)

type FileValidator struct {
	rootDir         string
	requiredFiles   []string
	additionalFiles []string
}

func NewFileValidator(readmePath string, modulePath string, additionalFiles []string) *FileValidator {
	requiredFiles := []string{
		readmePath,
		filepath.Join(modulePath, "outputs.tf"),
		filepath.Join(modulePath, "variables.tf"),
		filepath.Join(modulePath, "terraform.tf"),
	}

	var absAdditionalFiles []string
	for _, file := range additionalFiles {
		if !filepath.IsAbs(file) {
			absAdditionalFiles = append(absAdditionalFiles, filepath.Join(modulePath, file))
		} else {
			absAdditionalFiles = append(absAdditionalFiles, file)
		}
	}

	return &FileValidator{
		rootDir:         modulePath,
		requiredFiles:   requiredFiles,
		additionalFiles: absAdditionalFiles,
	}
}

func (fv *FileValidator) Validate() []error {
	var allErrors []error

	for _, filePath := range fv.requiredFiles {
		if err := validateFile(filePath); err != nil {
			allErrors = append(allErrors, fmt.Errorf("required %v", err))
		}
	}

	for _, filePath := range fv.additionalFiles {
		if err := validateFile(filePath); err != nil {
			allErrors = append(allErrors, fmt.Errorf("additional %v", err))
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
