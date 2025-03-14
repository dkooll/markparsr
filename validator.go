package markparsr

import (
	"fmt"
	"os"
	"path/filepath"
)

type Validator interface {
	Validate() []error
}

// ReadmeValidator is the main validator that coordinates validation
// of Terraform module documentation.
type ReadmeValidator struct {
	readmePath string
	markdown   *MarkdownContent
	terraform  *TerraformContent
	validators []Validator
}

// NewReadmeValidator creates a new ReadmeValidator for the specified README file.
// It initializes all required validators to check various aspects of the documentation.
// The readmePath can be overridden by setting the README_PATH environment variable.
// Returns:
//   - A pointer to the initialized ReadmeValidator
//   - An error if initialization fails (file not found, parsing errors, etc.)
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

// Validate runs all registered validators and collects their errors.
// Each validator is executed independently, and errors from all validators
// are combined into a single slice.
// Returns:
//   - A slice of errors from all validators. Empty if validation is successful.
func (rv *ReadmeValidator) Validate() []error {
	var allErrors []error
	for _, validator := range rv.validators {
		allErrors = append(allErrors, validator.Validate()...)
	}
	return allErrors
}
