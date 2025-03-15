package markparsr

import (
	"fmt"
	"os"
	"path/filepath"
)

// Validator defines the interface for all validation components.
type Validator interface {
	Validate() []error
}

// ReadmeValidator coordinates validation of Terraform module documentation.
type ReadmeValidator struct {
	readmePath string
	modulePath string
	markdown   *MarkdownContent
	terraform  *TerraformContent
	validators []Validator
}

// NewReadmeValidator creates a validator for the specified README file.
// The optional modulePath parameter specifies where to find Terraform files.
// Without modulePath, the README's directory is used.
// Environment variables README_PATH and MODULE_PATH take precedence when set.
func NewReadmeValidator(readmePath string, modulePath ...string) (*ReadmeValidator, error) {
	// Check for README_PATH override
	if envPath := os.Getenv("README_PATH"); envPath != "" {
		readmePath = envPath
	}

	absReadmePath, err := filepath.Abs(readmePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for README: %w", err)
	}

	// Determine the module path
	var moduleDir string

	// Check for MODULE_PATH override first
	if envModulePath := os.Getenv("MODULE_PATH"); envModulePath != "" {
		moduleDir = envModulePath
	} else if len(modulePath) > 0 && modulePath[0] != "" {
		// Use explicitly provided module path if available
		moduleDir = modulePath[0]
	} else {
		// Default to README's directory
		moduleDir = filepath.Dir(absReadmePath)
	}

	absModulePath, err := filepath.Abs(moduleDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute module path: %w", err)
	}

	data, err := os.ReadFile(absReadmePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	markdown := NewMarkdownContent(string(data))

	terraform, err := NewTerraformContent(absModulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize terraform content: %w", err)
	}

	validator := &ReadmeValidator{
		readmePath: absReadmePath,
		modulePath: absModulePath,
		markdown:   markdown,
		terraform:  terraform,
	}

	sectionValidator := NewSectionValidator(markdown)
	validator.validators = []Validator{
		sectionValidator,
		NewFileValidator(absReadmePath, absModulePath),
		NewURLValidator(markdown),
		NewTerraformDefinitionValidator(markdown, terraform),
		NewItemValidator(markdown, terraform, "Variables", "variable", []string{"Required Inputs", "Optional Inputs"}, "variables.tf"),
		NewItemValidator(markdown, terraform, "Outputs", "output", []string{"Outputs"}, "outputs.tf"),
	}

	return validator, nil
}

// Validate runs all validators and collects their errors.
func (rv *ReadmeValidator) Validate() []error {
	var allErrors []error
	for _, validator := range rv.validators {
		allErrors = append(allErrors, validator.Validate()...)
	}
	return allErrors
}
