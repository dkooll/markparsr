package markparsr

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileValidator checks for the existence and non-emptiness of required files
// for a Terraform module.
type FileValidator struct {
	rootDir string
	files   []string
}

// NewFileValidator creates a new FileValidator that checks for required files
// in the directory containing the README.md file.
// Required files include README.md, outputs.tf, variables.tf, terraform.tf, and Makefile.
// Parameters:
//   - readmePath: Path to the README.md file
//
// Returns:
//   - A pointer to the initialized FileValidator
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

// Validate checks that all required files exist and are not empty.
// For each file in the validator's list, it checks existence and size,
// adding an error to the returned slice for each missing or empty file.
// Returns:
//   - A slice of errors for missing or empty files. Empty if all files are valid.
func (fv *FileValidator) Validate() []error {
	var allErrors []error
	for _, filePath := range fv.files {
		if err := validateFile(filePath); err != nil {
			allErrors = append(allErrors, err)
		}
	}
	return allErrors
}

// validateFile checks if a specific file exists and is not empty.
// Parameters:
//   - filePath: Path to the file to validate
//
// Returns:
//   - nil if the file exists and is not empty
//   - An error if the file is missing, cannot be accessed, or is empty
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
