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
// It runs multiple specialized validators to ensure documentation completeness.
type ReadmeValidator struct {
	readmePath string
	markdown   *MarkdownContent
	terraform  *TerraformContent
	validators []Validator
}

// NewReadmeValidator creates a validator for the specified README file.
// The path can be overridden by setting the README_PATH environment variable.
func NewReadmeValidator(readmePath string) (*ReadmeValidator, error) {
	if envPath := os.Getenv("README_PATH"); envPath != "" {
		readmePath = envPath
	}

	absReadmePath, err := filepath.Abs(readmePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	data, err := os.ReadFile(absReadmePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	markdown := NewMarkdownContent(string(data))

	terraform, err := NewTerraformContent()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize terraform content: %w", err)
	}

	validator := &ReadmeValidator{
		readmePath: absReadmePath,
		markdown:   markdown,
		terraform:  terraform,
	}

	sectionValidator := NewSectionValidator(markdown)
	validator.validators = []Validator{
		sectionValidator,
		NewFileValidator(absReadmePath),
		NewURLValidator(markdown),
		NewTerraformDefinitionValidator(markdown, terraform),
		NewItemValidator(markdown, terraform, "Variables", "variable", []string{"Required Inputs", "Optional Inputs"}, "variables.tf"),
		NewItemValidator(markdown, terraform, "Outputs", "output", []string{"Outputs"}, "outputs.tf"),
	}

	return validator, nil
}

// Validate runs all validators and collects their errors.
// Returns an empty slice if validation is successful.
func (rv *ReadmeValidator) Validate() []error {
	var allErrors []error
	for _, validator := range rv.validators {
		allErrors = append(allErrors, validator.Validate()...)
	}
	return allErrors
}
